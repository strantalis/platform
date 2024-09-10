package resourcemapping

import (
	"context"

	"connectrpc.com/connect"
	"github.com/opentdf/platform/protocol/go/policy/resourcemapping"
)

type ResourceMappingServiceGRPCGateway struct { //nolint:revive // ResourceMappingService is a valid name for this struct
	resourcemapping.UnimplementedResourceMappingServiceServer
	ConnectRPC ResourceMappingService
}

/*
	Resource Mapping Groups
*/

func (s ResourceMappingServiceGRPCGateway) ListResourceMappingGroups(ctx context.Context, req *resourcemapping.ListResourceMappingGroupsRequest) (*resourcemapping.ListResourceMappingGroupsResponse, error) {
	rsp, err := s.ConnectRPC.ListResourceMappingGroups(ctx, &connect.Request[resourcemapping.ListResourceMappingGroupsRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) GetResourceMappingGroup(ctx context.Context, req *resourcemapping.GetResourceMappingGroupRequest) (*resourcemapping.GetResourceMappingGroupResponse, error) {
	rsp, err := s.ConnectRPC.GetResourceMappingGroup(ctx, &connect.Request[resourcemapping.GetResourceMappingGroupRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) CreateResourceMappingGroup(ctx context.Context, req *resourcemapping.CreateResourceMappingGroupRequest) (*resourcemapping.CreateResourceMappingGroupResponse, error) {
	rsp, err := s.ConnectRPC.CreateResourceMappingGroup(ctx, &connect.Request[resourcemapping.CreateResourceMappingGroupRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) UpdateResourceMappingGroup(ctx context.Context, req *resourcemapping.UpdateResourceMappingGroupRequest) (*resourcemapping.UpdateResourceMappingGroupResponse, error) {
	rsp, err := s.ConnectRPC.UpdateResourceMappingGroup(ctx, &connect.Request[resourcemapping.UpdateResourceMappingGroupRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) DeleteResourceMappingGroup(ctx context.Context, req *resourcemapping.DeleteResourceMappingGroupRequest) (*resourcemapping.DeleteResourceMappingGroupResponse, error) {
	rsp, err := s.ConnectRPC.DeleteResourceMappingGroup(ctx, &connect.Request[resourcemapping.DeleteResourceMappingGroupRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

/*
	Resource Mappings
*/

func (s ResourceMappingServiceGRPCGateway) CreateResourceMapping(ctx context.Context, req *resourcemapping.CreateResourceMappingRequest) (*resourcemapping.CreateResourceMappingResponse, error) {
	rsp, err := s.ConnectRPC.CreateResourceMapping(ctx, &connect.Request[resourcemapping.CreateResourceMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) ListResourceMappings(ctx context.Context, req *resourcemapping.ListResourceMappingsRequest) (*resourcemapping.ListResourceMappingsResponse, error) {
	rsp, err := s.ConnectRPC.ListResourceMappings(ctx, &connect.Request[resourcemapping.ListResourceMappingsRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) ListResourceMappingsByGroupFqns(ctx context.Context, req *resourcemapping.ListResourceMappingsByGroupFqnsRequest) (*resourcemapping.ListResourceMappingsByGroupFqnsResponse, error) {
	rsp, err := s.ConnectRPC.ListResourceMappingsByGroupFqns(ctx, &connect.Request[resourcemapping.ListResourceMappingsByGroupFqnsRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) GetResourceMapping(ctx context.Context, req *resourcemapping.GetResourceMappingRequest) (*resourcemapping.GetResourceMappingResponse, error) {
	rsp, err := s.ConnectRPC.GetResourceMapping(ctx, &connect.Request[resourcemapping.GetResourceMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) UpdateResourceMapping(ctx context.Context, req *resourcemapping.UpdateResourceMappingRequest) (*resourcemapping.UpdateResourceMappingResponse, error) {
	rsp, err := s.ConnectRPC.UpdateResourceMapping(ctx, &connect.Request[resourcemapping.UpdateResourceMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s ResourceMappingServiceGRPCGateway) DeleteResourceMapping(ctx context.Context, req *resourcemapping.DeleteResourceMappingRequest) (*resourcemapping.DeleteResourceMappingResponse, error) {
	rsp, err := s.ConnectRPC.DeleteResourceMapping(ctx, &connect.Request[resourcemapping.DeleteResourceMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}
