package attributes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentdf/platform/protocol/go/policy/attributes"
	"github.com/opentdf/platform/protocol/go/policy/attributes/attributesconnect"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/logger/audit"
	"github.com/opentdf/platform/service/pkg/db"
	"github.com/opentdf/platform/service/pkg/serviceregistry"
	policydb "github.com/opentdf/platform/service/policy/db"
)

type AttributesService struct { //nolint:revive // AttributesService is a valid name for this struct
	attributes.UnimplementedAttributesServiceServer
	dbClient policydb.PolicyDBClient
	logger   *logger.Logger
}

func NewRegistration() serviceregistry.Registration {
	return serviceregistry.Registration{
		ServiceDesc: &attributes.AttributesService_ServiceDesc,
		RegisterFunc: func(srp serviceregistry.RegistrationParams) (any, serviceregistry.HandlerServer) {
			attrSvc := &AttributesService{dbClient: policydb.NewClient(srp.DBClient, srp.Logger), logger: srp.Logger}

			grpcGateway := &AttributesServiceGRPCGateway{
				ConnectRPC: *attrSvc,
			}

			return grpcGateway, func(ctx context.Context, connectRPC *http.ServeMux, mux *runtime.ServeMux, server any) error {
				if srv, ok := server.(attributes.AttributesServiceServer); ok {
					path, attrConnect := attributesconnect.NewAttributesServiceHandler(attrSvc)
					connectRPC.Handle(path, attrConnect)
					return attributes.RegisterAttributesServiceHandlerServer(ctx, mux, srv)
				}
				return fmt.Errorf("failed to assert server as attributes.AttributesServiceServer")
			}
		},
	}
}

func (s AttributesService) CreateAttribute(ctx context.Context,
	req *connect.Request[attributes.CreateAttributeRequest],
) (*connect.Response[attributes.CreateAttributeResponse], error) {
	s.logger.Debug("creating new attribute definition", slog.String("name", req.Msg.GetName()))
	rsp := &attributes.CreateAttributeResponse{}

	auditParams := audit.PolicyEventParams{
		ObjectType: audit.ObjectTypeAttributeDefinition,
		ActionType: audit.ActionTypeCreate,
	}

	item, err := s.dbClient.CreateAttribute(ctx, req.Msg)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextCreationFailed, slog.String("attribute", req.Msg.String()))
	}

	s.logger.Debug("created new attribute definition", slog.String("name", req.Msg.GetName()))

	auditParams.ObjectID = item.GetId()
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	rsp.Attribute = item
	return &connect.Response[attributes.CreateAttributeResponse]{Msg: rsp}, nil
}

func (s *AttributesService) ListAttributes(ctx context.Context,
	req *connect.Request[attributes.ListAttributesRequest],
) (*connect.Response[attributes.ListAttributesResponse], error) {
	state := policydb.GetDBStateTypeTransformedEnum(req.Msg.GetState())
	namespace := req.Msg.GetNamespace()
	s.logger.Debug("listing attribute definitions", slog.String("state", state))
	rsp := &attributes.ListAttributesResponse{}

	list, err := s.dbClient.ListAllAttributes(ctx, state, namespace)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed)
	}
	rsp.Attributes = list

	return &connect.Response[attributes.ListAttributesResponse]{Msg: rsp}, nil
}

func (s *AttributesService) GetAttribute(ctx context.Context,
	req *connect.Request[attributes.GetAttributeRequest],
) (*connect.Response[attributes.GetAttributeResponse], error) {
	rsp := &attributes.GetAttributeResponse{}

	item, err := s.dbClient.GetAttribute(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}
	rsp.Attribute = item

	return &connect.Response[attributes.GetAttributeResponse]{Msg: rsp}, err
}

func (s *AttributesService) GetAttributeValuesByFqns(ctx context.Context,
	req *connect.Request[attributes.GetAttributeValuesByFqnsRequest],
) (*connect.Response[attributes.GetAttributeValuesByFqnsResponse], error) {
	rsp := &attributes.GetAttributeValuesByFqnsResponse{}

	fqnsToAttributes, err := s.dbClient.GetAttributesByValueFqns(ctx, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("fqns", fmt.Sprintf("%v", req.Msg.GetFqns())))
	}
	rsp.FqnAttributeValues = fqnsToAttributes

	return &connect.Response[attributes.GetAttributeValuesByFqnsResponse]{Msg: rsp}, nil
}

func (s *AttributesService) UpdateAttribute(ctx context.Context,
	req *connect.Request[attributes.UpdateAttributeRequest],
) (*connect.Response[attributes.UpdateAttributeResponse], error) {
	rsp := &attributes.UpdateAttributeResponse{}

	attributeID := req.Msg.GetId()
	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeUpdate,
		ObjectType: audit.ObjectTypeAttributeDefinition,
		ObjectID:   attributeID,
	}

	original, err := s.dbClient.GetAttribute(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", attributeID))
	}

	item, err := s.dbClient.UpdateAttribute(ctx, req.Msg.GetId(), req.Msg)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()), slog.String("attribute", req.Msg.String()))
	}

	// Item above only contains the attribute ID so we need to get the full
	// attribute definition to compute the diff.
	updated, err := s.dbClient.GetAttribute(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", attributeID))
	}

	auditParams.Original = original
	auditParams.Updated = updated
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	rsp.Attribute = item
	return &connect.Response[attributes.UpdateAttributeResponse]{Msg: rsp}, nil
}

func (s *AttributesService) DeactivateAttribute(ctx context.Context,
	req *connect.Request[attributes.DeactivateAttributeRequest],
) (*connect.Response[attributes.DeactivateAttributeResponse], error) {
	rsp := &attributes.DeactivateAttributeResponse{}

	attributeID := req.Msg.GetId()
	auditParams := audit.PolicyEventParams{
		ObjectType: audit.ObjectTypeAttributeDefinition,
		ActionType: audit.ActionTypeUpdate,
		ObjectID:   attributeID,
	}

	originalAttribute, err := s.dbClient.GetAttribute(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", attributeID))
	}

	// DeactivateAttribute actually returns the entire attribute so we can use it
	// to compute the diff.
	deactivatedAttribute, err := s.dbClient.DeactivateAttribute(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextDeactivationFailed, slog.String("id", attributeID))
	}

	auditParams.Original = originalAttribute
	auditParams.Updated = deactivatedAttribute
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	rsp.Attribute = deactivatedAttribute
	return &connect.Response[attributes.DeactivateAttributeResponse]{Msg: rsp}, nil
}

///
/// Attribute Values
///

func (s *AttributesService) CreateAttributeValue(ctx context.Context, req *connect.Request[attributes.CreateAttributeValueRequest]) (*connect.Response[attributes.CreateAttributeValueResponse], error) {
	auditParams := audit.PolicyEventParams{
		ObjectType: audit.ObjectTypeAttributeValue,
		ActionType: audit.ActionTypeCreate,
	}

	item, err := s.dbClient.CreateAttributeValue(ctx, req.Msg.GetAttributeId(), req.Msg)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextCreationFailed, slog.String("attributeId", req.Msg.GetAttributeId()), slog.String("value", req.Msg.String()))
	}

	auditParams.ObjectID = item.GetId()
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[attributes.CreateAttributeValueResponse]{Msg: &attributes.CreateAttributeValueResponse{
		Value: item,
	}}, nil
}

func (s *AttributesService) ListAttributeValues(ctx context.Context, req *connect.Request[attributes.ListAttributeValuesRequest]) (*connect.Response[attributes.ListAttributeValuesResponse], error) {
	state := policydb.GetDBStateTypeTransformedEnum(req.Msg.GetState())
	s.logger.Debug("listing attribute values", slog.String("attributeId", req.Msg.GetAttributeId()), slog.String("state", state))
	list, err := s.dbClient.ListAttributeValues(ctx, req.Msg.GetAttributeId(), state)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed, slog.String("attributeId", req.Msg.GetAttributeId()))
	}

	return &connect.Response[attributes.ListAttributeValuesResponse]{Msg: &attributes.ListAttributeValuesResponse{
		Values: list,
	}}, nil
}

func (s *AttributesService) GetAttributeValue(ctx context.Context, req *connect.Request[attributes.GetAttributeValueRequest]) (*connect.Response[attributes.GetAttributeValueResponse], error) {
	item, err := s.dbClient.GetAttributeValue(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	return &connect.Response[attributes.GetAttributeValueResponse]{Msg: &attributes.GetAttributeValueResponse{
		Value: item,
	}}, nil
}

func (s *AttributesService) UpdateAttributeValue(ctx context.Context, req *connect.Request[attributes.UpdateAttributeValueRequest]) (*connect.Response[attributes.UpdateAttributeValueResponse], error) {
	attributeID := req.Msg.GetId()
	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeUpdate,
		ObjectType: audit.ObjectTypeAttributeValue,
		ObjectID:   attributeID,
	}

	original, err := s.dbClient.GetAttributeValue(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", attributeID))
	}

	item, err := s.dbClient.UpdateAttributeValue(ctx, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()), slog.String("value", req.Msg.String()))
	}

	// UpdateAttributeValue only returns the attribute ID so we need to get the
	// full attribute value to compute the diff.
	updated, err := s.dbClient.GetAttributeValue(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", attributeID))
	}

	auditParams.Original = original
	auditParams.Updated = updated
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[attributes.UpdateAttributeValueResponse]{Msg: &attributes.UpdateAttributeValueResponse{
		Value: item,
	}}, nil
}

func (s *AttributesService) DeactivateAttributeValue(ctx context.Context, req *connect.Request[attributes.DeactivateAttributeValueRequest]) (*connect.Response[attributes.DeactivateAttributeValueResponse], error) {
	attributeID := req.Msg.GetId()
	auditParams := audit.PolicyEventParams{
		ObjectType: audit.ObjectTypeAttributeValue,
		ActionType: audit.ActionTypeDelete,
		ObjectID:   attributeID,
	}

	original, err := s.dbClient.GetAttributeValue(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", attributeID))
	}

	// DeactivateAttributeValue actually returns the entire attribute value so we
	// can use it to compute the diff.
	deactivated, err := s.dbClient.DeactivateAttributeValue(ctx, attributeID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextDeactivationFailed, slog.String("id", attributeID))
	}

	auditParams.Original = original
	auditParams.Updated = deactivated
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[attributes.DeactivateAttributeValueResponse]{Msg: &attributes.DeactivateAttributeValueResponse{
		Value: deactivated,
	}}, nil
}

func (s *AttributesService) AssignKeyAccessServerToAttribute(ctx context.Context, req *connect.Request[attributes.AssignKeyAccessServerToAttributeRequest]) (*connect.Response[attributes.AssignKeyAccessServerToAttributeResponse], error) {
	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeCreate,
		ObjectType: audit.ObjectTypeKasAttributeDefinitionAssignment,
		ObjectID:   fmt.Sprintf("%s-%s", req.Msg.GetAttributeKeyAccessServer().GetAttributeId(), req.Msg.GetAttributeKeyAccessServer().GetKeyAccessServerId()),
	}

	attributeKas, err := s.dbClient.AssignKeyAccessServerToAttribute(ctx, req.Msg.GetAttributeKeyAccessServer())
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextCreationFailed, slog.String("attributeKas", req.Msg.GetAttributeKeyAccessServer().String()))
	}
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[attributes.AssignKeyAccessServerToAttributeResponse]{Msg: &attributes.AssignKeyAccessServerToAttributeResponse{
		AttributeKeyAccessServer: attributeKas,
	}}, nil
}

func (s *AttributesService) RemoveKeyAccessServerFromAttribute(ctx context.Context, req *connect.Request[attributes.RemoveKeyAccessServerFromAttributeRequest]) (*connect.Response[attributes.RemoveKeyAccessServerFromAttributeResponse], error) {
	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeDelete,
		ObjectType: audit.ObjectTypeKasAttributeDefinitionAssignment,
		ObjectID:   fmt.Sprintf("%s-%s", req.Msg.GetAttributeKeyAccessServer().GetAttributeId(), req.Msg.GetAttributeKeyAccessServer().GetKeyAccessServerId()),
	}

	attributeKas, err := s.dbClient.RemoveKeyAccessServerFromAttribute(ctx, req.Msg.GetAttributeKeyAccessServer())
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("attributeKas", req.Msg.GetAttributeKeyAccessServer().String()))
	}
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[attributes.RemoveKeyAccessServerFromAttributeResponse]{Msg: &attributes.RemoveKeyAccessServerFromAttributeResponse{
		AttributeKeyAccessServer: attributeKas,
	}}, nil
}

func (s *AttributesService) AssignKeyAccessServerToValue(ctx context.Context, req *connect.Request[attributes.AssignKeyAccessServerToValueRequest]) (*connect.Response[attributes.AssignKeyAccessServerToValueResponse], error) {
	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeCreate,
		ObjectType: audit.ObjectTypeKasAttributeValueAssignment,
		ObjectID:   fmt.Sprintf("%s-%s", req.Msg.GetValueKeyAccessServer().GetValueId(), req.Msg.GetValueKeyAccessServer().GetKeyAccessServerId()),
	}

	valueKas, err := s.dbClient.AssignKeyAccessServerToValue(ctx, req.Msg.GetValueKeyAccessServer())
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextCreationFailed, slog.String("attributeValueKas", req.Msg.GetValueKeyAccessServer().String()))
	}
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[attributes.AssignKeyAccessServerToValueResponse]{Msg: &attributes.AssignKeyAccessServerToValueResponse{
		ValueKeyAccessServer: valueKas,
	}}, nil
}

func (s *AttributesService) RemoveKeyAccessServerFromValue(ctx context.Context, req *connect.Request[attributes.RemoveKeyAccessServerFromValueRequest]) (*connect.Response[attributes.RemoveKeyAccessServerFromValueResponse], error) {
	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeDelete,
		ObjectType: audit.ObjectTypeKasAttributeValueAssignment,
		ObjectID:   fmt.Sprintf("%s-%s", req.Msg.GetValueKeyAccessServer().GetValueId(), req.Msg.GetValueKeyAccessServer().GetKeyAccessServerId()),
	}

	valueKas, err := s.dbClient.RemoveKeyAccessServerFromValue(ctx, req.Msg.GetValueKeyAccessServer())
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("attributeValueKas", req.Msg.GetValueKeyAccessServer().String()))
	}
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[attributes.RemoveKeyAccessServerFromValueResponse]{Msg: &attributes.RemoveKeyAccessServerFromValueResponse{
		ValueKeyAccessServer: valueKas,
	}}, nil
}
