package unsafe

import (
	"context"

	"connectrpc.com/connect"
	"github.com/opentdf/platform/protocol/go/policy/unsafe"
)

type UnsafeServiceGRPCGateway struct { //nolint:revive // UnsafeService is a valid name for this struct
	unsafe.UnimplementedUnsafeServiceServer
	ConnectRPC UnsafeService
}

//
// Unsafe Namespace RPCs
//

func (s UnsafeServiceGRPCGateway) UnsafeUpdateNamespace(ctx context.Context, req *unsafe.UnsafeUpdateNamespaceRequest) (*unsafe.UnsafeUpdateNamespaceResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeUpdateNamespace(ctx, &connect.Request[unsafe.UnsafeUpdateNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s UnsafeServiceGRPCGateway) UnsafeReactivateNamespace(ctx context.Context, req *unsafe.UnsafeReactivateNamespaceRequest) (*unsafe.UnsafeReactivateNamespaceResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeReactivateNamespace(ctx, &connect.Request[unsafe.UnsafeReactivateNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s UnsafeServiceGRPCGateway) UnsafeDeleteNamespace(ctx context.Context, req *unsafe.UnsafeDeleteNamespaceRequest) (*unsafe.UnsafeDeleteNamespaceResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeDeleteNamespace(ctx, &connect.Request[unsafe.UnsafeDeleteNamespaceRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

//
// Unsafe Attribute Definition RPCs
//

func (s UnsafeServiceGRPCGateway) UnsafeUpdateAttribute(ctx context.Context, req *unsafe.UnsafeUpdateAttributeRequest) (*unsafe.UnsafeUpdateAttributeResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeUpdateAttribute(ctx, &connect.Request[unsafe.UnsafeUpdateAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s UnsafeServiceGRPCGateway) UnsafeReactivateAttribute(ctx context.Context, req *unsafe.UnsafeReactivateAttributeRequest) (*unsafe.UnsafeReactivateAttributeResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeReactivateAttribute(ctx, &connect.Request[unsafe.UnsafeReactivateAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s UnsafeServiceGRPCGateway) UnsafeDeleteAttribute(ctx context.Context, req *unsafe.UnsafeDeleteAttributeRequest) (*unsafe.UnsafeDeleteAttributeResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeDeleteAttribute(ctx, &connect.Request[unsafe.UnsafeDeleteAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

//
// Unsafe Attribute Value RPCs
//

func (s UnsafeServiceGRPCGateway) UnsafeUpdateAttributeValue(ctx context.Context, req *unsafe.UnsafeUpdateAttributeValueRequest) (*unsafe.UnsafeUpdateAttributeValueResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeUpdateAttributeValue(ctx, &connect.Request[unsafe.UnsafeUpdateAttributeValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s UnsafeServiceGRPCGateway) UnsafeReactivateAttributeValue(ctx context.Context, req *unsafe.UnsafeReactivateAttributeValueRequest) (*unsafe.UnsafeReactivateAttributeValueResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeReactivateAttributeValue(ctx, &connect.Request[unsafe.UnsafeReactivateAttributeValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s UnsafeServiceGRPCGateway) UnsafeDeleteAttributeValue(ctx context.Context, req *unsafe.UnsafeDeleteAttributeValueRequest) (*unsafe.UnsafeDeleteAttributeValueResponse, error) {
	rsp, err := s.ConnectRPC.UnsafeDeleteAttributeValue(ctx, &connect.Request[unsafe.UnsafeDeleteAttributeValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}
