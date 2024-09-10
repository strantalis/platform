package kasregistry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	kasr "github.com/opentdf/platform/protocol/go/policy/kasregistry"
	"github.com/opentdf/platform/protocol/go/policy/kasregistry/kasregistryconnect"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/logger/audit"
	"github.com/opentdf/platform/service/pkg/db"
	"github.com/opentdf/platform/service/pkg/serviceregistry"
	policydb "github.com/opentdf/platform/service/policy/db"
)

type KeyAccessServerRegistry struct {
	dbClient policydb.PolicyDBClient
	logger   *logger.Logger
}

func NewRegistration() serviceregistry.Registration {
	return serviceregistry.Registration{
		ServiceDesc: &kasr.KeyAccessServerRegistryService_ServiceDesc,
		RegisterFunc: func(srp serviceregistry.RegistrationParams) (any, serviceregistry.HandlerServer) {
			svc := &KeyAccessServerRegistry{dbClient: policydb.NewClient(srp.DBClient, srp.Logger), logger: srp.Logger}

			grpcGateway := &KeyAccessServerRegistryGRPCGateway{
				ConnectRPC: *svc,
			}

			return grpcGateway, func(ctx context.Context, connectRPC *http.ServeMux, mux *runtime.ServeMux, s any) error {
				srv, ok := s.(kasr.KeyAccessServerRegistryServiceServer)
				if !ok {
					return fmt.Errorf("argument is not of type kasr.KeyAccessServerRegistryServiceServer")
				}

				path, kasRegistryConnect := kasregistryconnect.NewKeyAccessServerRegistryServiceHandler(svc)
				connectRPC.Handle(path, kasRegistryConnect)

				return kasr.RegisterKeyAccessServerRegistryServiceHandlerServer(ctx, mux, srv)
			}
		},
	}
}

func (s KeyAccessServerRegistry) CreateKeyAccessServer(ctx context.Context,
	req *connect.Request[kasr.CreateKeyAccessServerRequest],
) (*connect.Response[kasr.CreateKeyAccessServerResponse], error) {
	s.logger.Debug("creating key access server")

	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeCreate,
		ObjectType: audit.ObjectTypeKasRegistry,
	}

	ks, err := s.dbClient.CreateKeyAccessServer(ctx, req.Msg)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextCreationFailed, slog.String("keyAccessServer", req.Msg.String()))
	}

	auditParams.ObjectID = ks.GetId()
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[kasr.CreateKeyAccessServerResponse]{Msg: &kasr.CreateKeyAccessServerResponse{
		KeyAccessServer: ks,
	}}, nil
}

func (s KeyAccessServerRegistry) ListKeyAccessServers(ctx context.Context,
	_ *connect.Request[kasr.ListKeyAccessServersRequest],
) (*connect.Response[kasr.ListKeyAccessServersResponse], error) {
	keyAccessServers, err := s.dbClient.ListKeyAccessServers(ctx)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed)
	}

	return &connect.Response[kasr.ListKeyAccessServersResponse]{Msg: &kasr.ListKeyAccessServersResponse{
		KeyAccessServers: keyAccessServers,
	}}, nil
}

func (s KeyAccessServerRegistry) GetKeyAccessServer(ctx context.Context,
	req *connect.Request[kasr.GetKeyAccessServerRequest],
) (*connect.Response[kasr.GetKeyAccessServerResponse], error) {
	keyAccessServer, err := s.dbClient.GetKeyAccessServer(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	return &connect.Response[kasr.GetKeyAccessServerResponse]{Msg: &kasr.GetKeyAccessServerResponse{
		KeyAccessServer: keyAccessServer,
	}}, nil
}

func (s KeyAccessServerRegistry) UpdateKeyAccessServer(ctx context.Context,
	req *connect.Request[kasr.UpdateKeyAccessServerRequest],
) (*connect.Response[kasr.UpdateKeyAccessServerResponse], error) {
	kasID := req.Msg.GetId()

	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeUpdate,
		ObjectType: audit.ObjectTypeKasRegistry,
		ObjectID:   kasID,
	}

	originalKAS, err := s.dbClient.GetKeyAccessServer(ctx, kasID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", kasID))
	}

	item, err := s.dbClient.UpdateKeyAccessServer(ctx, kasID, req.Msg)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()), slog.String("keyAccessServer", req.Msg.String()))
	}

	// UpdateKeyAccessServer only returns the ID of the updated KAS, so we need to
	// fetch the updated KAS to compute the audit diff
	updatedKAS, err := s.dbClient.GetKeyAccessServer(ctx, kasID)
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", kasID))
	}

	auditParams.Original = originalKAS
	auditParams.Updated = updatedKAS
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)

	return &connect.Response[kasr.UpdateKeyAccessServerResponse]{Msg: &kasr.UpdateKeyAccessServerResponse{
		KeyAccessServer: item,
	}}, nil
}

func (s KeyAccessServerRegistry) DeleteKeyAccessServer(ctx context.Context,
	req *connect.Request[kasr.DeleteKeyAccessServerRequest],
) (*connect.Response[kasr.DeleteKeyAccessServerResponse], error) {
	kasID := req.Msg.GetId()
	auditParams := audit.PolicyEventParams{
		ActionType: audit.ActionTypeDelete,
		ObjectType: audit.ObjectTypeKasRegistry,
		ObjectID:   kasID,
	}

	keyAccessServer, err := s.dbClient.DeleteKeyAccessServer(ctx, req.Msg.GetId())
	if err != nil {
		s.logger.Audit.PolicyCRUDFailure(ctx, auditParams)
		return nil, db.StatusifyError(err, db.ErrTextDeletionFailed, slog.String("id", req.Msg.GetId()))
	}
	s.logger.Audit.PolicyCRUDSuccess(ctx, auditParams)
	return &connect.Response[kasr.DeleteKeyAccessServerResponse]{Msg: &kasr.DeleteKeyAccessServerResponse{
		KeyAccessServer: keyAccessServer,
	}}, nil
}

func (s KeyAccessServerRegistry) ListKeyAccessServerGrants(ctx context.Context,
	req *connect.Request[kasr.ListKeyAccessServerGrantsRequest],
) (*connect.Response[kasr.ListKeyAccessServerGrantsResponse], error) {
	keyAccessServerGrants, err := s.dbClient.ListKeyAccessServerGrants(ctx, req.Msg.GetKasId(), req.Msg.GetKasUri())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextListRetrievalFailed)
	}

	return &connect.Response[kasr.ListKeyAccessServerGrantsResponse]{Msg: &kasr.ListKeyAccessServerGrantsResponse{
		Grants: keyAccessServerGrants,
	}}, nil
}
