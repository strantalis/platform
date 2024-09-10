package resourcemapping

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentdf/platform/protocol/go/policy/resourcemapping"
	"github.com/opentdf/platform/protocol/go/policy/resourcemapping/resourcemappingconnect"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/logger/audit"
	"github.com/opentdf/platform/service/pkg/db"
	"github.com/opentdf/platform/service/pkg/serviceregistry"
	policydb "github.com/opentdf/platform/service/policy/db"
)

type ResourceMappingService struct { //nolint:revive // ResourceMappingService is a valid name for this struct
	dbClient policydb.PolicyDBClient
	logger   *logger.Logger
}

func NewRegistration() serviceregistry.Registration {
	return serviceregistry.Registration{
		ServiceDesc: &resourcemapping.ResourceMappingService_ServiceDesc,
		RegisterFunc: func(srp serviceregistry.RegistrationParams) (any, serviceregistry.HandlerServer) {
			svc := &ResourceMappingService{dbClient: policydb.NewClient(srp.DBClient, srp.Logger), logger: srp.Logger}

			grpcGateway := &ResourceMappingServiceGRPCGateway{
				ConnectRPC: *svc,
			}

			return grpcGateway, func(ctx context.Context, connectRPC *http.ServeMux, mux *runtime.ServeMux, s any) error {
				server, ok := s.(resourcemapping.ResourceMappingServiceServer)
				if !ok {
					return fmt.Errorf("failed to assert server as resourcemapping.ResourceMappingServiceServer")
				}
				path, resourcemappingConnect := resourcemappingconnect.NewResourceMappingServiceHandler(svc)
				connectRPC.Handle(path, resourcemappingConnect)

				return resourcemapping.RegisterResourceMappingServiceHandlerServer(ctx, mux, server)
			}
		},
	}
}

/*
	Resource Mapping Groups
*/

func (s ResourceMappingService) ListResourceMappingGroups(ctx context.Context, req *connect.Request[resourcemapping.ListResourceMappingGroupsRequest]) (*connect.Response[resourcemapping.ListResourceMappingGroupsResponse], error) {
	rmGroups, err := s.dbClient.ListResourceMappingGroups(ctx, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed)
	}

	return &connect.Response[resourcemapping.ListResourceMappingGroupsResponse]{Msg: &resourcemapping.ListResourceMappingGroupsResponse{
		ResourceMappingGroups: rmGroups,
	}}, nil
}

func (s ResourceMappingService) GetResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.GetResourceMappingGroupRequest]) (*connect.Response[resourcemapping.GetResourceMappingGroupResponse], error) {
	rmGroup, err := s.dbClient.GetResourceMappingGroup(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	return &connect.Response[resourcemapping.GetResourceMappingGroupResponse]{Msg: &resourcemapping.GetResourceMappingGroupResponse{
		ResourceMappingGroup: rmGroup,
	}}, nil
}

func (s ResourceMappingService) CreateResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.CreateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.CreateResourceMappingGroupResponse], error) {
	rmGroup, err := s.dbClient.CreateResourceMappingGroup(ctx, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextCreationFailed, slog.String("resourceMappingGroup", req.Msg.String()))
	}

	return &connect.Response[resourcemapping.CreateResourceMappingGroupResponse]{Msg: &resourcemapping.CreateResourceMappingGroupResponse{
		ResourceMappingGroup: rmGroup,
	}}, nil
}

func (s ResourceMappingService) UpdateResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.UpdateResourceMappingGroupRequest]) (*connect.Response[resourcemapping.UpdateResourceMappingGroupResponse], error) {
	id := req.Msg.GetId()

	rmGroup, err := s.dbClient.UpdateResourceMappingGroup(ctx, id, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", id), slog.String("resourceMappingGroup", req.Msg.String()))
	}

	return &connect.Response[resourcemapping.UpdateResourceMappingGroupResponse]{Msg: &resourcemapping.UpdateResourceMappingGroupResponse{
		ResourceMappingGroup: rmGroup,
	}}, nil
}

func (s ResourceMappingService) DeleteResourceMappingGroup(ctx context.Context, req *connect.Request[resourcemapping.DeleteResourceMappingGroupRequest]) (*connect.Response[resourcemapping.DeleteResourceMappingGroupResponse], error) {
	id := req.Msg.GetId()

	rmGroup, err := s.dbClient.DeleteResourceMappingGroup(ctx, id)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextDeletionFailed, slog.String("id", id))
	}

	return &connect.Response[resourcemapping.DeleteResourceMappingGroupResponse]{Msg: &resourcemapping.DeleteResourceMappingGroupResponse{
		ResourceMappingGroup: rmGroup,
	}}, nil
}

/*
	Resource Mappings
*/

func (s ResourceMappingService) CreateResourceMapping(ctx context.Context,
	req *connect.Request[resourcemapping.CreateResourceMappingRequest],
) (*connect.Response[resourcemapping.CreateResourceMappingResponse], error) {
	s.logger.Debug("creating resource mapping")

	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeCreate,
		ObjectType: audit.ObjectTypeResourceMapping,
	}

	rm, err := s.dbClient.CreateResourceMapping(ctx, req.Msg)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextCreationFailed, slog.String("resourceMapping", req.Msg.String()))
	}

	auditParams.ObjectID = rm.GetId()
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[resourcemapping.CreateResourceMappingResponse]{Msg: &resourcemapping.CreateResourceMappingResponse{
		ResourceMapping: rm,
	}}, nil
}

func (s ResourceMappingService) ListResourceMappings(ctx context.Context,
	req *connect.Request[resourcemapping.ListResourceMappingsRequest],
) (*connect.Response[resourcemapping.ListResourceMappingsResponse], error) {
	resourceMappings, err := s.dbClient.ListResourceMappings(ctx, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed)
	}

	return &connect.Response[resourcemapping.ListResourceMappingsResponse]{Msg: &resourcemapping.ListResourceMappingsResponse{
		ResourceMappings: resourceMappings,
	}}, nil
}

func (s ResourceMappingService) ListResourceMappingsByGroupFqns(ctx context.Context, req *connect.Request[resourcemapping.ListResourceMappingsByGroupFqnsRequest]) (*connect.Response[resourcemapping.ListResourceMappingsByGroupFqnsResponse], error) {
	fqns := req.Msg.GetFqns()

	fqnRmGroupMap, err := s.dbClient.ListResourceMappingsByGroupFqns(ctx, fqns)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed, slog.Any("fqns", fqns))
	}

	return &connect.Response[resourcemapping.ListResourceMappingsByGroupFqnsResponse]{Msg: &resourcemapping.ListResourceMappingsByGroupFqnsResponse{
		FqnResourceMappingGroups: fqnRmGroupMap,
	}}, nil
}

func (s ResourceMappingService) GetResourceMapping(ctx context.Context,
	req *connect.Request[resourcemapping.GetResourceMappingRequest],
) (*connect.Response[resourcemapping.GetResourceMappingResponse], error) {
	rm, err := s.dbClient.GetResourceMapping(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	return &connect.Response[resourcemapping.GetResourceMappingResponse]{Msg: &resourcemapping.GetResourceMappingResponse{
		ResourceMapping: rm,
	}}, nil
}

func (s ResourceMappingService) UpdateResourceMapping(ctx context.Context,
	req *connect.Request[resourcemapping.UpdateResourceMappingRequest],
) (*connect.Response[resourcemapping.UpdateResourceMappingResponse], error) {
	resourceMappingID := req.Msg.GetId()

	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeUpdate,
		ObjectType: audit.ObjectTypeResourceMapping,
		ObjectID:   resourceMappingID,
	}

	originalRM, err := s.dbClient.GetResourceMapping(ctx, resourceMappingID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed)
	}

	item, err := s.dbClient.UpdateResourceMapping(
		ctx,
		resourceMappingID,
		req.Msg,
	)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()), slog.String("resourceMapping", req.Msg.String()))
	}

	// UpdateResourceMapping only returns the ID of the updated resource mapping
	// so we need to fetch the updated resource mapping to compute the audit diff
	updatedRM, err := s.dbClient.GetResourceMapping(ctx, resourceMappingID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed)
	}

	auditParams.Original = originalRM
	auditParams.Updated = updatedRM
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[resourcemapping.UpdateResourceMappingResponse]{Msg: &resourcemapping.UpdateResourceMappingResponse{
		ResourceMapping: item,
	}}, nil
}

func (s ResourceMappingService) DeleteResourceMapping(ctx context.Context,
	req *connect.Request[resourcemapping.DeleteResourceMappingRequest],
) (*connect.Response[resourcemapping.DeleteResourceMappingResponse], error) {
	resourceMappingID := req.Msg.GetId()

	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeDelete,
		ObjectType: audit.ObjectTypeResourceMapping,
		ObjectID:   resourceMappingID,
	}

	rm, err := s.dbClient.DeleteResourceMapping(ctx, resourceMappingID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextDeletionFailed, slog.String("id", resourceMappingID))
	}
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)
	return &connect.Response[resourcemapping.DeleteResourceMappingResponse]{Msg: &resourcemapping.DeleteResourceMappingResponse{
		ResourceMapping: rm,
	}}, nil
}
