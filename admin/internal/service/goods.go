package service

import (
	"context"

	v1 "admin/api/admin/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminService) GoodsList(ctx context.Context, req *v1.GoodsListRequest) (*v1.GoodsListReply, error) {
	return s.gu.List(ctx, req)
}

func (s *AdminService) CreateGoods(ctx context.Context, req *v1.CreateGoodsRequest) (*v1.CheckResponse, error) {
	if err := s.gu.Create(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *AdminService) GoodsDetail(ctx context.Context, req *v1.GoodsDetailRequest) (*v1.GoodsDetailReply, error) {
	return s.gu.Detail(ctx, req)
}

func (s *AdminService) CategoryList(ctx context.Context, req *emptypb.Empty) (*v1.CategoryListReply, error) {
	return s.gu.Categories(ctx)
}

func (s *AdminService) BrandList(ctx context.Context, req *v1.BrandListRequest) (*v1.BrandListReply, error) {
	return s.gu.Brands(ctx, req)
}

func (s *AdminService) CreateCategory(ctx context.Context, req *v1.CategorySaveRequest) (*v1.CategoryItem, error) {
	return s.gu.CreateCategory(ctx, req)
}

func (s *AdminService) UpdateCategory(ctx context.Context, req *v1.CategorySaveRequest) (*v1.CheckResponse, error) {
	if err := s.gu.UpdateCategory(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *AdminService) DeleteCategory(ctx context.Context, req *v1.CategoryDeleteRequest) (*v1.CheckResponse, error) {
	if err := s.gu.DeleteCategory(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
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
