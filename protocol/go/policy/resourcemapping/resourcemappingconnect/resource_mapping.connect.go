// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: policy/resourcemapping/resource_mapping.proto

package resourcemappingconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	resourcemapping "github.com/opentdf/platform/protocol/go/policy/resourcemapping"
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
	// ResourceMappingServiceName is the fully-qualified name of the ResourceMappingService service.
	ResourceMappingServiceName = "policy.resourcemapping.ResourceMappingService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ResourceMappingServiceListResourceMappingGroupsProcedure is the fully-qualified name of the
	// ResourceMappingService's ListResourceMappingGroups RPC.
	ResourceMappingServiceListResourceMappingGroupsProcedure = "/policy.resourcemapping.ResourceMappingService/ListResourceMappingGroups"
	// ResourceMappingServiceGetResourceMappingGroupProcedure is the fully-qualified name of the
	// ResourceMappingService's GetResourceMappingGroup RPC.
	ResourceMappingServiceGetResourceMappingGroupProcedure = "/policy.resourcemapping.ResourceMappingService/GetResourceMappingGroup"
	// ResourceMappingServiceCreateResourceMappingGroupProcedure is the fully-qualified name of the
	// ResourceMappingService's CreateResourceMappingGroup RPC.
	ResourceMappingServiceCreateResourceMappingGroupProcedure = "/policy.resourcemapping.ResourceMappingService/CreateResourceMappingGroup"
	// ResourceMappingServiceUpdateResourceMappingGroupProcedure is the fully-qualified name of the
	// ResourceMappingService's UpdateResourceMappingGroup RPC.
	ResourceMappingServiceUpdateResourceMappingGroupProcedure = "/policy.resourcemapping.ResourceMappingService/UpdateResourceMappingGroup"
	// ResourceMappingServiceDeleteResourceMappingGroupProcedure is the fully-qualified name of the
	// ResourceMappingService's DeleteResourceMappingGroup RPC.
	ResourceMappingServiceDeleteResourceMappingGroupProcedure = "/policy.resourcemapping.ResourceMappingService/DeleteResourceMappingGroup"
	// ResourceMappingServiceListResourceMappingsProcedure is the fully-qualified name of the
	// ResourceMappingService's ListResourceMappings RPC.
	ResourceMappingServiceListResourceMappingsProcedure = "/policy.resourcemapping.ResourceMappingService/ListResourceMappings"
	// ResourceMappingServiceListResourceMappingsByGroupFqnsProcedure is the fully-qualified name of the
	// ResourceMappingService's ListResourceMappingsByGroupFqns RPC.
	ResourceMappingServiceListResourceMappingsByGroupFqnsProcedure = "/policy.resourcemapping.ResourceMappingService/ListResourceMappingsByGroupFqns"
	// ResourceMappingServiceGetResourceMappingProcedure is the fully-qualified name of the
	// ResourceMappingService's GetResourceMapping RPC.
	ResourceMappingServiceGetResourceMappingProcedure = "/policy.resourcemapping.ResourceMappingService/GetResourceMapping"
	// ResourceMappingServiceCreateResourceMappingProcedure is the fully-qualified name of the
	// ResourceMappingService's CreateResourceMapping RPC.
	ResourceMappingServiceCreateResourceMappingProcedure = "/policy.resourcemapping.ResourceMappingService/CreateResourceMapping"
	// ResourceMappingServiceUpdateResourceMappingProcedure is the fully-qualified name of the
	// ResourceMappingService's UpdateResourceMapping RPC.
	ResourceMappingServiceUpdateResourceMappingProcedure = "/policy.resourcemapping.ResourceMappingService/UpdateResourceMapping"
	// ResourceMappingServiceDeleteResourceMappingProcedure is the fully-qualified name of the
	// ResourceMappingService's DeleteResourceMapping RPC.
	ResourceMappingServiceDeleteResourceMappingProcedure = "/policy.resourcemapping.ResourceMappingService/DeleteResourceMapping"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	resourceMappingServiceServiceDescriptor                               = resourcemapping.File_policy_resourcemapping_resource_mapping_proto.Services().ByName("ResourceMappingService")
	resourceMappingServiceListResourceMappingGroupsMethodDescriptor       = resourceMappingServiceServiceDescriptor.Methods().ByName("ListResourceMappingGroups")
	resourceMappingServiceGetResourceMappingGroupMethodDescriptor         = resourceMappingServiceServiceDescriptor.Methods().ByName("GetResourceMappingGroup")
	resourceMappingServiceCreateResourceMappingGroupMethodDescriptor      = resourceMappingServiceServiceDescriptor.Methods().ByName("CreateResourceMappingGroup")
	resourceMappingServiceUpdateResourceMappingGroupMethodDescriptor      = resourceMappingServiceServiceDescriptor.Methods().ByName("UpdateResourceMappingGroup")
	resourceMappingServiceDeleteResourceMappingGroupMethodDescriptor      = resourceMappingServiceServiceDescriptor.Methods().ByName("DeleteResourceMappingGroup")
	resourceMappingServiceListResourceMappingsMethodDescriptor            = resourceMappingServiceServiceDescriptor.Methods().ByName("ListResourceMappings")
	resourceMappingServiceListResourceMappingsByGroupFqnsMethodDescriptor = resourceMappingServiceServiceDescriptor.Methods().ByName("ListResourceMappingsByGroupFqns")
	resourceMappingServiceGetResourceMappingMethodDescriptor              = resourceMappingServiceServiceDescriptor.Methods().ByName("GetResourceMapping")
	resourceMappingServiceCreateResourceMappingMethodDescriptor           = resourceMappingServiceServiceDescriptor.Methods().ByName("CreateResourceMapping")
	resourceMappingServiceUpdateResourceMappingMethodDescriptor           = resourceMappingServiceServiceDescriptor.Methods().ByName("UpdateResourceMapping")
	resourceMappingServiceDeleteResourceMappingMethodDescriptor           = resourceMappingServiceServiceDescriptor.Methods().ByName("DeleteResourceMapping")
)

// ResourceMappingServiceClient is a client for the policy.resourcemapping.ResourceMappingService
// service.
type ResourceMappingServiceClient interface {
	ListResourceMappingGroups(context.Context, *connect.Request[resourcemapping.ListResourceMappingGroupsRequest]) (*connect.Response[resourcemapping.ListResourceMappingGroupsResponse], error)
	GetResourceMappingGroup(context.Context, *connect.Request[resourcemapping.GetResourceMappingGroupRequest]) (*connect.Response[resourcemapping.GetResourceMappingGroupResponse], error)
	CreateResourceMappingGroup(context.Context, *connect.Request[resourcemapping.CreateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.CreateResourceMappingGroupResponse], error)
	UpdateResourceMappingGroup(context.Context, *connect.Request[resourcemapping.UpdateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingGroupResponse], error)
	DeleteResourceMappingGroup(context.Context, *connect.Request[resourcemapping.DeleteResourceMappingGroupRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingGroupResponse], error)
	ListResourceMappings(context.Context, *connect.Request[resourcemapping.ListResourceMappingsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsResponse], error)
	ListResourceMappingsByGroupFqns(context.Context, *connect.Request[resourcemapping.ListResourceMappingsByGroupFqnsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsByGroupFqnsResponse], error)
	GetResourceMapping(context.Context, *connect.Request[resourcemapping.GetResourceMappingRequest]) (*connect.Response[resourcemapping.GetResourceMappingResponse], error)
	CreateResourceMapping(context.Context, *connect.Request[resourcemapping.CreateResourceMappingRequest]) (*connect.Response[resourcemapping.CreateResourceMappingResponse], error)
	UpdateResourceMapping(context.Context, *connect.Request[resourcemapping.UpdateResourceMappingRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingResponse], error)
	DeleteResourceMapping(context.Context, *connect.Request[resourcemapping.DeleteResourceMappingRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingResponse], error)
}

// NewResourceMappingServiceClient constructs a client for the
// policy.resourcemapping.ResourceMappingService service. By default, it uses the Connect protocol
// with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed requests. To
// use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or connect.WithGRPCWeb()
// options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewResourceMappingServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ResourceMappingServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &resourceMappingServiceClient{
		listResourceMappingGroups: connect.NewClient[resourcemapping.ListResourceMappingGroupsRequest, resourcemapping.ListResourceMappingGroupsResponse](
			httpClient,
			baseURL+ResourceMappingServiceListResourceMappingGroupsProcedure,
			connect.WithSchema(resourceMappingServiceListResourceMappingGroupsMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getResourceMappingGroup: connect.NewClient[resourcemapping.GetResourceMappingGroupRequest, resourcemapping.GetResourceMappingGroupResponse](
			httpClient,
			baseURL+ResourceMappingServiceGetResourceMappingGroupProcedure,
			connect.WithSchema(resourceMappingServiceGetResourceMappingGroupMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		createResourceMappingGroup: connect.NewClient[resourcemapping.CreateResourceMappingGroupRequest, resourcemapping.CreateResourceMappingGroupResponse](
			httpClient,
			baseURL+ResourceMappingServiceCreateResourceMappingGroupProcedure,
			connect.WithSchema(resourceMappingServiceCreateResourceMappingGroupMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		updateResourceMappingGroup: connect.NewClient[resourcemapping.UpdateResourceMappingGroupRequest, resourcemapping.UpdateResourceMappingGroupResponse](
			httpClient,
			baseURL+ResourceMappingServiceUpdateResourceMappingGroupProcedure,
			connect.WithSchema(resourceMappingServiceUpdateResourceMappingGroupMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		deleteResourceMappingGroup: connect.NewClient[resourcemapping.DeleteResourceMappingGroupRequest, resourcemapping.DeleteResourceMappingGroupResponse](
			httpClient,
			baseURL+ResourceMappingServiceDeleteResourceMappingGroupProcedure,
			connect.WithSchema(resourceMappingServiceDeleteResourceMappingGroupMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		listResourceMappings: connect.NewClient[resourcemapping.ListResourceMappingsRequest, resourcemapping.ListResourceMappingsResponse](
			httpClient,
			baseURL+ResourceMappingServiceListResourceMappingsProcedure,
			connect.WithSchema(resourceMappingServiceListResourceMappingsMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		listResourceMappingsByGroupFqns: connect.NewClient[resourcemapping.ListResourceMappingsByGroupFqnsRequest, resourcemapping.ListResourceMappingsByGroupFqnsResponse](
			httpClient,
			baseURL+ResourceMappingServiceListResourceMappingsByGroupFqnsProcedure,
			connect.WithSchema(resourceMappingServiceListResourceMappingsByGroupFqnsMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getResourceMapping: connect.NewClient[resourcemapping.GetResourceMappingRequest, resourcemapping.GetResourceMappingResponse](
			httpClient,
			baseURL+ResourceMappingServiceGetResourceMappingProcedure,
			connect.WithSchema(resourceMappingServiceGetResourceMappingMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		createResourceMapping: connect.NewClient[resourcemapping.CreateResourceMappingRequest, resourcemapping.CreateResourceMappingResponse](
			httpClient,
			baseURL+ResourceMappingServiceCreateResourceMappingProcedure,
			connect.WithSchema(resourceMappingServiceCreateResourceMappingMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		updateResourceMapping: connect.NewClient[resourcemapping.UpdateResourceMappingRequest, resourcemapping.UpdateResourceMappingResponse](
			httpClient,
			baseURL+ResourceMappingServiceUpdateResourceMappingProcedure,
			connect.WithSchema(resourceMappingServiceUpdateResourceMappingMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		deleteResourceMapping: connect.NewClient[resourcemapping.DeleteResourceMappingRequest, resourcemapping.DeleteResourceMappingResponse](
			httpClient,
			baseURL+ResourceMappingServiceDeleteResourceMappingProcedure,
			connect.WithSchema(resourceMappingServiceDeleteResourceMappingMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// resourceMappingServiceClient implements ResourceMappingServiceClient.
type resourceMappingServiceClient struct {
	listResourceMappingGroups       *connect.Client[resourcemapping.ListResourceMappingGroupsRequest, resourcemapping.ListResourceMappingGroupsResponse]
	getResourceMappingGroup         *connect.Client[resourcemapping.GetResourceMappingGroupRequest, resourcemapping.GetResourceMappingGroupResponse]
	createResourceMappingGroup      *connect.Client[resourcemapping.CreateResourceMappingGroupRequest, resourcemapping.CreateResourceMappingGroupResponse]
	updateResourceMappingGroup      *connect.Client[resourcemapping.UpdateResourceMappingGroupRequest, resourcemapping.UpdateResourceMappingGroupResponse]
	deleteResourceMappingGroup      *connect.Client[resourcemapping.DeleteResourceMappingGroupRequest, resourcemapping.DeleteResourceMappingGroupResponse]
	listResourceMappings            *connect.Client[resourcemapping.ListResourceMappingsRequest, resourcemapping.ListResourceMappingsResponse]
	listResourceMappingsByGroupFqns *connect.Client[resourcemapping.ListResourceMappingsByGroupFqnsRequest, resourcemapping.ListResourceMappingsByGroupFqnsResponse]
	getResourceMapping              *connect.Client[resourcemapping.GetResourceMappingRequest, resourcemapping.GetResourceMappingResponse]
	createResourceMapping           *connect.Client[resourcemapping.CreateResourceMappingRequest, resourcemapping.CreateResourceMappingResponse]
	updateResourceMapping           *connect.Client[resourcemapping.UpdateResourceMappingRequest, resourcemapping.UpdateResourceMappingResponse]
	deleteResourceMapping           *connect.Client[resourcemapping.DeleteResourceMappingRequest, resourcemapping.DeleteResourceMappingResponse]
}

// ListResourceMappingGroups calls
// policy.resourcemapping.ResourceMappingService.ListResourceMappingGroups.
func (c *resourceMappingServiceClient) ListResourceMappingGroups(ctx context.Context, req *connect.Request[resourcemapping.ListResourceMappingGroupsRequest]) (*connect.Response[resourcemapping.ListResourceMappingGroupsResponse], error) {
	return c.listResourceMappingGroups.CallUnary(ctx, req)
}

// GetResourceMappingGroup calls
// policy.resourcemapping.ResourceMappingService.GetResourceMappingGroup.
func (c *resourceMappingServiceClient) GetResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.GetResourceMappingGroupRequest]) (*connect.Response[resourcemapping.GetResourceMappingGroupResponse], error) {
	return c.getResourceMappingGroup.CallUnary(ctx, req)
}

// CreateResourceMappingGroup calls
// policy.resourcemapping.ResourceMappingService.CreateResourceMappingGroup.
func (c *resourceMappingServiceClient) CreateResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.CreateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.CreateResourceMappingGroupResponse], error) {
	return c.createResourceMappingGroup.CallUnary(ctx, req)
}

// UpdateResourceMappingGroup calls
// policy.resourcemapping.ResourceMappingService.UpdateResourceMappingGroup.
func (c *resourceMappingServiceClient) UpdateResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.UpdateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingGroupResponse], error) {
	return c.updateResourceMappingGroup.CallUnary(ctx, req)
}

// DeleteResourceMappingGroup calls
// policy.resourcemapping.ResourceMappingService.DeleteResourceMappingGroup.
func (c *resourceMappingServiceClient) DeleteResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.DeleteResourceMappingGroupRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingGroupResponse], error) {
	return c.deleteResourceMappingGroup.CallUnary(ctx, req)
}

// ListResourceMappings calls policy.resourcemapping.ResourceMappingService.ListResourceMappings.
func (c *resourceMappingServiceClient) ListResourceMappings(ctx context.Context, req *connect.Request[resourcemapping.ListResourceMappingsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsResponse], error) {
	return c.listResourceMappings.CallUnary(ctx, req)
}

// ListResourceMappingsByGroupFqns calls
// policy.resourcemapping.ResourceMappingService.ListResourceMappingsByGroupFqns.
func (c *resourceMappingServiceClient) ListResourceMappingsByGroupFqns(ctx context.Context, req *connect.Request[resourcemapping.ListResourceMappingsByGroupFqnsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsByGroupFqnsResponse], error) {
	return c.listResourceMappingsByGroupFqns.CallUnary(ctx, req)
}

// GetResourceMapping calls policy.resourcemapping.ResourceMappingService.GetResourceMapping.
func (c *resourceMappingServiceClient) GetResourceMapping(ctx context.Context, req *connect.Request[resourcemapping.GetResourceMappingRequest]) (*connect.Response[resourcemapping.GetResourceMappingResponse], error) {
	return c.getResourceMapping.CallUnary(ctx, req)
}

// CreateResourceMapping calls policy.resourcemapping.ResourceMappingService.CreateResourceMapping.
func (c *resourceMappingServiceClient) CreateResourceMapping(ctx context.Context, req *connect.Request[resourcemapping.CreateResourceMappingRequest]) (*connect.Response[resourcemapping.CreateResourceMappingResponse], error) {
	return c.createResourceMapping.CallUnary(ctx, req)
}

// UpdateResourceMapping calls policy.resourcemapping.ResourceMappingService.UpdateResourceMapping.
func (c *resourceMappingServiceClient) UpdateResourceMapping(ctx context.Context, req *connect.Request[resourcemapping.UpdateResourceMappingRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingResponse], error) {
	return c.updateResourceMapping.CallUnary(ctx, req)
}

// DeleteResourceMapping calls policy.resourcemapping.ResourceMappingService.DeleteResourceMapping.
func (c *resourceMappingServiceClient) DeleteResourceMapping(ctx context.Context, req *connect.Request[resourcemapping.DeleteResourceMappingRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingResponse], error) {
	return c.deleteResourceMapping.CallUnary(ctx, req)
}

// ResourceMappingServiceHandler is an implementation of the
// policy.resourcemapping.ResourceMappingService service.
type ResourceMappingServiceHandler interface {
	ListResourceMappingGroups(context.Context, *connect.Request[resourcemapping.ListResourceMappingGroupsRequest]) (*connect.Response[resourcemapping.ListResourceMappingGroupsResponse], error)
	GetResourceMappingGroup(context.Context, *connect.Request[resourcemapping.GetResourceMappingGroupRequest]) (*connect.Response[resourcemapping.GetResourceMappingGroupResponse], error)
	CreateResourceMappingGroup(context.Context, *connect.Request[resourcemapping.CreateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.CreateResourceMappingGroupResponse], error)
	UpdateResourceMappingGroup(context.Context, *connect.Request[resourcemapping.UpdateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingGroupResponse], error)
	DeleteResourceMappingGroup(context.Context, *connect.Request[resourcemapping.DeleteResourceMappingGroupRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingGroupResponse], error)
	ListResourceMappings(context.Context, *connect.Request[resourcemapping.ListResourceMappingsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsResponse], error)
	ListResourceMappingsByGroupFqns(context.Context, *connect.Request[resourcemapping.ListResourceMappingsByGroupFqnsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsByGroupFqnsResponse], error)
	GetResourceMapping(context.Context, *connect.Request[resourcemapping.GetResourceMappingRequest]) (*connect.Response[resourcemapping.GetResourceMappingResponse], error)
	CreateResourceMapping(context.Context, *connect.Request[resourcemapping.CreateResourceMappingRequest]) (*connect.Response[resourcemapping.CreateResourceMappingResponse], error)
	UpdateResourceMapping(context.Context, *connect.Request[resourcemapping.UpdateResourceMappingRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingResponse], error)
	DeleteResourceMapping(context.Context, *connect.Request[resourcemapping.DeleteResourceMappingRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingResponse], error)
}

// NewResourceMappingServiceHandler builds an HTTP handler from the service implementation. It
// returns the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewResourceMappingServiceHandler(svc ResourceMappingServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	resourceMappingServiceListResourceMappingGroupsHandler := connect.NewUnaryHandler(
		ResourceMappingServiceListResourceMappingGroupsProcedure,
		svc.ListResourceMappingGroups,
		connect.WithSchema(resourceMappingServiceListResourceMappingGroupsMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceGetResourceMappingGroupHandler := connect.NewUnaryHandler(
		ResourceMappingServiceGetResourceMappingGroupProcedure,
		svc.GetResourceMappingGroup,
		connect.WithSchema(resourceMappingServiceGetResourceMappingGroupMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceCreateResourceMappingGroupHandler := connect.NewUnaryHandler(
		ResourceMappingServiceCreateResourceMappingGroupProcedure,
		svc.CreateResourceMappingGroup,
		connect.WithSchema(resourceMappingServiceCreateResourceMappingGroupMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceUpdateResourceMappingGroupHandler := connect.NewUnaryHandler(
		ResourceMappingServiceUpdateResourceMappingGroupProcedure,
		svc.UpdateResourceMappingGroup,
		connect.WithSchema(resourceMappingServiceUpdateResourceMappingGroupMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceDeleteResourceMappingGroupHandler := connect.NewUnaryHandler(
		ResourceMappingServiceDeleteResourceMappingGroupProcedure,
		svc.DeleteResourceMappingGroup,
		connect.WithSchema(resourceMappingServiceDeleteResourceMappingGroupMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceListResourceMappingsHandler := connect.NewUnaryHandler(
		ResourceMappingServiceListResourceMappingsProcedure,
		svc.ListResourceMappings,
		connect.WithSchema(resourceMappingServiceListResourceMappingsMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceListResourceMappingsByGroupFqnsHandler := connect.NewUnaryHandler(
		ResourceMappingServiceListResourceMappingsByGroupFqnsProcedure,
		svc.ListResourceMappingsByGroupFqns,
		connect.WithSchema(resourceMappingServiceListResourceMappingsByGroupFqnsMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceGetResourceMappingHandler := connect.NewUnaryHandler(
		ResourceMappingServiceGetResourceMappingProcedure,
		svc.GetResourceMapping,
		connect.WithSchema(resourceMappingServiceGetResourceMappingMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceCreateResourceMappingHandler := connect.NewUnaryHandler(
		ResourceMappingServiceCreateResourceMappingProcedure,
		svc.CreateResourceMapping,
		connect.WithSchema(resourceMappingServiceCreateResourceMappingMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceUpdateResourceMappingHandler := connect.NewUnaryHandler(
		ResourceMappingServiceUpdateResourceMappingProcedure,
		svc.UpdateResourceMapping,
		connect.WithSchema(resourceMappingServiceUpdateResourceMappingMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	resourceMappingServiceDeleteResourceMappingHandler := connect.NewUnaryHandler(
		ResourceMappingServiceDeleteResourceMappingProcedure,
		svc.DeleteResourceMapping,
		connect.WithSchema(resourceMappingServiceDeleteResourceMappingMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/policy.resourcemapping.ResourceMappingService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ResourceMappingServiceListResourceMappingGroupsProcedure:
			resourceMappingServiceListResourceMappingGroupsHandler.ServeHTTP(w, r)
		case ResourceMappingServiceGetResourceMappingGroupProcedure:
			resourceMappingServiceGetResourceMappingGroupHandler.ServeHTTP(w, r)
		case ResourceMappingServiceCreateResourceMappingGroupProcedure:
			resourceMappingServiceCreateResourceMappingGroupHandler.ServeHTTP(w, r)
		case ResourceMappingServiceUpdateResourceMappingGroupProcedure:
			resourceMappingServiceUpdateResourceMappingGroupHandler.ServeHTTP(w, r)
		case ResourceMappingServiceDeleteResourceMappingGroupProcedure:
			resourceMappingServiceDeleteResourceMappingGroupHandler.ServeHTTP(w, r)
		case ResourceMappingServiceListResourceMappingsProcedure:
			resourceMappingServiceListResourceMappingsHandler.ServeHTTP(w, r)
		case ResourceMappingServiceListResourceMappingsByGroupFqnsProcedure:
			resourceMappingServiceListResourceMappingsByGroupFqnsHandler.ServeHTTP(w, r)
		case ResourceMappingServiceGetResourceMappingProcedure:
			resourceMappingServiceGetResourceMappingHandler.ServeHTTP(w, r)
		case ResourceMappingServiceCreateResourceMappingProcedure:
			resourceMappingServiceCreateResourceMappingHandler.ServeHTTP(w, r)
		case ResourceMappingServiceUpdateResourceMappingProcedure:
			resourceMappingServiceUpdateResourceMappingHandler.ServeHTTP(w, r)
		case ResourceMappingServiceDeleteResourceMappingProcedure:
			resourceMappingServiceDeleteResourceMappingHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedResourceMappingServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedResourceMappingServiceHandler struct{}

func (UnimplementedResourceMappingServiceHandler) ListResourceMappingGroups(context.Context, *connect.Request[resourcemapping.ListResourceMappingGroupsRequest]) (*connect.Response[resourcemapping.ListResourceMappingGroupsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.ListResourceMappingGroups is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) GetResourceMappingGroup(context.Context, *connect.Request[resourcemapping.GetResourceMappingGroupRequest]) (*connect.Response[resourcemapping.GetResourceMappingGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.GetResourceMappingGroup is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) CreateResourceMappingGroup(context.Context, *connect.Request[resourcemapping.CreateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.CreateResourceMappingGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.CreateResourceMappingGroup is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) UpdateResourceMappingGroup(context.Context, *connect.Request[resourcemapping.UpdateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.UpdateResourceMappingGroup is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) DeleteResourceMappingGroup(context.Context, *connect.Request[resourcemapping.DeleteResourceMappingGroupRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.DeleteResourceMappingGroup is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) ListResourceMappings(context.Context, *connect.Request[resourcemapping.ListResourceMappingsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.ListResourceMappings is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) ListResourceMappingsByGroupFqns(context.Context, *connect.Request[resourcemapping.ListResourceMappingsByGroupFqnsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsByGroupFqnsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.ListResourceMappingsByGroupFqns is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) GetResourceMapping(context.Context, *connect.Request[resourcemapping.GetResourceMappingRequest]) (*connect.Response[resourcemapping.GetResourceMappingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.GetResourceMapping is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) CreateResourceMapping(context.Context, *connect.Request[resourcemapping.CreateResourceMappingRequest]) (*connect.Response[resourcemapping.CreateResourceMappingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.CreateResourceMapping is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) UpdateResourceMapping(context.Context, *connect.Request[resourcemapping.UpdateResourceMappingRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.UpdateResourceMapping is not implemented"))
}

func (UnimplementedResourceMappingServiceHandler) DeleteResourceMapping(context.Context, *connect.Request[resourcemapping.DeleteResourceMappingRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("policy.resourcemapping.ResourceMappingService.DeleteResourceMapping is not implemented"))
}