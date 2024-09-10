package unsafe

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentdf/platform/protocol/go/policy/unsafe"
	"github.com/opentdf/platform/protocol/go/policy/unsafe/unsafeconnect"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/pkg/db"
	"github.com/opentdf/platform/service/pkg/serviceregistry"
	policydb "github.com/opentdf/platform/service/policy/db"
)

type UnsafeService struct { //nolint:revive // UnsafeService is a valid name for this struct
	dbClient policydb.PolicyDBClient
	logger   *logger.Logger
}

func NewRegistration() serviceregistry.Registration {
	return serviceregistry.Registration{
		ServiceDesc: &unsafe.UnsafeService_ServiceDesc,
		RegisterFunc: func(srp serviceregistry.RegistrationParams) (any, serviceregistry.HandlerServer) {
			svc := &UnsafeService{dbClient: policydb.NewClient(srp.DBClient, srp.Logger), logger: srp.Logger}

			grpcGateway := &UnsafeServiceGRPCGateway{
				ConnectRPC: *svc,
			}

			return grpcGateway, func(ctx context.Context, connectRPC *http.ServeMux, mux *runtime.ServeMux, server any) error {
				if srv, ok := server.(unsafe.UnsafeServiceServer); ok {
					path, unsafeConnect := unsafeconnect.NewUnsafeServiceHandler(svc)
					connectRPC.Handle(path, unsafeConnect)
					return unsafe.RegisterUnsafeServiceHandlerServer(ctx, mux, srv)
				}

				return fmt.Errorf("failed to assert server as unsafe.UnsafeServiceServer")
			}
		},
	}
}

//
// Unsafe Namespace RPCs
//

func (s *UnsafeService) UnsafeUpdateNamespace(ctx context.Context, req *connect.Request[unsafe.UnsafeUpdateNamespaceRequest]) (*connect.Response[unsafe.UnsafeUpdateNamespaceResponse], error) {
	rsp := &unsafe.UnsafeUpdateNamespaceResponse{}

	_, err := s.dbClient.GetNamespace(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	item, err := s.dbClient.UnsafeUpdateNamespace(ctx, req.Msg.GetId(), req.Msg.GetName())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()), slog.String("namespace", req.Msg.GetName()))
	}
	rsp.Namespace = item

	return &connect.Response[unsafe.UnsafeUpdateNamespaceResponse]{Msg: rsp}, nil
}

func (s *UnsafeService) UnsafeReactivateNamespace(ctx context.Context, req *connect.Request[unsafe.UnsafeReactivateNamespaceRequest]) (*connect.Response[unsafe.UnsafeReactivateNamespaceResponse], error) {
	rsp := &unsafe.UnsafeReactivateNamespaceResponse{}

	_, err := s.dbClient.GetNamespace(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	item, err := s.dbClient.UnsafeReactivateNamespace(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()))
	}
	rsp.Namespace = item

	return &connect.Response[unsafe.UnsafeReactivateNamespaceResponse]{Msg: rsp}, nil
}

func (s *UnsafeService) UnsafeDeleteNamespace(ctx context.Context, req *connect.Request[unsafe.UnsafeDeleteNamespaceRequest]) (*connect.Response[unsafe.UnsafeDeleteNamespaceResponse], error) {
	rsp := &unsafe.UnsafeDeleteNamespaceResponse{}

	existing, err := s.dbClient.GetNamespace(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	deleted, err := s.dbClient.UnsafeDeleteNamespace(ctx, existing, req.Msg.GetFqn())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextDeletionFailed, slog.String("id", req.Msg.GetId()))
	}

	rsp.Namespace = deleted

	return &connect.Response[unsafe.UnsafeDeleteNamespaceResponse]{Msg: rsp}, nil
}

//
// Unsafe Attribute Definition RPCs
//

func (s *UnsafeService) UnsafeUpdateAttribute(ctx context.Context, req *connect.Request[unsafe.UnsafeUpdateAttributeRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeResponse], error) {
	rsp := &unsafe.UnsafeUpdateAttributeResponse{}

	_, err := s.dbClient.GetAttribute(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	item, err := s.dbClient.UnsafeUpdateAttribute(ctx, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()), slog.String("attribute", req.Msg.String()))
	}

	rsp.Attribute = item

	return &connect.Response[unsafe.UnsafeUpdateAttributeResponse]{Msg: rsp}, nil
}

func (s *UnsafeService) UnsafeReactivateAttribute(ctx context.Context, req *connect.Request[unsafe.UnsafeReactivateAttributeRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeResponse], error) {
	rsp := &unsafe.UnsafeReactivateAttributeResponse{}

	_, err := s.dbClient.GetAttribute(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	item, err := s.dbClient.UnsafeReactivateAttribute(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()))
	}

	rsp.Attribute = item

	return &connect.Response[unsafe.UnsafeReactivateAttributeResponse]{Msg: rsp}, nil
}

func (s *UnsafeService) UnsafeDeleteAttribute(ctx context.Context, req *connect.Request[unsafe.UnsafeDeleteAttributeRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeResponse], error) {
	rsp := &unsafe.UnsafeDeleteAttributeResponse{}

	existing, err := s.dbClient.GetAttribute(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	deleted, err := s.dbClient.UnsafeDeleteAttribute(ctx, existing, req.Msg.GetFqn())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextDeletionFailed, slog.String("id", req.Msg.GetId()))
	}

	rsp.Attribute = deleted

	return &connect.Response[unsafe.UnsafeDeleteAttributeResponse]{Msg: rsp}, nil
}

//
// Unsafe Attribute Value RPCs
//

func (s *UnsafeService) UnsafeUpdateAttributeValue(ctx context.Context, req *connect.Request[unsafe.UnsafeUpdateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeUpdateAttributeValueResponse], error) {
	rsp := &unsafe.UnsafeUpdateAttributeValueResponse{}
	_, err := s.dbClient.GetAttributeValue(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	item, err := s.dbClient.UnsafeUpdateAttributeValue(ctx, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()), slog.String("attribute_value", req.Msg.String()))
	}

	rsp.Value = item
	return &connect.Response[unsafe.UnsafeUpdateAttributeValueResponse]{Msg: rsp}, nil
}

func (s *UnsafeService) UnsafeReactivateAttributeValue(ctx context.Context, req *connect.Request[unsafe.UnsafeReactivateAttributeValueRequest]) (*connect.Response[unsafe.UnsafeReactivateAttributeValueResponse], error) {
	rsp := &unsafe.UnsafeReactivateAttributeValueResponse{}

	_, err := s.dbClient.GetAttributeValue(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	item, err := s.dbClient.UnsafeReactivateAttributeValue(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextUpdateFailed, slog.String("id", req.Msg.GetId()))
	}

	rsp.Value = item
	return &connect.Response[unsafe.UnsafeReactivateAttributeValueResponse]{Msg: rsp}, nil
}

func (s *UnsafeService) UnsafeDeleteAttributeValue(ctx context.Context, req *connect.Request[unsafe.UnsafeDeleteAttributeValueRequest]) (*connect.Response[unsafe.UnsafeDeleteAttributeValueResponse], error) {
	rsp := &unsafe.UnsafeDeleteAttributeValueResponse{}
	existing, err := s.dbClient.GetAttributeValue(ctx, req.Msg.GetId())
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextGetRetrievalFailed, slog.String("id", req.Msg.GetId()))
	}

	deleted, err := s.dbClient.UnsafeDeleteAttributeValue(ctx, existing, req.Msg)
	if err != nil {
		return nil, db.StatusifyError(err, db.ErrTextDeletionFailed, slog.String("id", req.Msg.GetId()))
	}

	rsp.Value = deleted
	return &connect.Response[unsafe.UnsafeDeleteAttributeValueResponse]{Msg: rsp}, nil
}
