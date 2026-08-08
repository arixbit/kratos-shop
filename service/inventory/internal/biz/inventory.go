package biz

import (
	"context"

	"inventory/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewInventoryUsecase)

type InventoryRepo interface {
	Query(ctx context.Context, skuIds []int64) ([]*domain.Inventory, error)
	TryLock(ctx context.Context, orderSn string, items []*domain.SkuItem) error
	ConfirmDeduct(ctx context.Context, orderSn string, items []*domain.SkuItem) error
	Release(ctx context.Context, orderSn string, items []*domain.SkuItem) error
	IsConsumed(ctx context.Context, eventID string) (bool, error)
	MarkConsumed(ctx context.Context, eventID, orderSn string) error
}

type InventoryUsecase struct {
	repo InventoryRepo
	log  *log.Helper
}

func NewInventoryUsecase(repo InventoryRepo, logger log.Logger) *InventoryUsecase {
	return &InventoryUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *InventoryUsecase) Query(ctx context.Context, skuIds []int64) ([]*domain.Inventory, error) {
	return uc.repo.Query(ctx, skuIds)
}

func (uc *InventoryUsecase) TryLock(ctx context.Context, orderSn string, items []*domain.SkuItem) error {
	return uc.repo.TryLock(ctx, orderSn, items)
}

func (uc *InventoryUsecase) ConfirmDeduct(ctx context.Context, orderSn string, items []*domain.SkuItem) error {
	return uc.repo.ConfirmDeduct(ctx, orderSn, items)
}

func (uc *InventoryUsecase) Release(ctx context.Context, orderSn string, items []*domain.SkuItem) error {
	return uc.repo.Release(ctx, orderSn, items)
}

func (uc *InventoryUsecase) IsConsumed(ctx context.Context, eventID string) (bool, error) {
	return uc.repo.IsConsumed(ctx, eventID)
}

func (uc *InventoryUsecase) MarkConsumed(ctx context.Context, eventID, orderSn string) error {
	return uc.repo.MarkConsumed(ctx, eventID, orderSn)
}
