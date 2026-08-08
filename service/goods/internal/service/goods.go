package service

import (
	"context"
	"encoding/json"
	"time"

	inventoryV1 "goods/api/service/inventory/v1"
	v1 "goods/api/goods/v1"
	"goods/internal/domain"
	"goods/internal/pkg/mq"
)

// CreateGoods 创建商品
func (g *GoodsService) CreateGoods(ctx context.Context, r *v1.CreateGoodsRequest) (*v1.CreateGoodsResponse, error) {
	var goodsSku []*domain.GoodsSku
	for _, sku := range r.Sku {
		res := &domain.GoodsSku{
			GoodsName:      r.Name,
			GoodsSn:        r.GoodsSn,
			SkuName:        sku.SkuName,
			SkuCode:        sku.Code,
			BarCode:        sku.BarCode,
			Price:          sku.Price,
			PromotionPrice: sku.PromotionPrice,
			Points:         sku.Points,
			Pic:            sku.Image,
			Inventory:      sku.Inventory,
			OnSale:         r.OnSale,
		}

		for _, specification := range sku.SpecificationInfo {
			s := &domain.SpecificationInfo{
				SpecificationID:      specification.SId,
				SpecificationValueID: specification.VId,
			}
			res.Specification = append(res.Specification, s)
		}
		for _, attrGroup := range sku.GroupAttrInfo {
			group := &domain.GroupAttr{
				GroupId:   attrGroup.GroupId,
				GroupName: attrGroup.GroupName,
			}
			for _, attr := range attrGroup.AttrInfo {
				s := &domain.Attr{
					AttrID:        attr.AttrId,
					AttrName:      attr.AttrName,
					AttrValueID:   attr.AttrValueId,
					AttrValueName: attr.AttrValueName,
				}
				group.Attr = append(group.Attr, s)
			}
			res.GroupAttr = append(res.GroupAttr, group)
		}
		goodsSku = append(goodsSku, res)
	}

	goodsInfo := domain.Goods{
		ID:              r.Id,
		CategoryID:      r.CategoryId,
		BrandsID:        r.BrandId,
		TypeID:          r.TypeId,
		Name:            r.Name,
		NameAlias:       r.NameAlias,
		GoodsSn:         r.GoodsSn,
		GoodsTags:       r.GoodsTags,
		MarketPrice:     r.MarketPrice,
		GoodsBrief:      r.GoodsBrief,
		GoodsFrontImage: r.GoodsFrontImage,
		GoodsImages:     r.GoodsImages,
		OnSale:          r.OnSale,
		ShipFree:        r.ShipFree,
		ShipID:          r.ShipId,
		IsNew:           r.IsNew,
		IsHot:           r.IsHot,
		Sku:             goodsSku,
	}

	result, err := g.g.CreateGoods(ctx, &goodsInfo)
	if err != nil {
		return nil, err
	}
	return &v1.CreateGoodsResponse{ID: result.GoodsID}, nil

}

func (g *GoodsService) GoodsList(ctx context.Context, r *v1.GoodsFilterRequest) (*v1.GoodsListResponse, error) {
	goodsFilter := &domain.ESGoodsFilter{
		ID:          r.Id,
		CategoryID:  r.CategoryId,
		BrandsID:    r.BrandId,
		Keywords:    r.Keywords,
		IsNew:       r.IsNew,
		IsHot:       r.IsHot,
		ClickNum:    r.ClickNum,
		SoldNum:     r.SoldNum,
		FavNum:      r.FavNum,
		MaxPrice:    r.MaxPrice,
		MinPrice:    r.MinPrice,
		Pages:       r.Pages,
		PagePerNums: r.PagePerNums,
	}

	result, err := g.esGoods.GoodsList(ctx, goodsFilter)
	if err != nil {
		return nil, err
	}
	response := v1.GoodsListResponse{
		Total: result.Total,
	}

	var brandIDs []int32
	categoryIDs := make(map[int32]struct{})
	for _, goods := range result.List {
		brandIDs = append(brandIDs, goods.BrandsID)
		categoryIDs[goods.CategoryID] = struct{}{}
	}
	brandList, _ := g.bc.ListByIds(ctx, brandIDs...)
	brandMap := make(map[int32]string, len(brandList))
	for _, brand := range brandList {
		brandMap[brand.ID] = brand.Name
	}
	categoryMap := make(map[int32]string, len(categoryIDs))
	for id := range categoryIDs {
		if cate, err := g.cac.GetCategoryByID(ctx, id); err == nil {
			categoryMap[id] = cate.Name
		}
	}

	for _, goods := range result.List {
		res := v1.GoodsInfoResponse{
			Id:          goods.ID,
			CategoryId:  goods.CategoryID,
			BrandId:     goods.BrandsID,
			Name:        goods.Name,
			GoodsSn:     goods.GoodsSn,
			ClickNum:    goods.ClickNum,
			SoldNum:     goods.SoldNum,
			FavNum:      goods.FavNum,
			MarketPrice: goods.MarketPrice,
			GoodsBrief:  goods.GoodsBrief,
			GoodsDesc:   goods.GoodsBrief,
			ShipFree:    goods.ShipFree,
			Images:      goods.GoodsFrontImage,
			GoodsImages: goods.GoodsImages,
			IsNew:       goods.IsNew,
			IsHot:       goods.IsHot,
			OnSale:      goods.OnSale,
			BrandName:   brandMap[goods.BrandsID],
			CategoryName: categoryMap[goods.CategoryID],
		}
		if skus, err := g.g.GetSkusByGoodsID(ctx, goods.ID); err == nil {
			for _, sku := range skus {
				res.Skus = append(res.Skus, &v1.SkuInfo{
					Id:             sku.ID,
					GoodsId:        sku.GoodsID,
					GoodsSn:        sku.GoodsSn,
					GoodsName:      sku.GoodsName,
					SkuName:        sku.SkuName,
					SkuCode:        sku.SkuCode,
					BarCode:        sku.BarCode,
					Price:          sku.Price,
					PromotionPrice: sku.PromotionPrice,
					Points:         sku.Points,
					Pic:            sku.Pic,
					OnSale:         sku.OnSale,
					Inventory:      sku.Inventory,
					AttrInfo:       sku.AttrInfo,
				})
			}
			g.applyStock(ctx, res.Skus)
		}
		response.List = append(response.List, &res)
	}
	return &response, nil
}

// SkuList 按 SKU ID 批量查询 SKU 信息
func (g *GoodsService) SkuList(ctx context.Context, r *v1.SkuListRequest) (*v1.SkuListResponse, error) {
	skus, err := g.g.SkuListByIds(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	resp := &v1.SkuListResponse{}
	for _, sku := range skus {
		resp.List = append(resp.List, &v1.SkuInfo{
			Id:             sku.ID,
			GoodsId:        sku.GoodsID,
			GoodsSn:        sku.GoodsSn,
			GoodsName:      sku.GoodsName,
			SkuName:        sku.SkuName,
			SkuCode:        sku.SkuCode,
			BarCode:        sku.BarCode,
			Price:          sku.Price,
			PromotionPrice: sku.PromotionPrice,
			Points:         sku.Points,
			Pic:            sku.Pic,
			OnSale:         sku.OnSale,
			Inventory:      sku.Inventory,
			AttrInfo:       sku.AttrInfo,
		})
	}
	g.applyStock(ctx, resp.List)
	return resp, nil
}

func (g *GoodsService) UpdateGoodsInfo(ctx context.Context, r *v1.UpdateGoodsInfoRequest) (*v1.CheckResponse, error) {
	if err := g.g.UpdateGoods(ctx, &domain.Goods{
		ID:              r.Id,
		Name:            r.Name,
		NameAlias:       r.NameAlias,
		GoodsBrief:      r.GoodsBrief,
		GoodsFrontImage: r.GoodsFrontImage,
		MarketPrice:     r.MarketPrice,
		ShipFree:        r.ShipFree,
		IsNew:           r.IsNew,
		IsHot:           r.IsHot,
	}); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (g *GoodsService) DeleteGoods(ctx context.Context, r *v1.DeleteGoodsRequest) (*v1.CheckResponse, error) {
	if err := g.g.DeleteGoods(ctx, r.Id); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (g *GoodsService) UpdateGoodsStatus(ctx context.Context, r *v1.UpdateGoodsStatusRequest) (*v1.CheckResponse, error) {
	if err := g.g.UpdateGoodsStatus(ctx, r.Id, r.OnSale); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (g *GoodsService) AdminGoodsList(ctx context.Context, r *v1.GoodsFilterRequest) (*v1.GoodsListResponse, error) {
	list, total, err := g.g.AdminGoodsList(ctx, r.Keywords, r.CategoryId, r.BrandId, int(r.Pages), int(r.PagePerNums))
	if err != nil {
		return nil, err
	}
	response := &v1.GoodsListResponse{Total: total}
	for _, goods := range list {
		response.List = append(response.List, &v1.GoodsInfoResponse{
			Id:          goods.ID,
			CategoryId:  goods.CategoryID,
			BrandId:     goods.BrandsID,
			Name:        goods.Name,
			GoodsSn:     goods.GoodsSn,
			MarketPrice: goods.MarketPrice,
			GoodsBrief:  goods.GoodsBrief,
			ShipFree:    goods.ShipFree,
			IsNew:       goods.IsNew,
			IsHot:       goods.IsHot,
			OnSale:      goods.OnSale,
		})
	}
	return response, nil
}

func (g *GoodsService) GoodsDetail(ctx context.Context, r *v1.GoodsDetailRequest) (*v1.GoodsDetailResponse, error) {
	goods, skus, images, err := g.g.GetDetail(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	resp := &v1.GoodsDetailResponse{
		Id:              goods.ID,
		CategoryId:      int64(goods.CategoryID),
		BrandId:         int64(goods.BrandsID),
		TypeId:          goods.TypeID,
		Name:            goods.Name,
		GoodsSn:         goods.GoodsSn,
		GoodsBrief:      goods.GoodsBrief,
		GoodsFrontImage: goods.GoodsFrontImage,
		GoodsImages:     goods.GoodsImages,
		MarketPrice:     goods.MarketPrice,
		OnSale:          goods.OnSale,
		ShipFree:        goods.ShipFree,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		SoldNum:         goods.SoldNum,
	}
	for _, sku := range skus {
		resp.Skus = append(resp.Skus, &v1.SkuDetail{
			Id:              sku.ID,
			GoodsId:         sku.GoodsID,
			SkuName:         sku.SkuName,
			SkuCode:         sku.SkuCode,
			Price:           sku.Price,
			PromotionPrice:  sku.PromotionPrice,
			Pic:             sku.Pic,
			Inventory:       sku.Inventory,
			OnSale:          sku.OnSale,
		})
	}
	g.applyStockDetail(ctx, resp.Skus)
	for _, image := range images {
		resp.Images = append(resp.Images, &v1.ImageInfo{
			Url:      image.Link,
			Position: image.Position,
			IsMaster: image.IsMaster,
		})
	}
	return resp, nil
}

func (g *GoodsService) applyStock(ctx context.Context, skus []*v1.SkuInfo) {
	if g.inv == nil || len(skus) == 0 {
		return
	}
	ids := make([]int64, 0, len(skus))
	for _, sku := range skus {
		ids = append(ids, sku.Id)
	}
	stock, err := g.inv.QueryStock(ctx, &inventoryV1.StockQueryRequest{SkuIds: ids})
	if err != nil {
		g.log.Errorf("query stock failed: %v", err)
		return
	}
	stockMap := make(map[int64]int64, len(stock.Items))
	for _, item := range stock.Items {
		stockMap[item.SkuId] = item.Available
	}
	for _, sku := range skus {
		if v, ok := stockMap[sku.Id]; ok {
			sku.Inventory = v
		}
	}
}

func (g *GoodsService) applyStockDetail(ctx context.Context, skus []*v1.SkuDetail) {
	if g.inv == nil || len(skus) == 0 {
		return
	}
	ids := make([]int64, 0, len(skus))
	for _, sku := range skus {
		ids = append(ids, sku.Id)
	}
	stock, err := g.inv.QueryStock(ctx, &inventoryV1.StockQueryRequest{SkuIds: ids})
	if err != nil {
		g.log.Errorf("query stock failed: %v", err)
		return
	}
	stockMap := make(map[int64]int64, len(stock.Items))
	for _, item := range stock.Items {
		stockMap[item.SkuId] = item.Available
	}
	for _, sku := range skus {
		if v, ok := stockMap[sku.Id]; ok {
			sku.Inventory = v
		}
	}
}

type orderPaidEvent struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	OrderSn   string `json:"order_sn"`
	UserID    int64  `json:"user_id"`
	Payload   struct {
		Skus []struct {
			SkuID int64 `json:"sku_id"`
			Num   int32 `json:"num"`
		} `json:"skus"`
	} `json:"payload"`
}

func (g *GoodsService) startPaidConsumer() {
	for {
		consumer, err := mq.NewConsumer(
			g.mqAddr,
			"order.exchange",
			"q.order.paid.goods",
			[]string{"order.paid"},
			g.handleOrderPaid,
		)
		if err != nil {
			g.log.Errorf("paid consumer init failed: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if err := consumer.Run(context.Background()); err != nil {
			g.log.Errorf("paid consumer stopped: %v", err)
		}
		consumer.Close()
		time.Sleep(5 * time.Second)
	}
}

func (g *GoodsService) handleOrderPaid(ctx context.Context, body []byte) error {
	var evt orderPaidEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}
	consumed, err := g.g.IsConsumed(ctx, evt.EventID)
	if err != nil {
		return err
	}
	if consumed {
		return nil
	}
	skuItems := make(map[int64]int32, len(evt.Payload.Skus))
	for _, sku := range evt.Payload.Skus {
		skuItems[sku.SkuID] += sku.Num
	}
	if err := g.g.IncrSoldNum(ctx, skuItems); err != nil {
		g.log.Errorf("incr sold num failed: order=%s err=%v", evt.OrderSn, err)
		return err
	}
	if err := g.g.MarkConsumed(ctx, evt.EventID, evt.OrderSn); err != nil {
		g.log.Errorf("mark consumed failed: order=%s err=%v", evt.OrderSn, err)
		return err
	}
	g.log.Infof("goods sold num updated: order=%s", evt.OrderSn)
	return nil
}
