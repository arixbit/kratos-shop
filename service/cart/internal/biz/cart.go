package biz

import (
	"cart/internal/domain"
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type CartRepo interface {
	Create(ctx context.Context, c *domain.ShopCart) (*domain.ShopCart, error)
	List(ctx context.Context, userId int64) (domain.ShopCartList, error)
	Update(ctx context.Context, c *domain.ShopCart) error
	Delete(ctx context.Context, id, userId int64) error
	Get(ctx context.Context, id, userId int64) (*domain.ShopCart, error)
}

type CartUsecase struct {
	repo CartRepo
	log  *log.Helper
}

func NewCartUsecase(repo CartRepo, logger log.Logger) *CartUsecase {
	return &CartUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *CartUsecase) CreateCart(ctx context.Context, c *domain.ShopCart) (*domain.ShopCart, error) {
	return uc.repo.Create(ctx, c)
}
func (uc *CartUsecase) List(ctx context.Context, userId int64) (domain.ShopCartList, error) {
	return uc.repo.List(ctx, userId)
}

func (uc *CartUsecase) UpdateCart(ctx context.Context, c *domain.ShopCart) error {
	return uc.repo.Update(ctx, c)
}

func (uc *CartUsecase) DeleteCart(ctx context.Context, id, userId int64) error {
	return uc.repo.Delete(ctx, id, userId)
}

func (uc *CartUsecase) GetCart(ctx context.Context, id, userId int64) (*domain.ShopCart, error) {
	return uc.repo.Get(ctx, id, userId)
}

func (uc *CartUsecase) ListSelected(ctx context.Context, userId int64) (domain.ShopCartList, error) {
	res, err := uc.repo.List(ctx, userId)
	if err != nil {
		return nil, err
	}
	res = res.ListSelected()
	return res, nil
}
