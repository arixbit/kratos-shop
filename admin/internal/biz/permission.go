package biz

import (
	"context"

	v1 "admin/api/admin/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type Permission struct {
	ID        int64
	Code      string
	Name      string
	GroupName string
	Sort      int32
}

type PermissionRepo interface {
	List(ctx context.Context) ([]*Permission, error)
	ListCodesByRole(ctx context.Context, role int32) ([]string, error)
	AssignRole(ctx context.Context, role int32, codes []string) error
}

type PermissionUsecase struct {
	repo PermissionRepo
	log  *log.Helper
}

func NewPermissionUsecase(repo PermissionRepo, logger log.Logger) *PermissionUsecase {
	return &PermissionUsecase{repo: repo, log: log.NewHelper(log.With(logger, "module", "usecase/permission"))}
}

func (uc *PermissionUsecase) List(ctx context.Context) (*v1.PermissionListReply, error) {
	list, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	reply := &v1.PermissionListReply{}
	for _, item := range list {
		reply.List = append(reply.List, &v1.PermissionItem{
			Id:        item.ID,
			Code:      item.Code,
			Name:      item.Name,
			GroupName: item.GroupName,
			Sort:      item.Sort,
		})
	}
	return reply, nil
}

func (uc *PermissionUsecase) RoleCodes(ctx context.Context, role int32) (*v1.RolePermissionListReply, error) {
	codes, err := uc.repo.ListCodesByRole(ctx, role)
	if err != nil {
		return nil, err
	}
	return &v1.RolePermissionListReply{Codes: codes}, nil
}

func (uc *PermissionUsecase) Assign(ctx context.Context, role int32, codes []string) error {
	return uc.repo.AssignRole(ctx, role, codes)
}

func (uc *PermissionUsecase) CodesForRole(ctx context.Context, role int32) ([]string, error) {
	return uc.repo.ListCodesByRole(ctx, role)
}
