package service

import (
	"context"

	v1 "shop/api/shop/v1"
)

func (s *ShopService) OrderCreate(ctx context.Context, req *v1.OrderCreateRequest) (*v1.OrderInfoReply, error) {
	return s.ou.Create(ctx, req)
}

func (s *ShopService) OrderCancel(ctx context.Context, req *v1.OrderCancelRequest) (*v1.CheckResponse, error) {
	if err := s.ou.Cancel(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *ShopService) OrderList(ctx context.Context, req *v1.OrderListRequest) (*v1.OrderListReply, error) {
	return s.ou.List(ctx, req)
}

func (s *ShopService) OrderDetail(ctx context.Context, req *v1.OrderDetailRequest) (*v1.OrderInfoReply, error) {
	return s.ou.Detail(ctx, req)
}
