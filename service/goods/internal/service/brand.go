package service

import (
	"context"
	v1 "goods/api/goods/v1"
	"goods/internal/biz"
	"goods/internal/domain"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateBrand 创建品牌
func (g *GoodsService) CreateBrand(ctx context.Context, r *v1.BrandRequest) (*v1.BrandInfoResponse, error) {
	brand, err := g.bc.CreateBrand(ctx, toBiz(r))
	if err != nil {
		return nil, err
	}

	return &v1.BrandInfoResponse{
		Id:    brand.ID,
		Name:  brand.Name,
		Logo:  brand.Logo,
		Desc:  brand.Desc,
		IsTab: brand.IsTab,
		Sort:  brand.Sort,
	}, nil
}

func (g *GoodsService) UpdateBrand(ctx context.Context, r *v1.BrandRequest) (*emptypb.Empty, error) {
	err := g.bc.UpdateBrand(ctx, toBiz(r))
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (g *GoodsService) BrandList(ctx context.Context, r *v1.BrandListRequest) (*v1.BrandListResponse, error) {
	list, total, err := g.bc.BrandList(ctx, &biz.Pagination{
		PageNum:  int(r.Pages),
		PageSize: int(r.PagePerNums),
	})
	if err != nil {
		return nil, err
	}
	var rs v1.BrandListResponse
	rs.Total = int32(total)
	for _, x := range list {
		info := toProto(x)
		rs.Data = append(rs.Data, info)
	}
	return &rs, nil
}

func toProto(r *domain.Brand) *v1.BrandInfoResponse {
	return &v1.BrandInfoResponse{
		Id:    r.ID,
		Name:  r.Name,
		Logo:  r.Logo,
		Desc:  r.Desc,
		IsTab: r.IsTab,
		Sort:  r.Sort,
	}
}

func toBiz(r *v1.BrandRequest) *domain.Brand {
	return &domain.Brand{
		ID:    r.Id,
		Name:  r.Name,
		Logo:  r.Logo,
		Desc:  r.Desc,
		IsTab: r.IsTab,
		Sort:  r.Sort,
	}
}
