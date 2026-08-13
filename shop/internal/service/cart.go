package service

import (
	"context"

	v1 "shop/api/shop/v1"
)

func (s *ShopService) CartList(ctx context.Context, req *v1.CartListRequest) (*v1.CartListReply, error) {
	return s.cu.List(ctx)
}

func (s *ShopService) CartAdd(ctx context.Context, req *v1.CartAddRequest) (*v1.CartItemReply, error) {
	return s.cu.Add(ctx, req)
}

func (s *ShopService) CartUpdate(ctx context.Context, req *v1.CartUpdateRequest) (*v1.CheckResponse, error) {
	if err := s.cu.Update(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *ShopService) CartDelete(ctx context.Context, req *v1.CartDeleteRequest) (*v1.CheckResponse, error) {
	if err := s.cu.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}
