package namespaces

import (
	"context"

	"connectrpc.com/connect"
	"github.com/opentdf/platform/protocol/go/policy/namespaces"
)

type NamespacesServiceGRPCGateway struct { //nolint:revive // NamespacesService is a valid name
	namespaces.UnimplementedNamespaceServiceServer
	ConnectRPC NamespacesService
}

func (ns NamespacesServiceGRPCGateway) ListNamespaces(ctx context.Context, req *namespaces.ListNamespacesRequest) (*namespaces.ListNamespacesResponse, error) {
	rsp, err := ns.ConnectRPC.ListNamespaces(ctx, &connect.Request[namespaces.ListNamespacesRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (ns NamespacesServiceGRPCGateway) GetNamespace(ctx context.Context, req *namespaces.GetNamespaceRequest) (*namespaces.GetNamespaceResponse, error) {
	rsp, err := ns.ConnectRPC.GetNamespace(ctx, &connect.Request[namespaces.GetNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (ns NamespacesServiceGRPCGateway) CreateNamespace(ctx context.Context, req *namespaces.CreateNamespaceRequest) (*namespaces.CreateNamespaceResponse, error) {
	rsp, err := ns.ConnectRPC.CreateNamespace(ctx, &connect.Request[namespaces.CreateNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (ns NamespacesServiceGRPCGateway) UpdateNamespace(ctx context.Context, req *namespaces.UpdateNamespaceRequest) (*namespaces.UpdateNamespaceResponse, error) {
	rsp, err := ns.ConnectRPC.UpdateNamespace(ctx, &connect.Request[namespaces.UpdateNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (ns NamespacesServiceGRPCGateway) DeactivateNamespace(ctx context.Context, req *namespaces.DeactivateNamespaceRequest) (*namespaces.DeactivateNamespaceResponse, error) {
	rsp, err := ns.ConnectRPC.DeactivateNamespace(ctx, &connect.Request[namespaces.DeactivateNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (ns NamespacesServiceGRPCGateway) AssignKeyAccessServerToNamespace(ctx context.Context, req *namespaces.AssignKeyAccessServerToNamespaceRequest) (*namespaces.AssignKeyAccessServerToNamespaceResponse, error) {
	rsp, err := ns.ConnectRPC.AssignKeyAccessServerToNamespace(ctx, &connect.Request[namespaces.AssignKeyAccessServerToNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (ns NamespacesServiceGRPCGateway) RemoveKeyAccessServerFromNamespace(ctx context.Context, req *namespaces.RemoveKeyAccessServerFromNamespaceRequest) (*namespaces.RemoveKeyAccessServerFromNamespaceResponse, error) {
	rsp, err := ns.ConnectRPC.RemoveKeyAccessServerFromNamespace(ctx, &connect.Request[namespaces.RemoveKeyAccessServerFromNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}
