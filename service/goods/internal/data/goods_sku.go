package data

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/net/context"
	"goods/internal/biz"
	"goods/internal/domain"
)

// GoodsSku 商品SKU 表
type GoodsSku struct {
	BaseFields
	GoodsID        int64  `gorm:"index:goods_id;type:int;comment:商品ID;not null"`
	GoodsSn        string `gorm:"type:varchar(100);not null;comment:商品编号"`
	GoodsName      string `gorm:"type:varchar(100);not null;comment:商品名称"`
	SkuName        string `gorm:"type:varchar(100);comment:SKU名称;not null"`
	SkuCode        string `gorm:"type:varchar(100);comment:SKUCode;not null"`
	BarCode        string `gorm:"type:varchar(100);comment:条码;not null"`
	Price          int64  `gorm:"type:int;comment:商品售价;not null"`
	PromotionPrice int64  `gorm:"type:int;comment:商品促销售价;not null"`
	Points         int64  `gorm:"type:int;comment:赠送积分;not null"`
	RemarksInfo    string `gorm:"type:varchar(100);comment:备注信息;not null"`
	Pic            string `gorm:"type:varchar(500);not null;comment:规格参数对应的图片" json:"pic"`
	OnSale         bool   `gorm:"comment:是否上架;default:false;not null"`
	AttrInfo       string `gorm:"type:varchar(2000);comment:商品属性信息JSON;not null"`
	Inventory      int64  `gorm:"type:int;comment:商品SKU库存冗余字段;not null"`
}

// GoodsSpecificationSku 商品规格和商品Sku关联表
type GoodsSpecificationSku struct {
	BaseFields
	SkuID           int64  `gorm:"index:sku_id;type:int;comment:商品SKU_ID;not null"`
	SkuCode         string `gorm:"type:varchar(100);comment:商品SKU_Code;not null"`
	SpecificationId int64  `gorm:"index:specification_id;type:int;comment:商品规格ID;not null"`
	ValueId         int64  `gorm:"index:value_id;type:int;comment:商品规格值表ID;not null"`
}

type goodsSkuRepo struct {
	data *Data
	log  *log.Helper
}

// NewGoodsSkuRepoRepo .
func NewGoodsSkuRepoRepo(data *Data, logger log.Logger) biz.GoodsSkuRepo {
	return &goodsSkuRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (p *GoodsSku) ToDomain() *domain.GoodsSku {
	return &domain.GoodsSku{
		ID:             p.ID,
		GoodsID:        p.GoodsID,
		GoodsSn:        p.GoodsSn,
		GoodsName:      p.GoodsName,
		SkuName:        p.SkuName,
		SkuCode:        p.SkuCode,
		BarCode:        p.BarCode,
		Price:          p.Price,
		PromotionPrice: p.PromotionPrice,
		Points:         p.Points,
		RemarksInfo:    p.RemarksInfo,
		Pic:            p.Pic,
		Inventory:      p.Inventory,
		OnSale:         p.OnSale,
		AttrInfo:       p.AttrInfo,
	}
}

func (g *goodsSkuRepo) Create(ctx context.Context, req *domain.GoodsSku) (*domain.GoodsSku, error) {
	sku := &GoodsSku{
		GoodsID:        req.GoodsID,
		GoodsSn:        req.GoodsSn,
		GoodsName:      req.GoodsName,
		SkuName:        req.SkuName,
		SkuCode:        req.SkuCode,
		BarCode:        req.BarCode,
		Price:          req.Price,
		PromotionPrice: req.PromotionPrice,
		Points:         req.Points,
		RemarksInfo:    req.RemarksInfo,
		Pic:            req.Pic,
		OnSale:         req.OnSale,
		AttrInfo:       req.AttrInfo,
		Inventory:      req.Inventory,
	}

	if err := g.data.DB(ctx).Save(sku).Error; err != nil {
		return nil, errors.InternalServer("SKU_SAVE_ERROR", err.Error())
	}
	return sku.ToDomain(), nil
}

func (g *goodsSkuRepo) CreateSkuRelation(ctx context.Context, req []*domain.GoodsSpecificationSku) error {
	var info []*GoodsSpecificationSku
	for _, sku := range req {
		i := GoodsSpecificationSku{
			SkuID:           sku.SkuID,
			SkuCode:         sku.SkuCode,
			SpecificationId: sku.SpecificationId,
			ValueId:         sku.ValueId,
		}
		info = append(info, &i)
	}
	if err := g.data.DB(ctx).Table("goods_specification_skus").Save(&info).Error; err != nil {
		return errors.InternalServer("SKU_RELATION_SAVE_ERROR", err.Error())
	}
	return nil
}

func (g *goodsSkuRepo) ListByIds(ctx context.Context, ids []int64) ([]*domain.GoodsSku, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var skus []GoodsSku
	if err := g.data.DB(ctx).Where("id IN (?)", ids).Find(&skus).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.GoodsSku, 0, len(skus))
	for i := range skus {
		res = append(res, skus[i].ToDomain())
	}
	return res, nil
}

func (g *goodsSkuRepo) ListByGoodsID(ctx context.Context, goodsID int64) ([]*domain.GoodsSku, error) {
	var skus []GoodsSku
	if err := g.data.DB(ctx).Where("goods_id = ?", goodsID).Find(&skus).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.GoodsSku, 0, len(skus))
	for i := range skus {
		res = append(res, skus[i].ToDomain())
	}
	return res, nil
}
