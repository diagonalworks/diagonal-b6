// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.28.3
// source: api.proto

package proto

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

const (
	B6_Evaluate_FullMethodName    = "/api.B6/Evaluate"
	B6_DeleteWorld_FullMethodName = "/api.B6/DeleteWorld"
	B6_ListWorlds_FullMethodName  = "/api.B6/ListWorlds"
)

// B6Client is the client API for B6 service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type B6Client interface {
	Evaluate(ctx context.Context, in *EvaluateRequestProto, opts ...grpc.CallOption) (*EvaluateResponseProto, error)
	DeleteWorld(ctx context.Context, in *DeleteWorldRequestProto, opts ...grpc.CallOption) (*DeleteWorldResponseProto, error)
	ListWorlds(ctx context.Context, in *ListWorldsRequestProto, opts ...grpc.CallOption) (*ListWorldsResponseProto, error)
}

type b6Client struct {
	cc grpc.ClientConnInterface
}

func NewB6Client(cc grpc.ClientConnInterface) B6Client {
	return &b6Client{cc}
}

func (c *b6Client) Evaluate(ctx context.Context, in *EvaluateRequestProto, opts ...grpc.CallOption) (*EvaluateResponseProto, error) {
	out := new(EvaluateResponseProto)
	err := c.cc.Invoke(ctx, B6_Evaluate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *b6Client) DeleteWorld(ctx context.Context, in *DeleteWorldRequestProto, opts ...grpc.CallOption) (*DeleteWorldResponseProto, error) {
	out := new(DeleteWorldResponseProto)
	err := c.cc.Invoke(ctx, B6_DeleteWorld_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *b6Client) ListWorlds(ctx context.Context, in *ListWorldsRequestProto, opts ...grpc.CallOption) (*ListWorldsResponseProto, error) {
	out := new(ListWorldsResponseProto)
	err := c.cc.Invoke(ctx, B6_ListWorlds_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// B6Server is the server API for B6 service.
// All implementations must embed UnimplementedB6Server
// for forward compatibility
type B6Server interface {
	Evaluate(context.Context, *EvaluateRequestProto) (*EvaluateResponseProto, error)
	DeleteWorld(context.Context, *DeleteWorldRequestProto) (*DeleteWorldResponseProto, error)
	ListWorlds(context.Context, *ListWorldsRequestProto) (*ListWorldsResponseProto, error)
	mustEmbedUnimplementedB6Server()
}

// UnimplementedB6Server must be embedded to have forward compatible implementations.
type UnimplementedB6Server struct {
}

func (UnimplementedB6Server) Evaluate(context.Context, *EvaluateRequestProto) (*EvaluateResponseProto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Evaluate not implemented")
}
func (UnimplementedB6Server) DeleteWorld(context.Context, *DeleteWorldRequestProto) (*DeleteWorldResponseProto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteWorld not implemented")
}
func (UnimplementedB6Server) ListWorlds(context.Context, *ListWorldsRequestProto) (*ListWorldsResponseProto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWorlds not implemented")
}
func (UnimplementedB6Server) mustEmbedUnimplementedB6Server() {}

// UnsafeB6Server may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to B6Server will
// result in compilation errors.
type UnsafeB6Server interface {
	mustEmbedUnimplementedB6Server()
}

func RegisterB6Server(s grpc.ServiceRegistrar, srv B6Server) {
	s.RegisterService(&B6_ServiceDesc, srv)
}

func _B6_Evaluate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluateRequestProto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(B6Server).Evaluate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: B6_Evaluate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(B6Server).Evaluate(ctx, req.(*EvaluateRequestProto))
	}
	return interceptor(ctx, in, info, handler)
}

func _B6_DeleteWorld_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteWorldRequestProto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(B6Server).DeleteWorld(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: B6_DeleteWorld_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(B6Server).DeleteWorld(ctx, req.(*DeleteWorldRequestProto))
	}
	return interceptor(ctx, in, info, handler)
}

func _B6_ListWorlds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListWorldsRequestProto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(B6Server).ListWorlds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: B6_ListWorlds_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(B6Server).ListWorlds(ctx, req.(*ListWorldsRequestProto))
	}
	return interceptor(ctx, in, info, handler)
}

// B6_ServiceDesc is the grpc.ServiceDesc for B6 service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var B6_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.B6",
	HandlerType: (*B6Server)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Evaluate",
			Handler:    _B6_Evaluate_Handler,
		},
		{
			MethodName: "DeleteWorld",
			Handler:    _B6_DeleteWorld_Handler,
		},
		{
			MethodName: "ListWorlds",
			Handler:    _B6_ListWorlds_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
