// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: authorization/authorization.proto

package authorization

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
	AuthorizationService_GetDecisions_FullMethodName    = "/authorization.AuthorizationService/GetDecisions"
	AuthorizationService_GetEntitlements_FullMethodName = "/authorization.AuthorizationService/GetEntitlements"
)

// AuthorizationServiceClient is the client API for AuthorizationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthorizationServiceClient interface {
	GetDecisions(ctx context.Context, in *GetDecisionsRequest, opts ...grpc.CallOption) (*GetDecisionsResponse, error)
	GetEntitlements(ctx context.Context, in *GetEntitlementsRequest, opts ...grpc.CallOption) (*GetEntitlementsResponse, error)
}

type authorizationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthorizationServiceClient(cc grpc.ClientConnInterface) AuthorizationServiceClient {
	return &authorizationServiceClient{cc}
}

func (c *authorizationServiceClient) GetDecisions(ctx context.Context, in *GetDecisionsRequest, opts ...grpc.CallOption) (*GetDecisionsResponse, error) {
	out := new(GetDecisionsResponse)
	err := c.cc.Invoke(ctx, AuthorizationService_GetDecisions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authorizationServiceClient) GetEntitlements(ctx context.Context, in *GetEntitlementsRequest, opts ...grpc.CallOption) (*GetEntitlementsResponse, error) {
	out := new(GetEntitlementsResponse)
	err := c.cc.Invoke(ctx, AuthorizationService_GetEntitlements_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthorizationServiceServer is the server API for AuthorizationService service.
// All implementations must embed UnimplementedAuthorizationServiceServer
// for forward compatibility
type AuthorizationServiceServer interface {
	GetDecisions(context.Context, *GetDecisionsRequest) (*GetDecisionsResponse, error)
	GetEntitlements(context.Context, *GetEntitlementsRequest) (*GetEntitlementsResponse, error)
	mustEmbedUnimplementedAuthorizationServiceServer()
}

// UnimplementedAuthorizationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuthorizationServiceServer struct {
}

func (UnimplementedAuthorizationServiceServer) GetDecisions(context.Context, *GetDecisionsRequest) (*GetDecisionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDecisions not implemented")
}
func (UnimplementedAuthorizationServiceServer) GetEntitlements(context.Context, *GetEntitlementsRequest) (*GetEntitlementsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEntitlements not implemented")
}
func (UnimplementedAuthorizationServiceServer) mustEmbedUnimplementedAuthorizationServiceServer() {}

// UnsafeAuthorizationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthorizationServiceServer will
// result in compilation errors.
type UnsafeAuthorizationServiceServer interface {
	mustEmbedUnimplementedAuthorizationServiceServer()
}

func RegisterAuthorizationServiceServer(s grpc.ServiceRegistrar, srv AuthorizationServiceServer) {
	s.RegisterService(&AuthorizationService_ServiceDesc, srv)
}

func _AuthorizationService_GetDecisions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDecisionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthorizationServiceServer).GetDecisions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthorizationService_GetDecisions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthorizationServiceServer).GetDecisions(ctx, req.(*GetDecisionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthorizationService_GetEntitlements_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetEntitlementsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthorizationServiceServer).GetEntitlements(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthorizationService_GetEntitlements_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthorizationServiceServer).GetEntitlements(ctx, req.(*GetEntitlementsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthorizationService_ServiceDesc is the grpc.ServiceDesc for AuthorizationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthorizationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "authorization.AuthorizationService",
	HandlerType: (*AuthorizationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDecisions",
			Handler:    _AuthorizationService_GetDecisions_Handler,
		},
		{
			MethodName: "GetEntitlements",
			Handler:    _AuthorizationService_GetEntitlements_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "authorization/authorization.proto",
}
