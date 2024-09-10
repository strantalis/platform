package entityresolution

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mitchellh/mapstructure"
	"github.com/opentdf/platform/protocol/go/entityresolution"
	"github.com/opentdf/platform/protocol/go/entityresolution/entityresolutionconnect"
	keycloak "github.com/opentdf/platform/service/entityresolution/keycloak"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/pkg/serviceregistry"
)

type EntityResolutionServiceGRPCGateway struct {
	entityresolution.UnimplementedEntityResolutionServiceServer
	idpConfig  keycloak.KeycloakConfig
	ConnectRPC EntityResolutionService
}

type EntityResolutionService struct { //nolint:revive // allow for simple naming
	idpConfig keycloak.KeycloakConfig
	logger    *logger.Logger
}

func NewRegistration() serviceregistry.Registration {
	return serviceregistry.Registration{
		Namespace:   "entityresolution",
		ServiceDesc: &entityresolution.EntityResolutionService_ServiceDesc,
		RegisterFunc: func(srp serviceregistry.RegistrationParams) (any, serviceregistry.HandlerServer) {
			var inputIdpConfig keycloak.KeycloakConfig

			if err := mapstructure.Decode(srp.Config, &inputIdpConfig); err != nil {
				panic(err)
			}

			srp.Logger.Debug("entity_resolution configuration", "config", inputIdpConfig)
			es := &EntityResolutionService{idpConfig: inputIdpConfig, logger: srp.Logger}

			svc := &EntityResolutionServiceGRPCGateway{idpConfig: inputIdpConfig, ConnectRPC: *es}

			return svc, func(ctx context.Context, connectRPC *http.ServeMux, mux *runtime.ServeMux, server any) error {
				path, ersConnect := entityresolutionconnect.NewEntityResolutionServiceHandler(es)
				connectRPC.Handle(path, ersConnect)
				return entityresolution.RegisterEntityResolutionServiceHandlerServer(ctx, mux, server.(entityresolution.EntityResolutionServiceServer)) //nolint:forcetypeassert // allow type assert, following other services
			}
		},
	}
}

func (s EntityResolutionServiceGRPCGateway) ResolveEntities(ctx context.Context, req *entityresolution.ResolveEntitiesRequest) (*entityresolution.ResolveEntitiesResponse, error) {
	rsp, err := s.ConnectRPC.ResolveEntities(ctx, &connect.Request[entityresolution.ResolveEntitiesRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s EntityResolutionServiceGRPCGateway) CreateEntityChainFromJwt(ctx context.Context, req *entityresolution.CreateEntityChainFromJwtRequest) (*entityresolution.CreateEntityChainFromJwtResponse, error) {
	rsp, err := s.ConnectRPC.CreateEntityChainFromJwt(ctx, &connect.Request[entityresolution.CreateEntityChainFromJwtRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s EntityResolutionService) ResolveEntities(ctx context.Context, req *connect.Request[entityresolution.ResolveEntitiesRequest]) (*connect.Response[entityresolution.ResolveEntitiesResponse], error) {
	resp, err := keycloak.EntityResolution(ctx, req.Msg, s.idpConfig, s.logger)
	return &connect.Response[entityresolution.ResolveEntitiesResponse]{Msg: &resp}, err
}

func (s EntityResolutionService) CreateEntityChainFromJwt(ctx context.Context, req *connect.Request[entityresolution.CreateEntityChainFromJwtRequest]) (*connect.Response[entityresolution.CreateEntityChainFromJwtResponse], error) {
	resp, err := keycloak.CreateEntityChainFromJwt(ctx, req.Msg, s.idpConfig, s.logger)
	return &connect.Response[entityresolution.CreateEntityChainFromJwtResponse]{Msg: &resp}, err
}
