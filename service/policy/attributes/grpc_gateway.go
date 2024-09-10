package attributes

import (
	"context"

	"connectrpc.com/connect"
	"github.com/opentdf/platform/protocol/go/policy/attributes"
)

type AttributesServiceGRPCGateway struct { //nolint:revive // AttributesService is a valid name for this struct
	attributes.UnimplementedAttributesServiceServer
	ConnectRPC AttributesService
}

func (s AttributesServiceGRPCGateway) CreateAttribute(ctx context.Context, req *attributes.CreateAttributeRequest) (*attributes.CreateAttributeResponse, error) {
	rsp, err := s.ConnectRPC.CreateAttribute(ctx, &connect.Request[attributes.CreateAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) ListAttributes(ctx context.Context, req *attributes.ListAttributesRequest) (*attributes.ListAttributesResponse, error) {
	rsp, err := s.ConnectRPC.ListAttributes(ctx, &connect.Request[attributes.ListAttributesRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) GetAttribute(ctx context.Context, req *attributes.GetAttributeRequest) (*attributes.GetAttributeResponse, error) {
	rsp, err := s.ConnectRPC.GetAttribute(ctx, &connect.Request[attributes.GetAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) GetAttributeValuesByFqns(ctx context.Context, req *attributes.GetAttributeValuesByFqnsRequest) (*attributes.GetAttributeValuesByFqnsResponse, error) {
	rsp, err := s.ConnectRPC.GetAttributeValuesByFqns(ctx, &connect.Request[attributes.GetAttributeValuesByFqnsRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) UpdateAttribute(ctx context.Context, req *attributes.UpdateAttributeRequest) (*attributes.UpdateAttributeResponse, error) {
	rsp, err := s.ConnectRPC.UpdateAttribute(ctx, &connect.Request[attributes.UpdateAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) DeactivateAttribute(ctx context.Context, req *attributes.DeactivateAttributeRequest) (*attributes.DeactivateAttributeResponse, error) {
	rsp, err := s.ConnectRPC.DeactivateAttribute(ctx, &connect.Request[attributes.DeactivateAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

///
/// Attribute Values
///

func (s *AttributesServiceGRPCGateway) CreateAttributeValue(ctx context.Context, req *attributes.CreateAttributeValueRequest) (*attributes.CreateAttributeValueResponse, error) {
	rsp, err := s.ConnectRPC.CreateAttributeValue(ctx, &connect.Request[attributes.CreateAttributeValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) ListAttributeValues(ctx context.Context, req *attributes.ListAttributeValuesRequest) (*attributes.ListAttributeValuesResponse, error) {
	rsp, err := s.ConnectRPC.ListAttributeValues(ctx, &connect.Request[attributes.ListAttributeValuesRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) GetAttributeValue(ctx context.Context, req *attributes.GetAttributeValueRequest) (*attributes.GetAttributeValueResponse, error) {
	rsp, err := s.ConnectRPC.GetAttributeValue(ctx, &connect.Request[attributes.GetAttributeValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) UpdateAttributeValue(ctx context.Context, req *attributes.UpdateAttributeValueRequest) (*attributes.UpdateAttributeValueResponse, error) {
	rsp, err := s.ConnectRPC.UpdateAttributeValue(ctx, &connect.Request[attributes.UpdateAttributeValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) DeactivateAttributeValue(ctx context.Context, req *attributes.DeactivateAttributeValueRequest) (*attributes.DeactivateAttributeValueResponse, error) {
	rsp, err := s.ConnectRPC.DeactivateAttributeValue(ctx, &connect.Request[attributes.DeactivateAttributeValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) AssignKeyAccessServerToAttribute(ctx context.Context, req *attributes.AssignKeyAccessServerToAttributeRequest) (*attributes.AssignKeyAccessServerToAttributeResponse, error) {
	rsp, err := s.ConnectRPC.AssignKeyAccessServerToAttribute(ctx, &connect.Request[attributes.AssignKeyAccessServerToAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) RemoveKeyAccessServerFromAttribute(ctx context.Context, req *attributes.RemoveKeyAccessServerFromAttributeRequest) (*attributes.RemoveKeyAccessServerFromAttributeResponse, error) {
	rsp, err := s.ConnectRPC.RemoveKeyAccessServerFromAttribute(ctx, &connect.Request[attributes.RemoveKeyAccessServerFromAttributeRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) AssignKeyAccessServerToValue(ctx context.Context, req *attributes.AssignKeyAccessServerToValueRequest) (*attributes.AssignKeyAccessServerToValueResponse, error) {
	rsp, err := s.ConnectRPC.AssignKeyAccessServerToValue(ctx, &connect.Request[attributes.AssignKeyAccessServerToValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s *AttributesServiceGRPCGateway) RemoveKeyAccessServerFromValue(ctx context.Context, req *attributes.RemoveKeyAccessServerFromValueRequest) (*attributes.RemoveKeyAccessServerFromValueResponse, error) {
	rsp, err := s.ConnectRPC.RemoveKeyAccessServerFromValue(ctx, &connect.Request[attributes.RemoveKeyAccessServerFromValueRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}
