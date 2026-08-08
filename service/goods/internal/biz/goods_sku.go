package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"goods/internal/domain"
)

type Sku struct {
	ID             int64
	GoodsID        int64
	GoodsSn        string
	GoodsName      string
	SkuName        string
	SkuCode        string
	BarCode        string
	Price          int64
	PromotionPrice int64
	Points         int64
	RemarksInfo    string
	Pic            string
	Inventory      int64
	OnSale         bool
	AttrInfo       string
}

type GoodsSkuRepo interface {
	Create(context.Context, *domain.GoodsSku) (*domain.GoodsSku, error)
	CreateSkuRelation(context.Context, []*domain.GoodsSpecificationSku) error
	ListByIds(context.Context, []int64) ([]*domain.GoodsSku, error)
	ListByGoodsID(context.Context, int64) ([]*domain.GoodsSku, error)
}

type GoodsSkuUsecase struct {
	repo GoodsSkuRepo
	log  *log.Helper
}

func NewGoodsSkuUsecase(repo GoodsSkuRepo, logger log.Logger) *GoodsSkuUsecase {
	return &GoodsSkuUsecase{repo: repo, log: log.NewHelper(logger)}
}
