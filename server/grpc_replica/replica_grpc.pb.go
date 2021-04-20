// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpcreplica

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ReplicaClient is the client API for Replica service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReplicaClient interface {
	GetLastKey(ctx context.Context, in *GetLastKeyRequest, opts ...grpc.CallOption) (*GetLastKeyReply, error)
	GetTrained(ctx context.Context, in *GetTrainedRequest, opts ...grpc.CallOption) (*GetTrainedReply, error)
	GetData(ctx context.Context, in *GetDataRequest, opts ...grpc.CallOption) (*GetDataReply, error)
	GetCurrentOplog(ctx context.Context, in *GetCurrentOplogRequest, opts ...grpc.CallOption) (*GetCurrentOplogReply, error)
}

type replicaClient struct {
	cc grpc.ClientConnInterface
}

func NewReplicaClient(cc grpc.ClientConnInterface) ReplicaClient {
	return &replicaClient{cc}
}

func (c *replicaClient) GetLastKey(ctx context.Context, in *GetLastKeyRequest, opts ...grpc.CallOption) (*GetLastKeyReply, error) {
	out := new(GetLastKeyReply)
	err := c.cc.Invoke(ctx, "/replica.Replica/GetLastKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicaClient) GetTrained(ctx context.Context, in *GetTrainedRequest, opts ...grpc.CallOption) (*GetTrainedReply, error) {
	out := new(GetTrainedReply)
	err := c.cc.Invoke(ctx, "/replica.Replica/GetTrained", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicaClient) GetData(ctx context.Context, in *GetDataRequest, opts ...grpc.CallOption) (*GetDataReply, error) {
	out := new(GetDataReply)
	err := c.cc.Invoke(ctx, "/replica.Replica/GetData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicaClient) GetCurrentOplog(ctx context.Context, in *GetCurrentOplogRequest, opts ...grpc.CallOption) (*GetCurrentOplogReply, error) {
	out := new(GetCurrentOplogReply)
	err := c.cc.Invoke(ctx, "/replica.Replica/GetCurrentOplog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicaServer is the server API for Replica service.
// All implementations must embed UnimplementedReplicaServer
// for forward compatibility
type ReplicaServer interface {
	GetLastKey(context.Context, *GetLastKeyRequest) (*GetLastKeyReply, error)
	GetTrained(context.Context, *GetTrainedRequest) (*GetTrainedReply, error)
	GetData(context.Context, *GetDataRequest) (*GetDataReply, error)
	GetCurrentOplog(context.Context, *GetCurrentOplogRequest) (*GetCurrentOplogReply, error)
	mustEmbedUnimplementedReplicaServer()
}

// UnimplementedReplicaServer must be embedded to have forward compatible implementations.
type UnimplementedReplicaServer struct {
}

func (UnimplementedReplicaServer) GetLastKey(context.Context, *GetLastKeyRequest) (*GetLastKeyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLastKey not implemented")
}
func (UnimplementedReplicaServer) GetTrained(context.Context, *GetTrainedRequest) (*GetTrainedReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTrained not implemented")
}
func (UnimplementedReplicaServer) GetData(context.Context, *GetDataRequest) (*GetDataReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetData not implemented")
}
func (UnimplementedReplicaServer) GetCurrentOplog(context.Context, *GetCurrentOplogRequest) (*GetCurrentOplogReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCurrentOplog not implemented")
}
func (UnimplementedReplicaServer) mustEmbedUnimplementedReplicaServer() {}

// UnsafeReplicaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReplicaServer will
// result in compilation errors.
type UnsafeReplicaServer interface {
	mustEmbedUnimplementedReplicaServer()
}

func RegisterReplicaServer(s grpc.ServiceRegistrar, srv ReplicaServer) {
	s.RegisterService(&Replica_ServiceDesc, srv)
}

func _Replica_GetLastKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLastKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaServer).GetLastKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/replica.Replica/GetLastKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaServer).GetLastKey(ctx, req.(*GetLastKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Replica_GetTrained_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTrainedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaServer).GetTrained(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/replica.Replica/GetTrained",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaServer).GetTrained(ctx, req.(*GetTrainedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Replica_GetData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaServer).GetData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/replica.Replica/GetData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaServer).GetData(ctx, req.(*GetDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Replica_GetCurrentOplog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCurrentOplogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaServer).GetCurrentOplog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/replica.Replica/GetCurrentOplog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaServer).GetCurrentOplog(ctx, req.(*GetCurrentOplogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Replica_ServiceDesc is the grpc.ServiceDesc for Replica service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Replica_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "replica.Replica",
	HandlerType: (*ReplicaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetLastKey",
			Handler:    _Replica_GetLastKey_Handler,
		},
		{
			MethodName: "GetTrained",
			Handler:    _Replica_GetTrained_Handler,
		},
		{
			MethodName: "GetData",
			Handler:    _Replica_GetData_Handler,
		},
		{
			MethodName: "GetCurrentOplog",
			Handler:    _Replica_GetCurrentOplog_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "replica.proto",
}