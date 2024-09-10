package wellknownconfiguration

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	wellknown "github.com/opentdf/platform/protocol/go/wellknownconfiguration"
	"github.com/opentdf/platform/protocol/go/wellknownconfiguration/wellknownconfigurationconnect"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/pkg/serviceregistry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type WellKnownConfigGrpcGateway struct {
	wellknown.UnimplementedWellKnownServiceServer
	ConnectRPC WellKnownService
}

type WellKnownService struct {
	logger *logger.Logger
}

var (
	wellKnownConfiguration = make(map[string]any)
	rwMutex                sync.RWMutex
)

func RegisterConfiguration(namespace string, config any) error {
	rwMutex.Lock()
	if _, ok := wellKnownConfiguration[namespace]; ok {
		return fmt.Errorf("namespace %s configuration already registered", namespace)
	}
	wellKnownConfiguration[namespace] = config
	rwMutex.Unlock()
	return nil
}

func NewRegistration() serviceregistry.Registration {
	return serviceregistry.Registration{
		Namespace:   "wellknown",
		ServiceDesc: &wellknown.WellKnownService_ServiceDesc,
		RegisterFunc: func(srp serviceregistry.RegistrationParams) (any, serviceregistry.HandlerServer) {
			svc := &WellKnownService{logger: srp.Logger}

			grpcGateway := &WellKnownConfigGrpcGateway{ConnectRPC: *svc}

			return grpcGateway, func(ctx context.Context, connectRPC *http.ServeMux, mux *runtime.ServeMux, server any) error {
				if srv, ok := server.(wellknown.WellKnownServiceServer); ok {
					path, wellknownConnect := wellknownconfigurationconnect.NewWellKnownServiceHandler(svc)
					connectRPC.Handle(path, wellknownConnect)
					return wellknown.RegisterWellKnownServiceHandlerServer(ctx, mux, srv)
				}

				return fmt.Errorf("failed to assert server as WellKnownServiceServer")
			}
		},
	}
}

func (s WellKnownConfigGrpcGateway) GetWellKnownConfiguration(ctx context.Context, _ *wellknown.GetWellKnownConfigurationRequest) (*wellknown.GetWellKnownConfigurationResponse, error) {
	rsp, err := s.ConnectRPC.GetWellKnownConfiguration(ctx, &connect.Request[wellknown.GetWellKnownConfigurationRequest]{Msg: &wellknown.GetWellKnownConfigurationRequest{}})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s WellKnownService) GetWellKnownConfiguration(_ context.Context, _ *connect.Request[wellknown.GetWellKnownConfigurationRequest]) (*connect.Response[wellknown.GetWellKnownConfigurationResponse], error) {
	rwMutex.RLock()
	cfg, err := structpb.NewStruct(wellKnownConfiguration)
	rwMutex.RUnlock()
	if err != nil {
		s.logger.Error("failed to create struct for wellknown configuration", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to create struct for wellknown configuration")
	}

	return &connect.Response[wellknown.GetWellKnownConfigurationResponse]{Msg: &wellknown.GetWellKnownConfigurationResponse{
		Configuration: cfg,
	}}, nil
}
