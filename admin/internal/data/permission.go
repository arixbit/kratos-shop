package data

import (
	"context"

	"admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type Permission struct {
	ID        int64
	Code      string
	Name      string
	GroupName string
	Sort      int32
}

func (Permission) TableName() string {
	return "permissions"
}

type RolePermission struct {
	ID           int64
	RoleID       int32
	PermissionID int64
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

type permissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewPermissionRepo(data *Data, logger log.Logger) biz.PermissionRepo {
	return &permissionRepo{data: data, log: log.NewHelper(log.With(logger, "module", "repo/permission"))}
}

func (r *permissionRepo) List(ctx context.Context) ([]*biz.Permission, error) {
	var rows []Permission
	if err := r.data.db.WithContext(ctx).Order("group_name ASC, sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]*biz.Permission, 0, len(rows))
	for i := range rows {
		res = append(res, &biz.Permission{
			ID:        rows[i].ID,
			Code:      rows[i].Code,
			Name:      rows[i].Name,
			GroupName: rows[i].GroupName,
			Sort:      rows[i].Sort,
		})
	}
	return res, nil
}

func (r *permissionRepo) ListCodesByRole(ctx context.Context, role int32) ([]string, error) {
	var codes []string
	if err := r.data.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", role).
		Order("permissions.sort ASC, permissions.id ASC").
		Pluck("permissions.code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *permissionRepo) AssignRole(ctx context.Context, role int32, codes []string) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", role).Delete(&RolePermission{}).Error; err != nil {
			return err
		}
		if len(codes) == 0 {
			return nil
		}
		var perms []Permission
		if err := tx.Where("code IN ?", codes).Find(&perms).Error; err != nil {
			return err
		}
		for _, p := range perms {
			if err := tx.Create(&RolePermission{RoleID: role, PermissionID: p.ID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
