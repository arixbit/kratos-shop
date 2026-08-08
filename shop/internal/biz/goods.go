package biz

import (
	"context"

	goodsV1 "shop/api/service/goods/v1"
	v1 "shop/api/shop/v1"

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
		if !item.OnSale {
			continue
		}
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

func (uc *GoodsUsecase) Detail(ctx context.Context, req *v1.GoodsDetailRequest) (*v1.GoodsDetailReply, error) {
	resp, err := uc.gc.GoodsDetail(ctx, &goodsV1.GoodsDetailRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	reply := &v1.GoodsDetailReply{
		Id:              resp.Id,
		CategoryId:      resp.CategoryId,
		BrandId:         resp.BrandId,
		TypeId:          resp.TypeId,
		Name:            resp.Name,
		GoodsSn:         resp.GoodsSn,
		GoodsBrief:      resp.GoodsBrief,
		GoodsFrontImage: resp.GoodsFrontImage,
		GoodsImages:     resp.GoodsImages,
		MarketPrice:     resp.MarketPrice,
		OnSale:          resp.OnSale,
		ShipFree:        resp.ShipFree,
		IsNew:           resp.IsNew,
		IsHot:           resp.IsHot,
		SoldNum:         resp.SoldNum,
	}
	for _, sku := range resp.Skus {
		reply.Skus = append(reply.Skus, &v1.GoodsSkuDetail{
			Id:             sku.Id,
			GoodsId:        sku.GoodsId,
			SkuName:        sku.SkuName,
			SkuCode:        sku.SkuCode,
			Price:          sku.Price,
			PromotionPrice: sku.PromotionPrice,
			Pic:            sku.Pic,
			Inventory:      sku.Inventory,
			OnSale:         sku.OnSale,
		})
	}
	for _, image := range resp.Images {
		reply.Images = append(reply.Images, &v1.GoodsImageDetail{
			Url:      image.Url,
			Position: image.Position,
			IsMaster: image.IsMaster,
		})
	}
	return reply, nil
}
