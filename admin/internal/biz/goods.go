package biz

import (
	"context"
	"encoding/json"

	v1 "admin/api/admin/v1"
	goodsV1 "admin/api/service/goods/v1"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
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
		SkuCode:     req.SkuCode,
	})
	if err != nil {
		return nil, err
	}
	reply := &v1.GoodsListReply{Total: int32(resp.Total)}
	for _, item := range resp.List {
		reply.List = append(reply.List, &v1.GoodsItem{
			Id:           item.Id,
			Name:         item.Name,
			GoodsSn:      item.GoodsSn,
			MarketPrice:  item.MarketPrice,
			OnSale:       item.OnSale,
			IsNew:        item.IsNew,
			IsHot:        item.IsHot,
			CategoryId:   int64(item.CategoryId),
			BrandId:      int64(item.BrandId),
			CategoryName: item.CategoryName,
			BrandName:    item.BrandName,
			SkuCount:     item.SkuCount,
			SoldNum:      item.SoldNum,
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
		MarketPrice:     resp.MarketPrice,
		GoodsBrief:      resp.GoodsBrief,
		GoodsFrontImage: resp.GoodsFrontImage,
		GoodsImages:     resp.GoodsImages,
		OnSale:          resp.OnSale,
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
			Inventory:      sku.Inventory,
			OnSale:         sku.OnSale,
			Pic:            sku.Pic,
		})
	}
	for _, image := range resp.Images {
		reply.Images = append(reply.Images, &v1.GoodsImageDetail{
			Url:      image.Url,
			Position: image.Position,
			IsMaster: image.IsMaster,
		})
	}
	if cats, err := uc.Categories(ctx); err == nil {
		for _, item := range cats.List {
			if item.Id == resp.CategoryId {
				reply.CategoryName = item.Name
				break
			}
		}
	}
	if brands, err := uc.Brands(ctx, &v1.BrandListRequest{PageSize: 100}); err == nil {
		for _, item := range brands.List {
			if item.Id == resp.BrandId {
				reply.BrandName = item.Name
				break
			}
		}
	}
	return reply, nil
}

func (uc *GoodsUsecase) Categories(ctx context.Context) (*v1.CategoryListReply, error) {
	resp, err := uc.gc.GetAllCategoryList(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	type categoryNode struct {
		ID               int32           `json:"ID"`
		Name             string          `json:"Name"`
		ParentCategoryID int32           `json:"ParentCategoryID"`
		Level            int32           `json:"Level"`
		SubCategory      []*categoryNode `json:"SubCategory"`
	}
	var nodes []*categoryNode
	if err := json.Unmarshal([]byte(resp.JsonData), &nodes); err != nil {
		return nil, err
	}
	reply := &v1.CategoryListReply{}
	var flatten func([]*categoryNode, int32)
	flatten = func(list []*categoryNode, parent int32) {
		for _, item := range list {
			reply.List = append(reply.List, &v1.CategoryItem{
				Id:             int64(item.ID),
				Name:           item.Name,
				ParentCategory: int64(item.ParentCategoryID),
				Level:          item.Level,
			})
			if len(item.SubCategory) > 0 {
				flatten(item.SubCategory, int32(item.ID))
			}
		}
	}
	flatten(nodes, 0)
	return reply, nil
}

func (uc *GoodsUsecase) Brands(ctx context.Context, req *v1.BrandListRequest) (*v1.BrandListReply, error) {
	pages := req.Page
	pageSize := req.PageSize
	if pages <= 0 {
		pages = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	resp, err := uc.gc.BrandList(ctx, &goodsV1.BrandListRequest{
		Pages:       pages,
		PagePerNums: pageSize,
	})
	if err != nil {
		return nil, err
	}
	reply := &v1.BrandListReply{}
	for _, item := range resp.Data {
		reply.List = append(reply.List, &v1.BrandItem{
			Id:   int64(item.Id),
			Name: item.Name,
			Logo: item.Logo,
		})
	}
	return reply, nil
}

func (uc *GoodsUsecase) CreateCategory(ctx context.Context, req *v1.CategorySaveRequest) (*v1.CategoryItem, error) {
	resp, err := uc.gc.CreateCategory(ctx, &goodsV1.CategoryInfoRequest{
		Id:             int32(req.Id),
		Name:           req.Name,
		ParentCategory: int32(req.ParentCategory),
		Level:          req.Level,
		Sort:           req.Sort,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CategoryItem{
		Id:             int64(resp.Id),
		Name:           resp.Name,
		ParentCategory: int64(resp.ParentCategory),
		Level:          resp.Level,
	}, nil
}

func (uc *GoodsUsecase) UpdateCategory(ctx context.Context, req *v1.CategorySaveRequest) error {
	_, err := uc.gc.UpdateCategory(ctx, &goodsV1.CategoryInfoRequest{
		Id:             int32(req.Id),
		Name:           req.Name,
		ParentCategory: int32(req.ParentCategory),
		Level:          req.Level,
		Sort:           req.Sort,
	})
	return err
}

func (uc *GoodsUsecase) DeleteCategory(ctx context.Context, id int64) error {
	_, err := uc.gc.DeleteCategory(ctx, &goodsV1.DeleteCategoryRequest{Id: int32(id)})
	return err
}

func (uc *GoodsUsecase) Create(ctx context.Context, req *v1.CreateGoodsRequest) error {
	_, err := uc.gc.CreateGoods(ctx, &goodsV1.CreateGoodsRequest{
		CategoryId:      req.CategoryId,
		BrandId:         req.BrandId,
		TypeId:          req.TypeId,
		Name:            req.Name,
		GoodsSn:         req.GoodsSn,
		MarketPrice:     req.MarketPrice,
		Inventory:       req.Inventory,
		GoodsBrief:      req.GoodsBrief,
		GoodsFrontImage: req.GoodsFrontImage,
		ShipFree:        req.ShipFree,
		IsNew:           req.IsNew,
		IsHot:           req.IsHot,
		OnSale:          req.OnSale,
		Sku: []*goodsV1.CreateGoodsRequestGoodsSku{
			{
				SkuName:        req.SkuName,
				Code:           req.SkuCode,
				BarCode:        req.BarCode,
				Price:          req.SkuPrice,
				PromotionPrice: req.PromotionPrice,
				Inventory:      req.SkuInventory,
				Image:          req.SkuImage,
			},
		},
	})
	return err
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
