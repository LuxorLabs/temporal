// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package worker

import (
	"sync/atomic"
	"time"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	persistenceClient "github.com/uber/cadence/common/persistence/client"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/service/worker/archiver"
	"github.com/uber/cadence/service/worker/batcher"
	"github.com/uber/cadence/service/worker/indexer"
	"github.com/uber/cadence/service/worker/parentclosepolicy"
	"github.com/uber/cadence/service/worker/replicator"
	"github.com/uber/cadence/service/worker/scanner"
	"github.com/uber/cadence/service/worker/scanner/executions"
)

type (
	// Service represents the cadence-worker service. This service hosts all background processing needed for cadence cluster:
	// 1. Replicator: Handles applying replication tasks generated by remote clusters.
	// 2. Indexer: Handles uploading of visibility records to elastic search.
	// 3. Archiver: Handles archival of workflow histories.
	Service struct {
		resource.Resource

		status int32
		stopC  chan struct{}
		params *service.BootstrapParams
		config *Config
	}

	// Config contains all the service config for worker
	Config struct {
		ReplicationCfg                *replicator.Config
		ArchiverConfig                *archiver.Config
		IndexerCfg                    *indexer.Config
		ScannerCfg                    *scanner.Config
		BatcherCfg                    *batcher.Config
		ThrottledLogRPS               dynamicconfig.IntPropertyFn
		PersistenceGlobalMaxQPS       dynamicconfig.IntPropertyFn
		EnableBatcher                 dynamicconfig.BoolPropertyFn
		EnableParentClosePolicyWorker dynamicconfig.BoolPropertyFn
	}
)

// NewService builds a new cadence-worker service
func NewService(
	params *service.BootstrapParams,
) (resource.Resource, error) {

	serviceConfig := NewConfig(params)

	serviceResource, err := resource.New(
		params,
		common.WorkerServiceName,
		serviceConfig.ReplicationCfg.PersistenceMaxQPS,
		serviceConfig.PersistenceGlobalMaxQPS,
		serviceConfig.ThrottledLogRPS,
		func(
			persistenceBean persistenceClient.Bean,
			logger log.Logger,
		) (persistence.VisibilityManager, error) {
			return persistenceBean.GetVisibilityManager(), nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		Resource: serviceResource,
		status:   common.DaemonStatusInitialized,
		config:   serviceConfig,
		params:   params,
		stopC:    make(chan struct{}),
	}, nil
}

// NewConfig builds the new Config for cadence-worker service
func NewConfig(params *service.BootstrapParams) *Config {
	dc := dynamicconfig.NewCollection(params.DynamicConfig, params.Logger)
	config := &Config{
		ReplicationCfg: &replicator.Config{
			PersistenceMaxQPS:                  dc.GetIntProperty(dynamicconfig.WorkerPersistenceMaxQPS, 500),
			ReplicatorMetaTaskConcurrency:      dc.GetIntProperty(dynamicconfig.WorkerReplicatorMetaTaskConcurrency, 64),
			ReplicatorTaskConcurrency:          dc.GetIntProperty(dynamicconfig.WorkerReplicatorTaskConcurrency, 256),
			ReplicatorMessageConcurrency:       dc.GetIntProperty(dynamicconfig.WorkerReplicatorMessageConcurrency, 2048),
			ReplicatorActivityBufferRetryCount: dc.GetIntProperty(dynamicconfig.WorkerReplicatorActivityBufferRetryCount, 8),
			ReplicatorHistoryBufferRetryCount:  dc.GetIntProperty(dynamicconfig.WorkerReplicatorHistoryBufferRetryCount, 8),
			ReplicationTaskMaxRetryCount:       dc.GetIntProperty(dynamicconfig.WorkerReplicationTaskMaxRetryCount, 400),
			ReplicationTaskMaxRetryDuration:    dc.GetDurationProperty(dynamicconfig.WorkerReplicationTaskMaxRetryDuration, 15*time.Minute),
			ReplicationTaskContextTimeout:      dc.GetDurationProperty(dynamicconfig.WorkerReplicationTaskContextDuration, 30*time.Second),
			ReReplicationContextTimeout:        dc.GetDurationPropertyFilteredByDomainID(dynamicconfig.WorkerReReplicationContextTimeout, 0*time.Second),
			EnableRPCReplication:               dc.GetBoolProperty(dynamicconfig.WorkerEnableRPCReplication, false),
		},
		ArchiverConfig: &archiver.Config{
			ArchiverConcurrency:           dc.GetIntProperty(dynamicconfig.WorkerArchiverConcurrency, 50),
			ArchivalsPerIteration:         dc.GetIntProperty(dynamicconfig.WorkerArchivalsPerIteration, 1000),
			TimeLimitPerArchivalIteration: dc.GetDurationProperty(dynamicconfig.WorkerTimeLimitPerArchivalIteration, archiver.MaxArchivalIterationTimeout()),
		},
		ScannerCfg: &scanner.Config{
			PersistenceMaxQPS:        dc.GetIntProperty(dynamicconfig.ScannerPersistenceMaxQPS, 100),
			Persistence:              &params.PersistenceConfig,
			ClusterMetadata:          params.ClusterMetadata,
			TaskListScannerEnabled:   dc.GetBoolProperty(dynamicconfig.TaskListScannerEnabled, true),
			HistoryScannerEnabled:    dc.GetBoolProperty(dynamicconfig.HistoryScannerEnabled, true),
			ExecutionsScannerEnabled: dc.GetBoolProperty(dynamicconfig.ExecutionsScannerEnabled, false),
			ExecutionScannerConfig: &executions.ScannerWorkflowDynamicConfig{
				Enabled:                 dc.GetBoolProperty(dynamicconfig.ExecutionsScannerEnabled, false),
				Concurrency:             dc.GetIntProperty(dynamicconfig.ExecutionsScannerConcurrency, 25),
				ExecutionsPageSize:      dc.GetIntProperty(dynamicconfig.ExecutionsScannerPersistencePageSize, 1000),
				BlobstoreFlushThreshold: dc.GetIntProperty(dynamicconfig.ExecutionsScannerBlobstoreFlushThreshold, 100),
				DynamicConfigInvariantCollections: executions.DynamicConfigInvariantCollections{
					InvariantCollectionMutableState: dc.GetBoolProperty(dynamicconfig.ExecutionsScannerInvariantCollectionMutableState, false),
					InvariantCollectionHistory:      dc.GetBoolProperty(dynamicconfig.ExecutionsScannerInvariantCollectionHistory, false),
				},
			},
		},
		BatcherCfg: &batcher.Config{
			AdminOperationToken: dc.GetStringProperty(dynamicconfig.AdminOperationToken, common.DefaultAdminOperationToken),
			ClusterMetadata:     params.ClusterMetadata,
		},
		EnableBatcher:                 dc.GetBoolProperty(dynamicconfig.EnableBatcher, false),
		EnableParentClosePolicyWorker: dc.GetBoolProperty(dynamicconfig.EnableParentClosePolicyWorker, true),
		ThrottledLogRPS:               dc.GetIntProperty(dynamicconfig.WorkerThrottledLogRPS, 20),
		PersistenceGlobalMaxQPS:       dc.GetIntProperty(dynamicconfig.WorkerPersistenceGlobalMaxQPS, 0),
	}
	advancedVisWritingMode := dc.GetStringProperty(
		dynamicconfig.AdvancedVisibilityWritingMode,
		common.GetDefaultAdvancedVisibilityWritingMode(params.PersistenceConfig.IsAdvancedVisibilityConfigExist()),
	)
	if advancedVisWritingMode() != common.AdvancedVisibilityWritingModeOff {
		config.IndexerCfg = &indexer.Config{
			IndexerConcurrency:       dc.GetIntProperty(dynamicconfig.WorkerIndexerConcurrency, 1000),
			ESProcessorNumOfWorkers:  dc.GetIntProperty(dynamicconfig.WorkerESProcessorNumOfWorkers, 1),
			ESProcessorBulkActions:   dc.GetIntProperty(dynamicconfig.WorkerESProcessorBulkActions, 1000),
			ESProcessorBulkSize:      dc.GetIntProperty(dynamicconfig.WorkerESProcessorBulkSize, 2<<24), // 16MB
			ESProcessorFlushInterval: dc.GetDurationProperty(dynamicconfig.WorkerESProcessorFlushInterval, 1*time.Second),
			ValidSearchAttributes:    dc.GetMapProperty(dynamicconfig.ValidSearchAttributes, definition.GetDefaultIndexedKeys()),
		}
	}
	return config
}

// Start is called to start the service
func (s *Service) Start() {
	if !atomic.CompareAndSwapInt32(&s.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	logger := s.GetLogger()
	logger.Info("worker starting", tag.ComponentWorker)

	s.Resource.Start()

	s.ensureSystemDomainExists()
	s.startScanner()
	if s.config.IndexerCfg != nil {
		s.startIndexer()
	}

	if s.GetClusterMetadata().IsGlobalDomainEnabled() {
		s.startReplicator()
	}
	if s.GetArchivalMetadata().GetHistoryConfig().ClusterConfiguredForArchival() {
		s.startArchiver()
	}
	if s.config.EnableBatcher() {
		s.startBatcher()
	}
	if s.config.EnableParentClosePolicyWorker() {
		s.startParentClosePolicyProcessor()
	}

	logger.Info("worker started", tag.ComponentWorker)
	<-s.stopC
}

// Stop is called to stop the service
func (s *Service) Stop() {
	if !atomic.CompareAndSwapInt32(&s.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(s.stopC)

	s.Resource.Stop()

	s.params.Logger.Info("worker stopped", tag.ComponentWorker)
}

func (s *Service) startParentClosePolicyProcessor() {
	params := &parentclosepolicy.BootstrapParams{
		ServiceClient: s.params.PublicClient,
		MetricsClient: s.GetMetricsClient(),
		Logger:        s.GetLogger(),
		TallyScope:    s.params.MetricScope,
		ClientBean:    s.GetClientBean(),
	}
	processor := parentclosepolicy.New(params)
	if err := processor.Start(); err != nil {
		s.GetLogger().Fatal("error starting parentclosepolicy processor", tag.Error(err))
	}
}

func (s *Service) startBatcher() {
	params := &batcher.BootstrapParams{
		Config:        *s.config.BatcherCfg,
		ServiceClient: s.params.PublicClient,
		MetricsClient: s.GetMetricsClient(),
		Logger:        s.GetLogger(),
		TallyScope:    s.params.MetricScope,
		ClientBean:    s.GetClientBean(),
	}
	if err := batcher.New(params).Start(); err != nil {
		s.GetLogger().Fatal("error starting batcher", tag.Error(err))
	}
}

func (s *Service) startScanner() {
	params := &scanner.BootstrapParams{
		Config:     *s.config.ScannerCfg,
		TallyScope: s.params.MetricScope,
	}
	if err := scanner.New(s.Resource, params).Start(); err != nil {
		s.GetLogger().Fatal("error starting scanner", tag.Error(err))
	}
}

func (s *Service) startReplicator() {
	domainReplicationTaskExecutor := domain.NewReplicationTaskExecutor(
		s.GetMetadataManager(),
		s.GetLogger(),
	)
	msgReplicator := replicator.NewReplicator(
		s.GetClusterMetadata(),
		s.GetMetadataManager(),
		s.GetDomainCache(),
		s.GetClientBean(),
		s.config.ReplicationCfg,
		s.GetMessagingClient(),
		s.GetLogger(),
		s.GetMetricsClient(),
		s.GetHostInfo(),
		s.GetWorkerServiceResolver(),
		s.GetDomainReplicationQueue(),
		domainReplicationTaskExecutor,
	)
	if err := msgReplicator.Start(); err != nil {
		msgReplicator.Stop()
		s.GetLogger().Fatal("fail to start replicator", tag.Error(err))
	}
}

func (s *Service) startIndexer() {
	visibilityIndexer := indexer.NewIndexer(
		s.config.IndexerCfg,
		s.GetMessagingClient(),
		s.params.ESClient,
		s.params.ESConfig,
		s.GetLogger(),
		s.GetMetricsClient(),
	)
	if err := visibilityIndexer.Start(); err != nil {
		visibilityIndexer.Stop()
		s.GetLogger().Fatal("fail to start indexer", tag.Error(err))
	}
}

func (s *Service) startArchiver() {
	bc := &archiver.BootstrapContainer{
		PublicClient:     s.GetSDKClient(),
		MetricsClient:    s.GetMetricsClient(),
		Logger:           s.GetLogger(),
		HistoryV2Manager: s.GetHistoryManager(),
		DomainCache:      s.GetDomainCache(),
		Config:           s.config.ArchiverConfig,
		ArchiverProvider: s.GetArchiverProvider(),
	}
	clientWorker := archiver.NewClientWorker(bc)
	if err := clientWorker.Start(); err != nil {
		clientWorker.Stop()
		s.GetLogger().Fatal("failed to start archiver", tag.Error(err))
	}
}

func (s *Service) ensureSystemDomainExists() {
	_, err := s.GetMetadataManager().GetDomain(&persistence.GetDomainRequest{Name: common.SystemLocalDomainName})
	switch err.(type) {
	case nil:
		// noop
	case *shared.EntityNotExistsError:
		s.GetLogger().Info("cadence-system domain does not exist, attempting to register domain")
		s.registerSystemDomain()
	default:
		s.GetLogger().Fatal("failed to verify if cadence system domain exists", tag.Error(err))
	}
}

func (s *Service) registerSystemDomain() {

	currentClusterName := s.GetClusterMetadata().GetCurrentClusterName()
	_, err := s.GetMetadataManager().CreateDomain(&persistence.CreateDomainRequest{
		Info: &persistence.DomainInfo{
			ID:          common.SystemDomainID,
			Name:        common.SystemLocalDomainName,
			Description: "Cadence internal system domain",
		},
		Config: &persistence.DomainConfig{
			Retention:  common.SystemDomainRetentionDays,
			EmitMetric: true,
		},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: currentClusterName,
			Clusters:          persistence.GetOrUseDefaultClusters(currentClusterName, nil),
		},
		IsGlobalDomain:  false,
		FailoverVersion: common.EmptyVersion,
	})
	if err != nil {
		if _, ok := err.(*shared.DomainAlreadyExistsError); ok {
			return
		}
		s.GetLogger().Fatal("failed to register system domain", tag.Error(err))
	}
}
