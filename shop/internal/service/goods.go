package service

import (
	"context"

	v1 "shop/api/shop/v1"
)

func (s *ShopService) GoodsList(ctx context.Context, req *v1.GoodsListRequest) (*v1.GoodsListReply, error) {
	return s.gu.List(ctx, req)
}

func (s *ShopService) GoodsDetail(ctx context.Context, req *v1.GoodsDetailRequest) (*v1.GoodsDetailReply, error) {
	return s.gu.Detail(ctx, req)
}
