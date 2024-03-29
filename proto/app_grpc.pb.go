// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.4
// source: proto/app.proto

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
	URLShortener_MakeURL_FullMethodName        = "/shortener.URLShortener/MakeURL"
	URLShortener_GetOriginalURL_FullMethodName = "/shortener.URLShortener/GetOriginalURL"
	URLShortener_GetUserURLs_FullMethodName    = "/shortener.URLShortener/GetUserURLs"
	URLShortener_DeleteURLs_FullMethodName     = "/shortener.URLShortener/DeleteURLs"
	URLShortener_PingDB_FullMethodName         = "/shortener.URLShortener/PingDB"
	URLShortener_GetStats_FullMethodName       = "/shortener.URLShortener/GetStats"
)

// URLShortenerClient is the client API for URLShortener service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type URLShortenerClient interface {
	MakeURL(ctx context.Context, in *MakeURLRequest, opts ...grpc.CallOption) (*MakeURLResponse, error)
	GetOriginalURL(ctx context.Context, in *GetURLRequest, opts ...grpc.CallOption) (*GetURLResponse, error)
	GetUserURLs(ctx context.Context, in *UserURLsRequest, opts ...grpc.CallOption) (*UserURLsResponse, error)
	DeleteURLs(ctx context.Context, in *DeleteURLsRequest, opts ...grpc.CallOption) (*DeleteURLsResponse, error)
	PingDB(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	GetStats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error)
}

type uRLShortenerClient struct {
	cc grpc.ClientConnInterface
}

func NewURLShortenerClient(cc grpc.ClientConnInterface) URLShortenerClient {
	return &uRLShortenerClient{cc}
}

func (c *uRLShortenerClient) MakeURL(ctx context.Context, in *MakeURLRequest, opts ...grpc.CallOption) (*MakeURLResponse, error) {
	out := new(MakeURLResponse)
	err := c.cc.Invoke(ctx, URLShortener_MakeURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) GetOriginalURL(ctx context.Context, in *GetURLRequest, opts ...grpc.CallOption) (*GetURLResponse, error) {
	out := new(GetURLResponse)
	err := c.cc.Invoke(ctx, URLShortener_GetOriginalURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) GetUserURLs(ctx context.Context, in *UserURLsRequest, opts ...grpc.CallOption) (*UserURLsResponse, error) {
	out := new(UserURLsResponse)
	err := c.cc.Invoke(ctx, URLShortener_GetUserURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) DeleteURLs(ctx context.Context, in *DeleteURLsRequest, opts ...grpc.CallOption) (*DeleteURLsResponse, error) {
	out := new(DeleteURLsResponse)
	err := c.cc.Invoke(ctx, URLShortener_DeleteURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) PingDB(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, URLShortener_PingDB_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) GetStats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error) {
	out := new(StatsResponse)
	err := c.cc.Invoke(ctx, URLShortener_GetStats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// URLShortenerServer is the server API for URLShortener service.
// All implementations must embed UnimplementedURLShortenerServer
// for forward compatibility
type URLShortenerServer interface {
	MakeURL(context.Context, *MakeURLRequest) (*MakeURLResponse, error)
	GetOriginalURL(context.Context, *GetURLRequest) (*GetURLResponse, error)
	GetUserURLs(context.Context, *UserURLsRequest) (*UserURLsResponse, error)
	DeleteURLs(context.Context, *DeleteURLsRequest) (*DeleteURLsResponse, error)
	PingDB(context.Context, *PingRequest) (*PingResponse, error)
	GetStats(context.Context, *StatsRequest) (*StatsResponse, error)
	mustEmbedUnimplementedURLShortenerServer()
}

// UnimplementedURLShortenerServer must be embedded to have forward compatible implementations.
type UnimplementedURLShortenerServer struct {
}

func (UnimplementedURLShortenerServer) MakeURL(context.Context, *MakeURLRequest) (*MakeURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MakeURL not implemented")
}
func (UnimplementedURLShortenerServer) GetOriginalURL(context.Context, *GetURLRequest) (*GetURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOriginalURL not implemented")
}
func (UnimplementedURLShortenerServer) GetUserURLs(context.Context, *UserURLsRequest) (*UserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserURLs not implemented")
}
func (UnimplementedURLShortenerServer) DeleteURLs(context.Context, *DeleteURLsRequest) (*DeleteURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteURLs not implemented")
}
func (UnimplementedURLShortenerServer) PingDB(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingDB not implemented")
}
func (UnimplementedURLShortenerServer) GetStats(context.Context, *StatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
func (UnimplementedURLShortenerServer) mustEmbedUnimplementedURLShortenerServer() {}

// UnsafeURLShortenerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to URLShortenerServer will
// result in compilation errors.
type UnsafeURLShortenerServer interface {
	mustEmbedUnimplementedURLShortenerServer()
}

func RegisterURLShortenerServer(s grpc.ServiceRegistrar, srv URLShortenerServer) {
	s.RegisterService(&URLShortener_ServiceDesc, srv)
}

func _URLShortener_MakeURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MakeURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).MakeURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_MakeURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).MakeURL(ctx, req.(*MakeURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_GetOriginalURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).GetOriginalURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_GetOriginalURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).GetOriginalURL(ctx, req.(*GetURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_GetUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).GetUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_GetUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).GetUserURLs(ctx, req.(*UserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_DeleteURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).DeleteURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_DeleteURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).DeleteURLs(ctx, req.(*DeleteURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_PingDB_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).PingDB(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_PingDB_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).PingDB(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_GetStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).GetStats(ctx, req.(*StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// URLShortener_ServiceDesc is the grpc.ServiceDesc for URLShortener service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var URLShortener_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "shortener.URLShortener",
	HandlerType: (*URLShortenerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MakeURL",
			Handler:    _URLShortener_MakeURL_Handler,
		},
		{
			MethodName: "GetOriginalURL",
			Handler:    _URLShortener_GetOriginalURL_Handler,
		},
		{
			MethodName: "GetUserURLs",
			Handler:    _URLShortener_GetUserURLs_Handler,
		},
		{
			MethodName: "DeleteURLs",
			Handler:    _URLShortener_DeleteURLs_Handler,
		},
		{
			MethodName: "PingDB",
			Handler:    _URLShortener_PingDB_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    _URLShortener_GetStats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/app.proto",
}
