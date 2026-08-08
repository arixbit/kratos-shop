package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"goods/internal/domain"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type GoodsRepo interface {
	CreateGoods(ctx context.Context, goods *domain.Goods) (*domain.Goods, error)
	GoodsListByIDs(context.Context, ...int64) ([]*domain.Goods, error)
	IncrSoldNum(context.Context, map[int64]int32) error
	IsConsumed(context.Context, string) (bool, error)
	MarkConsumed(context.Context, string, string) error
	Update(context.Context, *domain.Goods) error
	Delete(context.Context, int64) error
	UpdateStatus(context.Context, int64, bool) error
	ListPage(context.Context, string, int32, int32, int, int) ([]*domain.Goods, int64, error)
	GetByID(context.Context, int64) (*domain.Goods, error)
	ListAll(context.Context) ([]*domain.Goods, error)
	ListImagesByGoodsID(context.Context, int64) ([]*domain.GoodsImage, error)
}

type GoodsUsecase struct {
	repo              GoodsRepo
	tr                Transaction
	skuRepo           GoodsSkuRepo
	categoryRepo      CategoryRepo
	brandRepo         BrandRepo
	typeRepo          GoodsTypeRepo
	specificationRepo SpecificationRepo
	goodsAttrRepo     GoodsAttrRepo
	inventoryRepo     InventoryRepo
	esGoodsRepo       EsGoodsRepo
	log               *log.Helper
}

func NewGoodsUsecase(repo GoodsRepo, skuRepo GoodsSkuRepo, tx Transaction, gRepo GoodsTypeRepo, cRepo CategoryRepo,
	bRepo BrandRepo, sRepo SpecificationRepo, aRepo GoodsAttrRepo, iRepo InventoryRepo, es EsGoodsRepo, logger log.Logger) *GoodsUsecase {

	return &GoodsUsecase{
		repo:              repo,
		skuRepo:           skuRepo,
		tr:                tx,
		typeRepo:          gRepo,
		categoryRepo:      cRepo,
		brandRepo:         bRepo,
		specificationRepo: sRepo,
		goodsAttrRepo:     aRepo,
		inventoryRepo:     iRepo,
		esGoodsRepo:       es,
		log:               log.NewHelper(logger),
	}
}

func (g GoodsUsecase) SkuListByIds(ctx context.Context, ids []int64) ([]*domain.GoodsSku, error) {
	return g.skuRepo.ListByIds(ctx, ids)
}

func (g GoodsUsecase) IncrSoldNum(ctx context.Context, skuItems map[int64]int32) error {
	return g.repo.IncrSoldNum(ctx, skuItems)
}

func (g GoodsUsecase) IsConsumed(ctx context.Context, eventID string) (bool, error) {
	return g.repo.IsConsumed(ctx, eventID)
}

func (g GoodsUsecase) MarkConsumed(ctx context.Context, eventID, orderSn string) error {
	return g.repo.MarkConsumed(ctx, eventID, orderSn)
}

func (g GoodsUsecase) UpdateGoods(ctx context.Context, goods *domain.Goods) error {
	return g.repo.Update(ctx, goods)
}

func (g GoodsUsecase) DeleteGoods(ctx context.Context, id int64) error {
	return g.repo.Delete(ctx, id)
}

func (g GoodsUsecase) UpdateGoodsStatus(ctx context.Context, id int64, onSale bool) error {
	return g.repo.UpdateStatus(ctx, id, onSale)
}

func (g GoodsUsecase) AdminGoodsList(ctx context.Context, keywords string, categoryID, brandID int32, page, pageSize int) ([]*domain.Goods, int64, error) {
	return g.repo.ListPage(ctx, keywords, categoryID, brandID, page, pageSize)
}

func (g GoodsUsecase) GetDetail(ctx context.Context, id int64) (*domain.Goods, []*domain.GoodsSku, []*domain.GoodsImage, error) {
	goods, err := g.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	skus, err := g.skuRepo.ListByGoodsID(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	images, err := g.repo.ListImagesByGoodsID(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	return goods, skus, images, nil
}

func (g GoodsUsecase) GetSkusByGoodsID(ctx context.Context, goodsID int64) ([]*domain.GoodsSku, error) {
	return g.skuRepo.ListByGoodsID(ctx, goodsID)
}

func (g GoodsUsecase) ReindexAll(ctx context.Context) error {
	goodsList, err := g.repo.ListAll(ctx)
	if err != nil {
		return err
	}
	for _, goods := range goodsList {
		es := domain.ESGoods{
			ID:              goods.ID,
			CategoryID:      goods.CategoryID,
			BrandsID:        goods.BrandsID,
			TypeID:          goods.TypeID,
			Name:            goods.Name,
			GoodsTags:       goods.GoodsTags,
			ClickNum:        goods.ClickNum,
			SoldNum:         goods.SoldNum,
			FavNum:          goods.FavNum,
			MarketPrice:     goods.MarketPrice,
			GoodsBrief:      goods.GoodsBrief,
			OnSale:          goods.OnSale,
			ShipFree:        goods.ShipFree,
			IsNew:           goods.IsNew,
			IsHot:           goods.IsHot,
		}
		if brand, err := g.brandRepo.IsBrandByID(ctx, goods.BrandsID); err == nil {
			es.BrandName = brand.Name
		}
		if cate, err := g.categoryRepo.GetCategoryByID(ctx, goods.CategoryID); err == nil {
			es.CategoryName = cate.Name
		}
		if gt, err := g.typeRepo.IsExistsByID(ctx, goods.TypeID); err == nil {
			es.TypeName = gt.Name
		}
		skus, err := g.skuRepo.ListByGoodsID(ctx, goods.ID)
		if err == nil {
			for _, sku := range skus {
				es.Sku = append(es.Sku, domain.EsSku{
					SkuID:    sku.ID,
					SkuName:  sku.SkuName,
					SkuPrice: sku.Price,
				})
			}
		}
		if err := g.esGoodsRepo.InsertEsGoods(ctx, es); err != nil {
			return err
		}
	}
	return nil
}

func (g GoodsUsecase) CreateGoods(ctx context.Context, r *domain.Goods) (*domain.GoodsInfoResponse, error) {
	var (
		err     error
		goods   *domain.Goods
		EsGoods domain.ESGoods
	)

	// 判断品牌是否存在
	brand, err := g.brandRepo.IsBrandByID(ctx, r.BrandsID)
	if err != nil {
		return nil, err
	}

	// 判断分类是否存在
	cate, err := g.categoryRepo.GetCategoryByID(ctx, r.CategoryID)
	if err != nil {
		return nil, err
	}

	// 判断商品类型是否存在
	goodsType, err := g.typeRepo.IsExistsByID(ctx, r.TypeID)
	if err != nil {
		return nil, err
	}

	// 判断商品规格和属性是否存在
	for _, sku := range r.Sku {
		var sIDs []*int64
		for _, info := range sku.Specification {
			sIDs = append(sIDs, &info.SpecificationID)
		}

		specList, err := g.specificationRepo.ListByIds(ctx, sIDs...)
		if err != nil {
			return nil, err
		}
		for _, sId := range sIDs {
			info := specList.FindById(*sId)
			if info == nil {
				return nil, errors.NotFound("SPECIFICATION_NOT_FOUND", "商品规格不存在")
			}
		}
		var attrIDs []int64
		for _, attr := range sku.GroupAttr {
			for _, id := range attr.Attr {
				attrIDs = append(attrIDs, id.AttrID)
			}
		}
		attrList, err := g.goodsAttrRepo.ListByIds(ctx, attrIDs...)
		if err != nil {
			return nil, err
		}

		for _, attr := range sku.GroupAttr {
			for _, id := range attr.Attr {
				attrIDs = append(attrIDs, id.AttrID)
				true := attrList.IsNotExist(attr.GroupId, id.AttrID)
				if true {
					return nil, errors.NotFound("ATTR_NOT_FOUND", "商品属性不存在")
				}
			}
		}
	}

	err = g.tr.ExecTx(ctx, func(ctx context.Context) error {
		// 更新商品表
		goods, err = g.repo.CreateGoods(ctx, &domain.Goods{
			CategoryID:      r.CategoryID,
			BrandsID:        r.BrandsID,
			TypeID:          r.TypeID,
			Name:            r.Name,
			NameAlias:       r.NameAlias,
			GoodsSn:         r.GoodsSn,
			GoodsTags:       r.GoodsTags,
			MarketPrice:     r.MarketPrice,
			GoodsBrief:      r.GoodsBrief,
			GoodsFrontImage: r.GoodsFrontImage,
			GoodsImages:     r.GoodsImages,
			OnSale:          r.OnSale,
			IsNew:           r.IsNew,
			IsHot:           r.IsHot,
			ShipFree:        r.ShipFree,
			ShipID:          r.ShipID,
		})
		if err != nil {
			return err
		}
		// 更新商品 SKU 表
		for _, v := range r.Sku {
			res := &domain.GoodsSku{
				GoodsID:        goods.ID,
				GoodsSn:        goods.GoodsSn,
				GoodsName:      goods.Name,
				SkuName:        v.SkuName,
				SkuCode:        v.SkuCode,
				BarCode:        v.BarCode,
				Price:          v.Price,
				PromotionPrice: v.PromotionPrice,
				Points:         v.Points,
				RemarksInfo:    v.RemarksInfo,
				Pic:            v.Pic,
				Inventory:      v.Inventory,
				OnSale:         v.OnSale,
			}

			goodsAttr, err := json.Marshal(v.GroupAttr)
			if err != nil {
				return err
			}
			res.AttrInfo = string(goodsAttr)

			// 插入 sku 表
			skuInfo, err := g.skuRepo.Create(ctx, res)
			if err != nil {
				return err
			}

			// 插入库存表
			_, err = g.inventoryRepo.Create(ctx, &domain.Inventory{
				SkuID:     skuInfo.ID,
				Inventory: skuInfo.Inventory,
			})
			if err != nil {
				return err
			}
			// 插入 sku 规格关联关系表
			var skuRelation []*domain.GoodsSpecificationSku
			for _, spec := range v.Specification {
				skuRelation = append(skuRelation, &domain.GoodsSpecificationSku{
					SkuID:           skuInfo.ID,
					SkuCode:         skuInfo.SkuCode,
					SpecificationId: spec.SpecificationID,
					ValueId:         spec.SpecificationValueID,
				})
			}

			// 插入商品规格关联关系表
			err = g.skuRepo.CreateSkuRelation(ctx, skuRelation)
			if err != nil {
				return err
			}

			// esModel
			{
				EsGoods.Sku = append(EsGoods.Sku, domain.EsSku{
					SkuID:    skuInfo.ID,
					SkuName:  skuInfo.SkuName,
					SkuPrice: skuInfo.Price,
				})
				EsGoods.BrandsID = brand.ID
				EsGoods.BrandName = brand.Name
				EsGoods.CategoryID = cate.ID
				EsGoods.CategoryName = cate.Name
				EsGoods.TypeID = goodsType.ID
				EsGoods.TypeName = goodsType.Name
				EsGoods.Name = goodsType.Name
				EsGoods.ID = goods.ID
				EsGoods.OnSale = goods.OnSale
				EsGoods.ShipFree = goods.ShipFree
				EsGoods.IsNew = goods.IsNew
				EsGoods.IsHot = goods.IsHot
				EsGoods.Name = goods.Name
				EsGoods.GoodsTags = goods.GoodsTags
				EsGoods.ClickNum = goods.ClickNum
				EsGoods.SoldNum = goods.SoldNum
				EsGoods.FavNum = goods.FavNum
				EsGoods.MarketPrice = goods.MarketPrice
				EsGoods.GoodsBrief = goods.GoodsBrief

			}
			fmt.Println("EsGoods", EsGoods)
			// 插入 EsGoods
			err = g.esGoodsRepo.InsertEsGoods(ctx, EsGoods)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &domain.GoodsInfoResponse{GoodsID: goods.ID}, nil
}
