package kasregistry

import (
	"context"

	"connectrpc.com/connect"
	kasr "github.com/opentdf/platform/protocol/go/policy/kasregistry"
)

type KeyAccessServerRegistryGRPCGateway struct {
	kasr.UnimplementedKeyAccessServerRegistryServiceServer
	ConnectRPC KeyAccessServerRegistry
}

func (s KeyAccessServerRegistryGRPCGateway) CreateKeyAccessServer(ctx context.Context, req *kasr.CreateKeyAccessServerRequest) (*kasr.CreateKeyAccessServerResponse, error) {
	rsp, err := s.ConnectRPC.CreateKeyAccessServer(ctx, &connect.Request[kasr.CreateKeyAccessServerRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s KeyAccessServerRegistryGRPCGateway) ListKeyAccessServers(ctx context.Context, _ *kasr.ListKeyAccessServersRequest) (*kasr.ListKeyAccessServersResponse, error) {
	rsp, err := s.ConnectRPC.ListKeyAccessServers(ctx, &connect.Request[kasr.ListKeyAccessServersRequest]{Msg: &kasr.ListKeyAccessServersRequest{}})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s KeyAccessServerRegistryGRPCGateway) GetKeyAccessServer(ctx context.Context, req *kasr.GetKeyAccessServerRequest) (*kasr.GetKeyAccessServerResponse, error) {
	rsp, err := s.ConnectRPC.GetKeyAccessServer(ctx, &connect.Request[kasr.GetKeyAccessServerRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s KeyAccessServerRegistryGRPCGateway) UpdateKeyAccessServer(ctx context.Context, req *kasr.UpdateKeyAccessServerRequest) (*kasr.UpdateKeyAccessServerResponse, error) {
	rsp, err := s.ConnectRPC.UpdateKeyAccessServer(ctx, &connect.Request[kasr.UpdateKeyAccessServerRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s KeyAccessServerRegistryGRPCGateway) DeleteKeyAccessServer(ctx context.Context, req *kasr.DeleteKeyAccessServerRequest) (*kasr.DeleteKeyAccessServerResponse, error) {
	rsp, err := s.ConnectRPC.DeleteKeyAccessServer(ctx, &connect.Request[kasr.DeleteKeyAccessServerRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s KeyAccessServerRegistryGRPCGateway) ListKeyAccessServerGrants(ctx context.Context, req *kasr.ListKeyAccessServerGrantsRequest) (*kasr.ListKeyAccessServerGrantsResponse, error) {
	rsp, err := s.ConnectRPC.ListKeyAccessServerGrants(ctx, &connect.Request[kasr.ListKeyAccessServerGrantsRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}
