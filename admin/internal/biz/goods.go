package biz

import (
	"context"

	goodsV1 "admin/api/service/goods/v1"
	v1 "admin/api/admin/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type GoodsUsecase struct {
	gc  goodsV1.GoodsClient
	log *log.Helper
}

func NewGoodsUsecase(gc goodsV1.GoodsClient, logger log.Logger) *GoodsUsecase {
	return &GoodsUsecase{gc: gc, log: log.NewHelper(logger)}
}

func (uc *GoodsUsecase) List(ctx context.Context, req *v1.GoodsListRequest) (*v1.GoodsListReply, error) {
	resp, err := uc.gc.AdminGoodsList(ctx, &goodsV1.GoodsFilterRequest{
		Pages:       int64(req.Page),
		PagePerNums: int64(req.PageSize),
		Keywords:    req.Keywords,
		CategoryId:  int32(req.CategoryId),
		BrandId:     int32(req.BrandId),
	})
	if err != nil {
		return nil, err
	}
	reply := &v1.GoodsListReply{Total: int32(resp.Total)}
	for _, item := range resp.List {
		reply.List = append(reply.List, &v1.GoodsItem{
			Id:          item.Id,
			Name:        item.Name,
			GoodsSn:     item.GoodsSn,
			MarketPrice: item.MarketPrice,
			OnSale:      item.OnSale,
			IsNew:       item.IsNew,
			IsHot:       item.IsHot,
		})
	}
	return reply, nil
}

func (uc *GoodsUsecase) Update(ctx context.Context, req *v1.UpdateGoodsRequest) error {
	_, err := uc.gc.UpdateGoodsInfo(ctx, &goodsV1.UpdateGoodsInfoRequest{
		Id:              req.Id,
		Name:            req.Name,
		NameAlias:       req.NameAlias,
		GoodsBrief:      req.GoodsBrief,
		GoodsFrontImage: req.GoodsFrontImage,
		MarketPrice:     req.MarketPrice,
		ShipFree:        req.ShipFree,
		IsNew:           req.IsNew,
		IsHot:           req.IsHot,
	})
	return err
}

func (uc *GoodsUsecase) Delete(ctx context.Context, id int64) error {
	_, err := uc.gc.DeleteGoods(ctx, &goodsV1.DeleteGoodsRequest{Id: id})
	return err
}

func (uc *GoodsUsecase) UpdateStatus(ctx context.Context, id int64, onSale bool) error {
	_, err := uc.gc.UpdateGoodsStatus(ctx, &goodsV1.UpdateGoodsStatusRequest{Id: id, OnSale: onSale})
	return err
}
