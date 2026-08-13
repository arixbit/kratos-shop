package service

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	v1 "shop/api/shop/v1"
	"shop/internal/biz"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewShopService)

// ShopService is a shop service.
type ShopService struct {
	v1.UnimplementedShopServer

	uc  *biz.UserUsecase
	ua  *biz.AddressUsecase
	gu  *biz.GoodsUsecase
	cu  *biz.CartUsecase
	ou  *biz.OrderUsecase
	pu  *biz.PaymentUsecase
	log *log.Helper
}

// NewShopService new a shop service.
func NewShopService(uc *biz.UserUsecase, ua *biz.AddressUsecase, gu *biz.GoodsUsecase, cu *biz.CartUsecase, ou *biz.OrderUsecase, pu *biz.PaymentUsecase, logger log.Logger) *ShopService {
	return &ShopService{
		uc:  uc,
		ua:  ua,
		gu:  gu,
		cu:  cu,
		ou:  ou,
		pu:  pu,
		log: log.NewHelper(log.With(logger, "module", "service/shop")),
	}
}
