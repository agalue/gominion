// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package mdt_dialout

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

// GRPCMdtDialoutClient is the client API for GRPCMdtDialout service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GRPCMdtDialoutClient interface {
	MdtDialout(ctx context.Context, opts ...grpc.CallOption) (GRPCMdtDialout_MdtDialoutClient, error)
}

type gRPCMdtDialoutClient struct {
	cc grpc.ClientConnInterface
}

func NewGRPCMdtDialoutClient(cc grpc.ClientConnInterface) GRPCMdtDialoutClient {
	return &gRPCMdtDialoutClient{cc}
}

func (c *gRPCMdtDialoutClient) MdtDialout(ctx context.Context, opts ...grpc.CallOption) (GRPCMdtDialout_MdtDialoutClient, error) {
	stream, err := c.cc.NewStream(ctx, &GRPCMdtDialout_ServiceDesc.Streams[0], "/mdt_dialout.gRPCMdtDialout/MdtDialout", opts...)
	if err != nil {
		return nil, err
	}
	x := &gRPCMdtDialoutMdtDialoutClient{stream}
	return x, nil
}

type GRPCMdtDialout_MdtDialoutClient interface {
	Send(*MdtDialoutArgs) error
	Recv() (*MdtDialoutArgs, error)
	grpc.ClientStream
}

type gRPCMdtDialoutMdtDialoutClient struct {
	grpc.ClientStream
}

func (x *gRPCMdtDialoutMdtDialoutClient) Send(m *MdtDialoutArgs) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gRPCMdtDialoutMdtDialoutClient) Recv() (*MdtDialoutArgs, error) {
	m := new(MdtDialoutArgs)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GRPCMdtDialoutServer is the server API for GRPCMdtDialout service.
// All implementations must embed UnimplementedGRPCMdtDialoutServer
// for forward compatibility
type GRPCMdtDialoutServer interface {
	MdtDialout(GRPCMdtDialout_MdtDialoutServer) error
	mustEmbedUnimplementedGRPCMdtDialoutServer()
}

// UnimplementedGRPCMdtDialoutServer must be embedded to have forward compatible implementations.
type UnimplementedGRPCMdtDialoutServer struct {
}

func (UnimplementedGRPCMdtDialoutServer) MdtDialout(GRPCMdtDialout_MdtDialoutServer) error {
	return status.Errorf(codes.Unimplemented, "method MdtDialout not implemented")
}
func (UnimplementedGRPCMdtDialoutServer) mustEmbedUnimplementedGRPCMdtDialoutServer() {}

// UnsafeGRPCMdtDialoutServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GRPCMdtDialoutServer will
// result in compilation errors.
type UnsafeGRPCMdtDialoutServer interface {
	mustEmbedUnimplementedGRPCMdtDialoutServer()
}

func RegisterGRPCMdtDialoutServer(s grpc.ServiceRegistrar, srv GRPCMdtDialoutServer) {
	s.RegisterService(&GRPCMdtDialout_ServiceDesc, srv)
}

func _GRPCMdtDialout_MdtDialout_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GRPCMdtDialoutServer).MdtDialout(&gRPCMdtDialoutMdtDialoutServer{stream})
}

type GRPCMdtDialout_MdtDialoutServer interface {
	Send(*MdtDialoutArgs) error
	Recv() (*MdtDialoutArgs, error)
	grpc.ServerStream
}

type gRPCMdtDialoutMdtDialoutServer struct {
	grpc.ServerStream
}

func (x *gRPCMdtDialoutMdtDialoutServer) Send(m *MdtDialoutArgs) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gRPCMdtDialoutMdtDialoutServer) Recv() (*MdtDialoutArgs, error) {
	m := new(MdtDialoutArgs)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GRPCMdtDialout_ServiceDesc is the grpc.ServiceDesc for GRPCMdtDialout service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GRPCMdtDialout_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mdt_dialout.gRPCMdtDialout",
	HandlerType: (*GRPCMdtDialoutServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "MdtDialout",
			Handler:       _GRPCMdtDialout_MdtDialout_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "mdt_dialout.proto",
}
