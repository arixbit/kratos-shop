package service

import (
	"context"

	v1 "shop/api/shop/v1"
)

func (s *ShopService) PaymentCreate(ctx context.Context, req *v1.PaymentCreateRequest) (*v1.PaymentInfoReply, error) {
	return s.pu.Create(ctx, req)
}

func (s *ShopService) PaymentMockPay(ctx context.Context, req *v1.PaymentMockPayRequest) (*v1.CheckResponse, error) {
	if err := s.pu.MockPay(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}
