package subjectmapping

import (
	"context"

	"connectrpc.com/connect"
	sm "github.com/opentdf/platform/protocol/go/policy/subjectmapping"
)

type SubjectMappingServiceGRPCGateway struct { //nolint:revive // SubjectMappingService is a valid name for this struct
	sm.UnimplementedSubjectMappingServiceServer
	ConnectRPC SubjectMappingService
}

/* ---------------------------------------------------
 * ----------------- SubjectMappings -----------------
 * --------------------------------------------------*/

func (s SubjectMappingServiceGRPCGateway) CreateSubjectMapping(ctx context.Context, req *sm.CreateSubjectMappingRequest) (*sm.CreateSubjectMappingResponse, error) {
	rsp, err := s.ConnectRPC.CreateSubjectMapping(ctx, &connect.Request[sm.CreateSubjectMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) ListSubjectMappings(ctx context.Context, _ *sm.ListSubjectMappingsRequest) (*sm.ListSubjectMappingsResponse, error) {
	rsp, err := s.ConnectRPC.ListSubjectMappings(ctx, &connect.Request[sm.ListSubjectMappingsRequest]{Msg: &sm.ListSubjectMappingsRequest{}})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) GetSubjectMapping(ctx context.Context, req *sm.GetSubjectMappingRequest) (*sm.GetSubjectMappingResponse, error) {
	rsp, err := s.ConnectRPC.GetSubjectMapping(ctx, &connect.Request[sm.GetSubjectMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) UpdateSubjectMapping(ctx context.Context, req *sm.UpdateSubjectMappingRequest) (*sm.UpdateSubjectMappingResponse, error) {
	rsp, err := s.ConnectRPC.UpdateSubjectMapping(ctx, &connect.Request[sm.UpdateSubjectMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) DeleteSubjectMapping(ctx context.Context, req *sm.DeleteSubjectMappingRequest) (*sm.DeleteSubjectMappingResponse, error) {
	rsp, err := s.ConnectRPC.DeleteSubjectMapping(ctx, &connect.Request[sm.DeleteSubjectMappingRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) MatchSubjectMappings(ctx context.Context, req *sm.MatchSubjectMappingsRequest) (*sm.MatchSubjectMappingsResponse, error) {
	rsp, err := s.ConnectRPC.MatchSubjectMappings(ctx, &connect.Request[sm.MatchSubjectMappingsRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

/* --------------------------------------------------------
 * ----------------- SubjectConditionSets -----------------
 * -------------------------------------------------------*/

func (s SubjectMappingServiceGRPCGateway) GetSubjectConditionSet(ctx context.Context, req *sm.GetSubjectConditionSetRequest) (*sm.GetSubjectConditionSetResponse, error) {
	rsp, err := s.ConnectRPC.GetSubjectConditionSet(ctx, &connect.Request[sm.GetSubjectConditionSetRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) ListSubjectConditionSets(ctx context.Context, _ *sm.ListSubjectConditionSetsRequest) (*sm.ListSubjectConditionSetsResponse, error) {
	rsp, err := s.ConnectRPC.ListSubjectConditionSets(ctx, &connect.Request[sm.ListSubjectConditionSetsRequest]{Msg: &sm.ListSubjectConditionSetsRequest{}})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) CreateSubjectConditionSet(ctx context.Context, req *sm.CreateSubjectConditionSetRequest) (*sm.CreateSubjectConditionSetResponse, error) {
	rsp, err := s.ConnectRPC.CreateSubjectConditionSet(ctx, &connect.Request[sm.CreateSubjectConditionSetRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) UpdateSubjectConditionSet(ctx context.Context, req *sm.UpdateSubjectConditionSetRequest) (*sm.UpdateSubjectConditionSetResponse, error) {
	rsp, err := s.ConnectRPC.UpdateSubjectConditionSet(ctx, &connect.Request[sm.UpdateSubjectConditionSetRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}

func (s SubjectMappingServiceGRPCGateway) DeleteSubjectConditionSet(ctx context.Context, req *sm.DeleteSubjectConditionSetRequest) (*sm.DeleteSubjectConditionSetResponse, error) {
	rsp, err := s.ConnectRPC.DeleteSubjectConditionSet(ctx, &connect.Request[sm.DeleteSubjectConditionSetRequest]{Msg: req})
	if err != nil {
		return nil, err
	}
	return rsp.Msg, nil
}
