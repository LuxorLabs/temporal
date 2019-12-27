// The MIT License (MIT)
// 
// Copyright (c) 2019 Uber Technologies, Inc.
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Code generated by thriftrw-plugin-yarpc
// @generated

package adminserviceclient

import (
	context "context"
	admin "github.com/temporalio/temporal/.gen/go/admin"
	shared "github.com/temporalio/temporal/.gen/go/shared"
	wire "go.uber.org/thriftrw/wire"
	yarpc "go.uber.org/yarpc"
	transport "go.uber.org/yarpc/api/transport"
	thrift "go.uber.org/yarpc/encoding/thrift"
	reflect "reflect"
)

// Interface is a client for the AdminService service.
type Interface interface {
	AddSearchAttribute(
		ctx context.Context,
		Request *admin.AddSearchAttributeRequest,
		opts ...yarpc.CallOption,
	) error

	CloseShard(
		ctx context.Context,
		Request *shared.CloseShardRequest,
		opts ...yarpc.CallOption,
	) error

	DescribeCluster(
		ctx context.Context,
		opts ...yarpc.CallOption,
	) (*admin.DescribeClusterResponse, error)

	DescribeHistoryHost(
		ctx context.Context,
		Request *shared.DescribeHistoryHostRequest,
		opts ...yarpc.CallOption,
	) (*shared.DescribeHistoryHostResponse, error)

	DescribeWorkflowExecution(
		ctx context.Context,
		Request *admin.DescribeWorkflowExecutionRequest,
		opts ...yarpc.CallOption,
	) (*admin.DescribeWorkflowExecutionResponse, error)

	GetWorkflowExecutionRawHistory(
		ctx context.Context,
		GetRequest *admin.GetWorkflowExecutionRawHistoryRequest,
		opts ...yarpc.CallOption,
	) (*admin.GetWorkflowExecutionRawHistoryResponse, error)

	GetWorkflowExecutionRawHistoryV2(
		ctx context.Context,
		GetRequest *admin.GetWorkflowExecutionRawHistoryV2Request,
		opts ...yarpc.CallOption,
	) (*admin.GetWorkflowExecutionRawHistoryV2Response, error)

	RemoveTask(
		ctx context.Context,
		Request *shared.RemoveTaskRequest,
		opts ...yarpc.CallOption,
	) error
}

// New builds a new client for the AdminService service.
//
// 	client := adminserviceclient.New(dispatcher.ClientConfig("adminservice"))
func New(c transport.ClientConfig, opts ...thrift.ClientOption) Interface {
	return client{
		c: thrift.New(thrift.Config{
			Service:      "AdminService",
			ClientConfig: c,
		}, opts...),
	}
}

func init() {
	yarpc.RegisterClientBuilder(
		func(c transport.ClientConfig, f reflect.StructField) Interface {
			return New(c, thrift.ClientBuilderOptions(c, f)...)
		},
	)
}

type client struct {
	c thrift.Client
}

func (c client) AddSearchAttribute(
	ctx context.Context,
	_Request *admin.AddSearchAttributeRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := admin.AdminService_AddSearchAttribute_Helper.Args(_Request)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_AddSearchAttribute_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	err = admin.AdminService_AddSearchAttribute_Helper.UnwrapResponse(&result)
	return
}

func (c client) CloseShard(
	ctx context.Context,
	_Request *shared.CloseShardRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := admin.AdminService_CloseShard_Helper.Args(_Request)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_CloseShard_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	err = admin.AdminService_CloseShard_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (success *admin.DescribeClusterResponse, err error) {

	args := admin.AdminService_DescribeCluster_Helper.Args()

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_DescribeCluster_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	success, err = admin.AdminService_DescribeCluster_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeHistoryHost(
	ctx context.Context,
	_Request *shared.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (success *shared.DescribeHistoryHostResponse, err error) {

	args := admin.AdminService_DescribeHistoryHost_Helper.Args(_Request)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_DescribeHistoryHost_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	success, err = admin.AdminService_DescribeHistoryHost_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeWorkflowExecution(
	ctx context.Context,
	_Request *admin.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (success *admin.DescribeWorkflowExecutionResponse, err error) {

	args := admin.AdminService_DescribeWorkflowExecution_Helper.Args(_Request)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_DescribeWorkflowExecution_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	success, err = admin.AdminService_DescribeWorkflowExecution_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetWorkflowExecutionRawHistory(
	ctx context.Context,
	_GetRequest *admin.GetWorkflowExecutionRawHistoryRequest,
	opts ...yarpc.CallOption,
) (success *admin.GetWorkflowExecutionRawHistoryResponse, err error) {

	args := admin.AdminService_GetWorkflowExecutionRawHistory_Helper.Args(_GetRequest)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_GetWorkflowExecutionRawHistory_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	success, err = admin.AdminService_GetWorkflowExecutionRawHistory_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	_GetRequest *admin.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (success *admin.GetWorkflowExecutionRawHistoryV2Response, err error) {

	args := admin.AdminService_GetWorkflowExecutionRawHistoryV2_Helper.Args(_GetRequest)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_GetWorkflowExecutionRawHistoryV2_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	success, err = admin.AdminService_GetWorkflowExecutionRawHistoryV2_Helper.UnwrapResponse(&result)
	return
}

func (c client) RemoveTask(
	ctx context.Context,
	_Request *shared.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := admin.AdminService_RemoveTask_Helper.Args(_Request)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_RemoveTask_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	err = admin.AdminService_RemoveTask_Helper.UnwrapResponse(&result)
	return
}
