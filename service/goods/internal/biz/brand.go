package biz

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"goods/internal/domain"
)

type Pagination struct {
	PageNum  int
	PageSize int
}

type BrandRepo interface {
	Create(context.Context, *domain.Brand) (*domain.Brand, error)
	GetBradByName(context.Context, string) (*domain.Brand, error)
	Update(context.Context, *domain.Brand) error
	List(context.Context, *Pagination) ([]*domain.Brand, int64, error)
	IsBrandByID(context.Context, int32) (*domain.Brand, error)
	IsBrand(context.Context, []int32) error
	ListByIds(context.Context, ...int32) (domain.BrandList, error)
}

type BrandUsecase struct {
	repo BrandRepo
	log  *log.Helper
}

func NewBrandUsecase(repo BrandRepo, logger log.Logger) *BrandUsecase {
	return &BrandUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *BrandUsecase) CreateBrand(ctx context.Context, b *domain.Brand) (*domain.Brand, error) {
	_, err := uc.repo.GetBradByName(ctx, b.Name)
	if err != nil {
		return uc.repo.Create(ctx, b)
	} else {
		return nil, errors.New("当前品牌已经存在")
	}
}

func (uc *BrandUsecase) UpdateBrand(ctx context.Context, b *domain.Brand) error {
	err := uc.repo.Update(ctx, b)
	if err != nil {
		return err
	}
	return nil
}

func (uc *BrandUsecase) BrandList(ctx context.Context, b *Pagination) ([]*domain.Brand, int64, error) {
	list, total, err := uc.repo.List(ctx, b)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil

}

func (uc *BrandUsecase) ListByIds(ctx context.Context, ids ...int32) (domain.BrandList, error) {
	return uc.repo.ListByIds(ctx, ids...)
}
