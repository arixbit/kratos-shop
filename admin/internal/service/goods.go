package service

import (
	"context"

	v1 "admin/api/admin/v1"
)

func (s *AdminService) GoodsList(ctx context.Context, req *v1.GoodsListRequest) (*v1.GoodsListReply, error) {
	return s.gu.List(ctx, req)
}

func (s *AdminService) UpdateGoods(ctx context.Context, req *v1.UpdateGoodsRequest) (*v1.CheckResponse, error) {
	if err := s.gu.Update(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *AdminService) DeleteGoods(ctx context.Context, req *v1.DeleteGoodsRequest) (*v1.CheckResponse, error) {
	if err := s.gu.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *AdminService) UpdateGoodsStatus(ctx context.Context, req *v1.UpdateGoodsStatusRequest) (*v1.CheckResponse, error) {
	if err := s.gu.UpdateStatus(ctx, req.Id, req.OnSale); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}
