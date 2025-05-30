// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: policy/unsafe/unsafe.proto

package unsafeconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	unsafe "github.com/opentdf/platform/protocol/go/policy/unsafe"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// UnsafeServiceName is the fully-qualified name of the UnsafeService service.
	UnsafeServiceName = "policy.unsafe.UnsafeService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// UnsafeServiceUnsafeUpdateNamespaceProcedure is the fully-qualified name of the UnsafeService's
	// UnsafeUpdateNamespace RPC.
	UnsafeServiceUnsafeUpdateNamespaceProcedure = "/policy.unsafe.UnsafeService/UnsafeUpdateNamespace"
	// UnsafeServiceUnsafeReactivateNamespaceProcedure is the fully-qualified name of the
	// UnsafeService's UnsafeReactivateNamespace RPC.
	UnsafeServiceUnsafeReactivateNamespaceProcedure = "/policy.unsafe.UnsafeService/UnsafeReactivateNamespace"
	// UnsafeServiceUnsafeDeleteNamespaceProcedure is the fully-qualified name of the UnsafeService's
	// UnsafeDeleteNamespace RPC.
	UnsafeServiceUnsafeDeleteNamespaceProcedure = "/policy.unsafe.UnsafeService/UnsafeDeleteNamespace"
	// UnsafeServiceUnsafeUpdateAttributeProcedure is the fully-qualified name of the UnsafeService's
	// UnsafeUpdateAttribute RPC.
	UnsafeServiceUnsafeUpdateAttributeProcedure = "/policy.unsafe.UnsafeService/UnsafeUpdateAttribute"
	// UnsafeServiceUnsafeReactivateAttributeProcedure is the fully-qualified name of the
	// UnsafeService's UnsafeReactivateAttribute RPC.
	UnsafeServiceUnsafeReactivateAttributeProcedure = "/policy.unsafe.UnsafeService/UnsafeReactivateAttribute"
	// UnsafeServiceUnsafeDeleteAttributeProcedure is the fully-qualified name of the UnsafeService's
	// UnsafeDeleteAttribute RPC.
	UnsafeServiceUnsafeDeleteAttributeProcedure = "/policy.unsafe.UnsafeService/UnsafeDeleteAttribute"
	// UnsafeServiceUnsafeUpdateAttributeValueProcedure is the fully-qualified name of the
	// UnsafeService's UnsafeUpdateAttributeValue RPC.
	UnsafeServiceUnsafeUpdateAttributeValueProcedure = "/policy.unsafe.UnsafeService/UnsafeUpdateAttributeValue"
	// UnsafeServiceUnsafeReactivateAttributeValueProcedure is the fully-qualified name of the
	// UnsafeService's UnsafeReactivateAttributeValue RPC.
	UnsafeServiceUnsafeReactivateAttributeValueProcedure = "/policy.unsafe.UnsafeService/UnsafeReactivateAttributeValue"
	// UnsafeServiceUnsafeDeleteAttributeValueProcedure is the fully-qualified name of the
	// UnsafeService's UnsafeDeleteAttributeValue RPC.
	UnsafeServiceUnsafeDeleteAttributeValueProcedure = "/policy.unsafe.UnsafeService/UnsafeDeleteAttributeValue"
	// UnsafeServiceUnsafeDeleteKasKeyProcedure is the fully-qualified name of the UnsafeService's
	// UnsafeDeleteKasKey RPC.
	UnsafeServiceUnsafeDeleteKasKeyProcedure = "/policy.unsafe.UnsafeService/UnsafeDeleteKasKey"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	unsafeServiceServiceDescriptor                              = unsafe.File_policy_unsafe_unsafe_proto.Services().ByName("UnsafeService")
	unsafeServiceUnsafeUpdateNamespaceMethodDescriptor          = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeUpdateNamespace")
	unsafeServiceUnsafeReactivateNamespaceMethodDescriptor      = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeReactivateNamespace")
	unsafeServiceUnsafeDeleteNamespaceMethodDescriptor          = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeDeleteNamespace")
	unsafeServiceUnsafeUpdateAttributeMethodDescriptor          = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeUpdateAttribute")
	unsafeServiceUnsafeReactivateAttributeMethodDescriptor      = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeReactivateAttribute")
	unsafeServiceUnsafeDeleteAttributeMethodDescriptor          = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeDeleteAttribute")
	unsafeServiceUnsafeUpdateAttributeValueMethodDescriptor     = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeUpdateAttributeValue")
	unsafeServiceUnsafeReactivateAttributeValueMethodDescriptor = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeReactivateAttributeValue")
	unsafeServiceUnsafeDeleteAttributeValueMethodDescriptor     = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeDeleteAttributeValue")
	unsafeServiceUnsafeDeleteKasKeyMethodDescriptor             = unsafeServiceServiceDescriptor.Methods().ByName("UnsafeDeleteKasKey")
)

// UnsafeServiceClient is a client for the policy.unsafe.UnsafeService service.
type UnsafeServiceClient interface {
	// --------------------------------------*
	// Namespace RPCs
	// ---------------------------------------
	UnsafeUpdateNamespace(context.Context, *connect.Request[unsafe.UnsafeUpdateNamespaceRequest]) (*connect.Response[unsafe.UnsafeUpdateNamespaceResponse], error)
	UnsafeReactivateNamespace(context.Context, *connect.Request[unsafe.UnsafeReactivateNamespaceRequest]) (*connect.Response[unsafe.UnsafeReactivateNamespaceResponse], error)
	UnsafeDeleteNamespace(context.Context, *connect.Request[unsafe.UnsafeDeleteNamespaceRequest]) (*connect.Response[unsafe.UnsafeDeleteNamespaceResponse], error)
	// --------------------------------------*
	// Attribute RPCs
	// ---------------------------------------
	UnsafeUpdateAttribute(context.Context, *connect.Request[unsafe.UnsafeUpdateAttributeRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeResponse], error)
	UnsafeReactivateAttribute(context.Context, *connect.Request[unsafe.UnsafeReactivateAttributeRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeResponse], error)
	UnsafeDeleteAttribute(context.Context, *connect.Request[unsafe.UnsafeDeleteAttributeRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeResponse], error)
	// --------------------------------------*
	// Value RPCs
	// ---------------------------------------
	UnsafeUpdateAttributeValue(context.Context, *connect.Request[unsafe.UnsafeUpdateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeValueResponse], error)
	UnsafeReactivateAttributeValue(context.Context, *connect.Request[unsafe.UnsafeReactivateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeValueResponse], error)
	UnsafeDeleteAttributeValue(context.Context, *connect.Request[unsafe.UnsafeDeleteAttributeValueRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeValueResponse], error)
	// --------------------------------------*
	// Kas Key RPCs
	// ---------------------------------------
	UnsafeDeleteKasKey(context.Context, *connect.Request[unsafe.UnsafeDeleteKasKeyRequest]) (*connect.Response[unsafe.UnsafeDeleteKasKeyResponse], error)
}

// NewUnsafeServiceClient constructs a client for the policy.unsafe.UnsafeService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewUnsafeServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) UnsafeServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &unsafeServiceClient{
		unsafeUpdateNamespace: connect.NewClient[unsafe.UnsafeUpdateNamespaceRequest, unsafe.UnsafeUpdateNamespaceResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeUpdateNamespaceProcedure,
			connect.WithSchema(unsafeServiceUnsafeUpdateNamespaceMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeReactivateNamespace: connect.NewClient[unsafe.UnsafeReactivateNamespaceRequest, unsafe.UnsafeReactivateNamespaceResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeReactivateNamespaceProcedure,
			connect.WithSchema(unsafeServiceUnsafeReactivateNamespaceMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeDeleteNamespace: connect.NewClient[unsafe.UnsafeDeleteNamespaceRequest, unsafe.UnsafeDeleteNamespaceResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeDeleteNamespaceProcedure,
			connect.WithSchema(unsafeServiceUnsafeDeleteNamespaceMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeUpdateAttribute: connect.NewClient[unsafe.UnsafeUpdateAttributeRequest, unsafe.UnsafeUpdateAttributeResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeUpdateAttributeProcedure,
			connect.WithSchema(unsafeServiceUnsafeUpdateAttributeMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeReactivateAttribute: connect.NewClient[unsafe.UnsafeReactivateAttributeRequest, unsafe.UnsafeReactivateAttributeResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeReactivateAttributeProcedure,
			connect.WithSchema(unsafeServiceUnsafeReactivateAttributeMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeDeleteAttribute: connect.NewClient[unsafe.UnsafeDeleteAttributeRequest, unsafe.UnsafeDeleteAttributeResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeDeleteAttributeProcedure,
			connect.WithSchema(unsafeServiceUnsafeDeleteAttributeMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeUpdateAttributeValue: connect.NewClient[unsafe.UnsafeUpdateAttributeValueRequest, unsafe.UnsafeUpdateAttributeValueResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeUpdateAttributeValueProcedure,
			connect.WithSchema(unsafeServiceUnsafeUpdateAttributeValueMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeReactivateAttributeValue: connect.NewClient[unsafe.UnsafeReactivateAttributeValueRequest, unsafe.UnsafeReactivateAttributeValueResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeReactivateAttributeValueProcedure,
			connect.WithSchema(unsafeServiceUnsafeReactivateAttributeValueMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeDeleteAttributeValue: connect.NewClient[unsafe.UnsafeDeleteAttributeValueRequest, unsafe.UnsafeDeleteAttributeValueResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeDeleteAttributeValueProcedure,
			connect.WithSchema(unsafeServiceUnsafeDeleteAttributeValueMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		unsafeDeleteKasKey: connect.NewClient[unsafe.UnsafeDeleteKasKeyRequest, unsafe.UnsafeDeleteKasKeyResponse](
			httpClient,
			baseURL+UnsafeServiceUnsafeDeleteKasKeyProcedure,
			connect.WithSchema(unsafeServiceUnsafeDeleteKasKeyMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// unsafeServiceClient implements UnsafeServiceClient.
type unsafeServiceClient struct {
	unsafeUpdateNamespace          *connect.Client[unsafe.UnsafeUpdateNamespaceRequest, unsafe.UnsafeUpdateNamespaceResponse]
	unsafeReactivateNamespace      *connect.Client[unsafe.UnsafeReactivateNamespaceRequest, unsafe.UnsafeReactivateNamespaceResponse]
	unsafeDeleteNamespace          *connect.Client[unsafe.UnsafeDeleteNamespaceRequest, unsafe.UnsafeDeleteNamespaceResponse]
	unsafeUpdateAttribute          *connect.Client[unsafe.UnsafeUpdateAttributeRequest, unsafe.UnsafeUpdateAttributeResponse]
	unsafeReactivateAttribute      *connect.Client[unsafe.UnsafeReactivateAttributeRequest, unsafe.UnsafeReactivateAttributeResponse]
	unsafeDeleteAttribute          *connect.Client[unsafe.UnsafeDeleteAttributeRequest, unsafe.UnsafeDeleteAttributeResponse]
	unsafeUpdateAttributeValue     *connect.Client[unsafe.UnsafeUpdateAttributeValueRequest, unsafe.UnsafeUpdateAttributeValueResponse]
	unsafeReactivateAttributeValue *connect.Client[unsafe.UnsafeReactivateAttributeValueRequest, unsafe.UnsafeReactivateAttributeValueResponse]
	unsafeDeleteAttributeValue     *connect.Client[unsafe.UnsafeDeleteAttributeValueRequest, unsafe.UnsafeDeleteAttributeValueResponse]
	unsafeDeleteKasKey             *connect.Client[unsafe.UnsafeDeleteKasKeyRequest, unsafe.UnsafeDeleteKasKeyResponse]
}

// UnsafeUpdateNamespace calls policy.unsafe.UnsafeService.UnsafeUpdateNamespace.
func (c *unsafeServiceClient) UnsafeUpdateNamespace(ctx context.Context, req *connect.Request[unsafe.UnsafeUpdateNamespaceRequest]) (*connect.Response[unsafe.UnsafeUpdateNamespaceResponse], error) {
	return c.unsafeUpdateNamespace.CallUnary(ctx, req)
}

// UnsafeReactivateNamespace calls policy.unsafe.UnsafeService.UnsafeReactivateNamespace.
func (c *unsafeServiceClient) UnsafeReactivateNamespace(ctx context.Context, req *connect.Request[unsafe.UnsafeReactivateNamespaceRequest]) (*connect.Response[unsafe.UnsafeReactivateNamespaceResponse], error) {
	return c.unsafeReactivateNamespace.CallUnary(ctx, req)
}

// UnsafeDeleteNamespace calls policy.unsafe.UnsafeService.UnsafeDeleteNamespace.
func (c *unsafeServiceClient) UnsafeDeleteNamespace(ctx context.Context, req *connect.Request[unsafe.UnsafeDeleteNamespaceRequest]) (*connect.Response[unsafe.UnsafeDeleteNamespaceResponse], error) {
	return c.unsafeDeleteNamespace.CallUnary(ctx, req)
}

// UnsafeUpdateAttribute calls policy.unsafe.UnsafeService.UnsafeUpdateAttribute.
func (c *unsafeServiceClient) UnsafeUpdateAttribute(ctx context.Context, req *connect.Request[unsafe.UnsafeUpdateAttributeRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeResponse], error) {
	return c.unsafeUpdateAttribute.CallUnary(ctx, req)
}

// UnsafeReactivateAttribute calls policy.unsafe.UnsafeService.UnsafeReactivateAttribute.
func (c *unsafeServiceClient) UnsafeReactivateAttribute(ctx context.Context, req *connect.Request[unsafe.UnsafeReactivateAttributeRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeResponse], error) {
	return c.unsafeReactivateAttribute.CallUnary(ctx, req)
}

// UnsafeDeleteAttribute calls policy.unsafe.UnsafeService.UnsafeDeleteAttribute.
func (c *unsafeServiceClient) UnsafeDeleteAttribute(ctx context.Context, req *connect.Request[unsafe.UnsafeDeleteAttributeRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeResponse], error) {
	return c.unsafeDeleteAttribute.CallUnary(ctx, req)
}

// UnsafeUpdateAttributeValue calls policy.unsafe.UnsafeService.UnsafeUpdateAttributeValue.
func (c *unsafeServiceClient) UnsafeUpdateAttributeValue(ctx context.Context, req *connect.Request[unsafe.UnsafeUpdateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeValueResponse], error) {
	return c.unsafeUpdateAttributeValue.CallUnary(ctx, req)
}

// UnsafeReactivateAttributeValue calls policy.unsafe.UnsafeService.UnsafeReactivateAttributeValue.
func (c *unsafeServiceClient) UnsafeReactivateAttributeValue(ctx context.Context, req *connect.Request[unsafe.UnsafeReactivateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeValueResponse], error) {
	return c.unsafeReactivateAttributeValue.CallUnary(ctx, req)
}

// UnsafeDeleteAttributeValue calls policy.unsafe.UnsafeService.UnsafeDeleteAttributeValue.
func (c *unsafeServiceClient) UnsafeDeleteAttributeValue(ctx context.Context, req *connect.Request[unsafe.UnsafeDeleteAttributeValueRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeValueResponse], error) {
	return c.unsafeDeleteAttributeValue.CallUnary(ctx, req)
}

// UnsafeDeleteKasKey calls policy.unsafe.UnsafeService.UnsafeDeleteKasKey.
func (c *unsafeServiceClient) UnsafeDeleteKasKey(ctx context.Context, req *connect.Request[unsafe.UnsafeDeleteKasKeyRequest]) (*connect.Response[unsafe.UnsafeDeleteKasKeyResponse], error) {
	return c.unsafeDeleteKasKey.CallUnary(ctx, req)
}

// UnsafeServiceHandler is an implementation of the policy.unsafe.UnsafeService service.
type UnsafeServiceHandler interface {
	// --------------------------------------*
	// Namespace RPCs
	// ---------------------------------------
	UnsafeUpdateNamespace(context.Context, *connect.Request[unsafe.UnsafeUpdateNamespaceRequest]) (*connect.Response[unsafe.UnsafeUpdateNamespaceResponse], error)
	UnsafeReactivateNamespace(context.Context, *connect.Request[unsafe.UnsafeReactivateNamespaceRequest]) (*connect.Response[unsafe.UnsafeReactivateNamespaceResponse], error)
	UnsafeDeleteNamespace(context.Context, *connect.Request[unsafe.UnsafeDeleteNamespaceRequest]) (*connect.Response[unsafe.UnsafeDeleteNamespaceResponse], error)
	// --------------------------------------*
	// Attribute RPCs
	// ---------------------------------------
	UnsafeUpdateAttribute(context.Context, *connect.Request[unsafe.UnsafeUpdateAttributeRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeResponse], error)
	UnsafeReactivateAttribute(context.Context, *connect.Request[unsafe.UnsafeReactivateAttributeRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeResponse], error)
	UnsafeDeleteAttribute(context.Context, *connect.Request[unsafe.UnsafeDeleteAttributeRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeResponse], error)
	// --------------------------------------*
	// Value RPCs
	// ---------------------------------------
	UnsafeUpdateAttributeValue(context.Context, *connect.Request[unsafe.UnsafeUpdateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeValueResponse], error)
	UnsafeReactivateAttributeValue(context.Context, *connect.Request[unsafe.UnsafeReactivateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeValueResponse], error)
	UnsafeDeleteAttributeValue(context.Context, *connect.Request[unsafe.UnsafeDeleteAttributeValueRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeValueResponse], error)
	// --------------------------------------*
	// Kas Key RPCs
	// ---------------------------------------
	UnsafeDeleteKasKey(context.Context, *connect.Request[unsafe.UnsafeDeleteKasKeyRequest]) (*connect.Response[unsafe.UnsafeDeleteKasKeyResponse], error)
}

// NewUnsafeServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewUnsafeServiceHandler(svc UnsafeServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	unsafeServiceUnsafeUpdateNamespaceHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeUpdateNamespaceProcedure,
		svc.UnsafeUpdateNamespace,
		connect.WithSchema(unsafeServiceUnsafeUpdateNamespaceMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeReactivateNamespaceHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeReactivateNamespaceProcedure,
		svc.UnsafeReactivateNamespace,
		connect.WithSchema(unsafeServiceUnsafeReactivateNamespaceMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeDeleteNamespaceHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeDeleteNamespaceProcedure,
		svc.UnsafeDeleteNamespace,
		connect.WithSchema(unsafeServiceUnsafeDeleteNamespaceMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeUpdateAttributeHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeUpdateAttributeProcedure,
		svc.UnsafeUpdateAttribute,
		connect.WithSchema(unsafeServiceUnsafeUpdateAttributeMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeReactivateAttributeHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeReactivateAttributeProcedure,
		svc.UnsafeReactivateAttribute,
		connect.WithSchema(unsafeServiceUnsafeReactivateAttributeMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeDeleteAttributeHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeDeleteAttributeProcedure,
		svc.UnsafeDeleteAttribute,
		connect.WithSchema(unsafeServiceUnsafeDeleteAttributeMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeUpdateAttributeValueHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeUpdateAttributeValueProcedure,
		svc.UnsafeUpdateAttributeValue,
		connect.WithSchema(unsafeServiceUnsafeUpdateAttributeValueMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeReactivateAttributeValueHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeReactivateAttributeValueProcedure,
		svc.UnsafeReactivateAttributeValue,
		connect.WithSchema(unsafeServiceUnsafeReactivateAttributeValueMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeDeleteAttributeValueHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeDeleteAttributeValueProcedure,
		svc.UnsafeDeleteAttributeValue,
		connect.WithSchema(unsafeServiceUnsafeDeleteAttributeValueMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	unsafeServiceUnsafeDeleteKasKeyHandler := connect.NewUnaryHandler(
		UnsafeServiceUnsafeDeleteKasKeyProcedure,
		svc.UnsafeDeleteKasKey,
		connect.WithSchema(unsafeServiceUnsafeDeleteKasKeyMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/policy.unsafe.UnsafeService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case UnsafeServiceUnsafeUpdateNamespaceProcedure:
			unsafeServiceUnsafeUpdateNamespaceHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeReactivateNamespaceProcedure:
			unsafeServiceUnsafeReactivateNamespaceHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeDeleteNamespaceProcedure:
			unsafeServiceUnsafeDeleteNamespaceHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeUpdateAttributeProcedure:
			unsafeServiceUnsafeUpdateAttributeHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeReactivateAttributeProcedure:
			unsafeServiceUnsafeReactivateAttributeHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeDeleteAttributeProcedure:
			unsafeServiceUnsafeDeleteAttributeHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeUpdateAttributeValueProcedure:
			unsafeServiceUnsafeUpdateAttributeValueHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeReactivateAttributeValueProcedure:
			unsafeServiceUnsafeReactivateAttributeValueHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeDeleteAttributeValueProcedure:
			unsafeServiceUnsafeDeleteAttributeValueHandler.ServeHTTP(w, r)
		case UnsafeServiceUnsafeDeleteKasKeyProcedure:
			unsafeServiceUnsafeDeleteKasKeyHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedUnsafeServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedUnsafeServiceHandler struct{}

func (UnimplementedUnsafeServiceHandler) UnsafeUpdateNamespace(context.Context, *connect.Request[unsafe.UnsafeUpdateNamespaceRequest]) (*connect.Response[unsafe.UnsafeUpdateNamespaceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeUpdateNamespace is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeReactivateNamespace(context.Context, *connect.Request[unsafe.UnsafeReactivateNamespaceRequest]) (*connect.Response[unsafe.UnsafeReactivateNamespaceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeReactivateNamespace is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeDeleteNamespace(context.Context, *connect.Request[unsafe.UnsafeDeleteNamespaceRequest]) (*connect.Response[unsafe.UnsafeDeleteNamespaceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeDeleteNamespace is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeUpdateAttribute(context.Context, *connect.Request[unsafe.UnsafeUpdateAttributeRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeUpdateAttribute is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeReactivateAttribute(context.Context, *connect.Request[unsafe.UnsafeReactivateAttributeRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeReactivateAttribute is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeDeleteAttribute(context.Context, *connect.Request[unsafe.UnsafeDeleteAttributeRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeDeleteAttribute is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeUpdateAttributeValue(context.Context, *connect.Request[unsafe.UnsafeUpdateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeValueResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeUpdateAttributeValue is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeReactivateAttributeValue(context.Context, *connect.Request[unsafe.UnsafeReactivateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeValueResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeReactivateAttributeValue is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeDeleteAttributeValue(context.Context, *connect.Request[unsafe.UnsafeDeleteAttributeValueRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeValueResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeDeleteAttributeValue is not implemented"))
}

func (UnimplementedUnsafeServiceHandler) UnsafeDeleteKasKey(context.Context, *connect.Request[unsafe.UnsafeDeleteKasKeyRequest]) (*connect.Response[unsafe.UnsafeDeleteKasKeyResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.unsafe.UnsafeService.UnsafeDeleteKasKey is not implemented"))
}
