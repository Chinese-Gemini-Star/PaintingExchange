// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.6.1
// source: internal/service/SearchService.proto

package service

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	SearchService_CreateImage_FullMethodName = "/service.medicons.SearchService/CreateImage"
	SearchService_UpdateImage_FullMethodName = "/service.medicons.SearchService/UpdateImage"
	SearchService_DeleteImage_FullMethodName = "/service.medicons.SearchService/DeleteImage"
	SearchService_SearchImage_FullMethodName = "/service.medicons.SearchService/SearchImage"
)

// SearchServiceClient is the client API for SearchService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SearchServiceClient interface {
	CreateImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*empty.Empty, error)
	UpdateImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*empty.Empty, error)
	DeleteImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*empty.Empty, error)
	SearchImage(ctx context.Context, in *Search, opts ...grpc.CallOption) (*Result, error)
}

type searchServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSearchServiceClient(cc grpc.ClientConnInterface) SearchServiceClient {
	return &searchServiceClient{cc}
}

func (c *searchServiceClient) CreateImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, SearchService_CreateImage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchServiceClient) UpdateImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, SearchService_UpdateImage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchServiceClient) DeleteImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, SearchService_DeleteImage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchServiceClient) SearchImage(ctx context.Context, in *Search, opts ...grpc.CallOption) (*Result, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Result)
	err := c.cc.Invoke(ctx, SearchService_SearchImage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SearchServiceServer is the server API for SearchService service.
// All implementations must embed UnimplementedSearchServiceServer
// for forward compatibility.
type SearchServiceServer interface {
	CreateImage(context.Context, *Image) (*empty.Empty, error)
	UpdateImage(context.Context, *Image) (*empty.Empty, error)
	DeleteImage(context.Context, *Image) (*empty.Empty, error)
	SearchImage(context.Context, *Search) (*Result, error)
	mustEmbedUnimplementedSearchServiceServer()
}

// UnimplementedSearchServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedSearchServiceServer struct{}

func (UnimplementedSearchServiceServer) CreateImage(context.Context, *Image) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateImage not implemented")
}
func (UnimplementedSearchServiceServer) UpdateImage(context.Context, *Image) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateImage not implemented")
}
func (UnimplementedSearchServiceServer) DeleteImage(context.Context, *Image) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteImage not implemented")
}
func (UnimplementedSearchServiceServer) SearchImage(context.Context, *Search) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchImage not implemented")
}
func (UnimplementedSearchServiceServer) mustEmbedUnimplementedSearchServiceServer() {}
func (UnimplementedSearchServiceServer) testEmbeddedByValue()                       {}

// UnsafeSearchServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SearchServiceServer will
// result in compilation errors.
type UnsafeSearchServiceServer interface {
	mustEmbedUnimplementedSearchServiceServer()
}

func RegisterSearchServiceServer(s grpc.ServiceRegistrar, srv SearchServiceServer) {
	// If the following call pancis, it indicates UnimplementedSearchServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&SearchService_ServiceDesc, srv)
}

func _SearchService_CreateImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Image)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchServiceServer).CreateImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SearchService_CreateImage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SearchServiceServer).CreateImage(ctx, req.(*Image))
	}
	return interceptor(ctx, in, info, handler)
}

func _SearchService_UpdateImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Image)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchServiceServer).UpdateImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SearchService_UpdateImage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SearchServiceServer).UpdateImage(ctx, req.(*Image))
	}
	return interceptor(ctx, in, info, handler)
}

func _SearchService_DeleteImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Image)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchServiceServer).DeleteImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SearchService_DeleteImage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SearchServiceServer).DeleteImage(ctx, req.(*Image))
	}
	return interceptor(ctx, in, info, handler)
}

func _SearchService_SearchImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Search)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchServiceServer).SearchImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SearchService_SearchImage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SearchServiceServer).SearchImage(ctx, req.(*Search))
	}
	return interceptor(ctx, in, info, handler)
}

// SearchService_ServiceDesc is the grpc.ServiceDesc for SearchService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SearchService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "service.medicons.SearchService",
	HandlerType: (*SearchServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateImage",
			Handler:    _SearchService_CreateImage_Handler,
		},
		{
			MethodName: "UpdateImage",
			Handler:    _SearchService_UpdateImage_Handler,
		},
		{
			MethodName: "DeleteImage",
			Handler:    _SearchService_DeleteImage_Handler,
		},
		{
			MethodName: "SearchImage",
			Handler:    _SearchService_SearchImage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/service/SearchService.proto",
}
