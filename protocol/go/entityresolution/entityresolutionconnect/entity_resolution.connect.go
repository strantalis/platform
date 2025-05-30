// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: entityresolution/entity_resolution.proto

package entityresolutionconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	entityresolution "github.com/opentdf/platform/protocol/go/entityresolution"
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
	// EntityResolutionServiceName is the fully-qualified name of the EntityResolutionService service.
	EntityResolutionServiceName = "entityresolution.EntityResolutionService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// EntityResolutionServiceResolveEntitiesProcedure is the fully-qualified name of the
	// EntityResolutionService's ResolveEntities RPC.
	EntityResolutionServiceResolveEntitiesProcedure = "/entityresolution.EntityResolutionService/ResolveEntities"
	// EntityResolutionServiceCreateEntityChainFromJwtProcedure is the fully-qualified name of the
	// EntityResolutionService's CreateEntityChainFromJwt RPC.
	EntityResolutionServiceCreateEntityChainFromJwtProcedure = "/entityresolution.EntityResolutionService/CreateEntityChainFromJwt"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	entityResolutionServiceServiceDescriptor                        = entityresolution.File_entityresolution_entity_resolution_proto.Services().ByName("EntityResolutionService")
	entityResolutionServiceResolveEntitiesMethodDescriptor          = entityResolutionServiceServiceDescriptor.Methods().ByName("ResolveEntities")
	entityResolutionServiceCreateEntityChainFromJwtMethodDescriptor = entityResolutionServiceServiceDescriptor.Methods().ByName("CreateEntityChainFromJwt")
)

// EntityResolutionServiceClient is a client for the entityresolution.EntityResolutionService
// service.
type EntityResolutionServiceClient interface {
	// Deprecated: use v2 ResolveEntities instead
	ResolveEntities(context.Context, *connect.Request[entityresolution.ResolveEntitiesRequest]) (*connect.Response[entityresolution.ResolveEntitiesResponse], error)
	// Deprecated: use v2 CreateEntityChainsFromTokens instead
	CreateEntityChainFromJwt(context.Context, *connect.Request[entityresolution.CreateEntityChainFromJwtRequest]) (*connect.Response[entityresolution.CreateEntityChainFromJwtResponse], error)
}

// NewEntityResolutionServiceClient constructs a client for the
// entityresolution.EntityResolutionService service. By default, it uses the Connect protocol with
// the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed requests. To use
// the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewEntityResolutionServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) EntityResolutionServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &entityResolutionServiceClient{
		resolveEntities: connect.NewClient[entityresolution.ResolveEntitiesRequest, entityresolution.ResolveEntitiesResponse](
			httpClient,
			baseURL+EntityResolutionServiceResolveEntitiesProcedure,
			connect.WithSchema(entityResolutionServiceResolveEntitiesMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		createEntityChainFromJwt: connect.NewClient[entityresolution.CreateEntityChainFromJwtRequest, entityresolution.CreateEntityChainFromJwtResponse](
			httpClient,
			baseURL+EntityResolutionServiceCreateEntityChainFromJwtProcedure,
			connect.WithSchema(entityResolutionServiceCreateEntityChainFromJwtMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// entityResolutionServiceClient implements EntityResolutionServiceClient.
type entityResolutionServiceClient struct {
	resolveEntities          *connect.Client[entityresolution.ResolveEntitiesRequest, entityresolution.ResolveEntitiesResponse]
	createEntityChainFromJwt *connect.Client[entityresolution.CreateEntityChainFromJwtRequest, entityresolution.CreateEntityChainFromJwtResponse]
}

// ResolveEntities calls entityresolution.EntityResolutionService.ResolveEntities.
func (c *entityResolutionServiceClient) ResolveEntities(ctx context.Context, req *connect.Request[entityresolution.ResolveEntitiesRequest]) (*connect.Response[entityresolution.ResolveEntitiesResponse], error) {
	return c.resolveEntities.CallUnary(ctx, req)
}

// CreateEntityChainFromJwt calls entityresolution.EntityResolutionService.CreateEntityChainFromJwt.
func (c *entityResolutionServiceClient) CreateEntityChainFromJwt(ctx context.Context, req *connect.Request[entityresolution.CreateEntityChainFromJwtRequest]) (*connect.Response[entityresolution.CreateEntityChainFromJwtResponse], error) {
	return c.createEntityChainFromJwt.CallUnary(ctx, req)
}

// EntityResolutionServiceHandler is an implementation of the
// entityresolution.EntityResolutionService service.
type EntityResolutionServiceHandler interface {
	// Deprecated: use v2 ResolveEntities instead
	ResolveEntities(context.Context, *connect.Request[entityresolution.ResolveEntitiesRequest]) (*connect.Response[entityresolution.ResolveEntitiesResponse], error)
	// Deprecated: use v2 CreateEntityChainsFromTokens instead
	CreateEntityChainFromJwt(context.Context, *connect.Request[entityresolution.CreateEntityChainFromJwtRequest]) (*connect.Response[entityresolution.CreateEntityChainFromJwtResponse], error)
}

// NewEntityResolutionServiceHandler builds an HTTP handler from the service implementation. It
// returns the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewEntityResolutionServiceHandler(svc EntityResolutionServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	entityResolutionServiceResolveEntitiesHandler := connect.NewUnaryHandler(
		EntityResolutionServiceResolveEntitiesProcedure,
		svc.ResolveEntities,
		connect.WithSchema(entityResolutionServiceResolveEntitiesMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	entityResolutionServiceCreateEntityChainFromJwtHandler := connect.NewUnaryHandler(
		EntityResolutionServiceCreateEntityChainFromJwtProcedure,
		svc.CreateEntityChainFromJwt,
		connect.WithSchema(entityResolutionServiceCreateEntityChainFromJwtMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/entityresolution.EntityResolutionService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EntityResolutionServiceResolveEntitiesProcedure:
			entityResolutionServiceResolveEntitiesHandler.ServeHTTP(w, r)
		case EntityResolutionServiceCreateEntityChainFromJwtProcedure:
			entityResolutionServiceCreateEntityChainFromJwtHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedEntityResolutionServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedEntityResolutionServiceHandler struct{}

func (UnimplementedEntityResolutionServiceHandler) ResolveEntities(context.Context, *connect.Request[entityresolution.ResolveEntitiesRequest]) (*connect.Response[entityresolution.ResolveEntitiesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("entityresolution.EntityResolutionService.ResolveEntities is not implemented"))
}

func (UnimplementedEntityResolutionServiceHandler) CreateEntityChainFromJwt(context.Context, *connect.Request[entityresolution.CreateEntityChainFromJwtRequest]) (*connect.Response[entityresolution.CreateEntityChainFromJwtResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("entityresolution.EntityResolutionService.CreateEntityChainFromJwt is not implemented"))
}
