package service

import (
	"context"

	v1 "admin/api/admin/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminService) PermissionList(ctx context.Context, req *emptypb.Empty) (*v1.PermissionListReply, error) {
	return s.pu.List(ctx)
}

func (s *AdminService) RolePermissionList(ctx context.Context, req *v1.RolePermissionListRequest) (*v1.RolePermissionListReply, error) {
	return s.pu.RoleCodes(ctx, req.Role)
}

func (s *AdminService) RolePermissionUpdate(ctx context.Context, req *v1.RolePermissionUpdateRequest) (*v1.CheckResponse, error) {
	if err := s.pu.Assign(ctx, req.Role, req.Codes); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}
