package kas

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	kaspb "github.com/opentdf/platform/protocol/go/kas"
	"github.com/opentdf/platform/service/kas/access"
	"github.com/opentdf/platform/service/pkg/serviceregistry"
)

func NewRegistration() serviceregistry.Registration {
	return serviceregistry.Registration{
		Namespace:   "kas",
		ServiceDesc: &kaspb.AccessService_ServiceDesc,
		RegisterFunc: func(srp serviceregistry.RegistrationParams) (any, serviceregistry.HandlerServer) {
			hsm := srp.OTDF.HSM
			if hsm == nil {
				slog.Error("hsm not enabled")
				panic(fmt.Errorf("hsm not enabled"))
			}
			// FIXME msg="mismatched key access url" keyAccessURL=http://localhost:9000 kasURL=https://:9000
			kasURLString := "https://" + srp.OTDF.HTTPServer.Addr
			kasURI, err := url.Parse(kasURLString)
			if err != nil {
				panic(fmt.Errorf("invalid kas address [%s] %w", kasURLString, err))
			}

			p := access.Provider{
				URI:     *kasURI,
				SDK:     srp.SDK,
				Session: *hsm,
			}
			return &p, func(ctx context.Context, mux *runtime.ServeMux, server any) error {
				kas, ok := server.(*access.Provider)
				if !ok {
					panic("invalid kas server object")
				}
				return kaspb.RegisterAccessServiceHandlerServer(ctx, mux, kas)
			}
		},
	}
}
