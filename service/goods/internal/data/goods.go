package data

import (
	"context"
	serrors "errors"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"goods/internal/biz"
	"goods/internal/domain"
	"gorm.io/gorm"
)

// Goods 商品表
type Goods struct {
	BaseFields
	CategoryID int32 `gorm:"index:category_id;type:int;comment:分类ID;not null"`
	BrandsID   int32 `gorm:"index:brand_id;type:int;comment:品牌ID ;not null"`
	TypeID     int64 `gorm:"index:type_id;type:int;comment:商品类型ID ;not null"`

	Name            string   `gorm:"type:varchar(100);not null;comment:商品名称"`
	NameAlias       string   `gorm:"type:varchar(100);not null;comment:商品别名"`
	GoodsSn         string   `gorm:"type:varchar(100);not null;comment:商品编号"`
	GoodsTags       string   `gorm:"type:varchar(100);not null;comment:商品标签"`
	MarketPrice     int64    `gorm:"type:int;default:0;not null;comment:商品展示价格"`
	GoodsBrief      string   `gorm:"type:varchar(100);not null;comment:商品简介"`
	GoodsFrontImage string   `gorm:"type:varchar(200);not null;comment:商品封面图"`
	GoodsImages     GormList `gorm:"type:text;not null;comment:商品的介绍图"` // 切片类型转为 json 到数据库，取出来是切片类型

	OnSale   bool  `gorm:"default:false;comment:是否上架;not null "`
	ShipFree bool  `gorm:"default:false;comment:是否免运费; not null"`
	ShipID   int32 `gorm:"type:int;comment:运费模版ID;not null"`
	IsNew    bool  `gorm:"default:false;comment:是否新品;not null"`
	IsHot    bool  `gorm:"comment:是否热卖商品;default:false;not null"`

	ClickNum int64 `gorm:"default:0;type:bigint;comment:商品详情点击数"`
	SoldNum  int64 `gorm:"default:0;type:bigint;comment:商品销售数"`
	FavNum   int64 `gorm:"default:0;type:bigint;comment:商品收藏数"`

	// 售前服务信息、售后服务信息、商品促销活动信息
}

type ConsumedEvent struct {
	EventID    string    `gorm:"column:event_id;type:varchar(64);primaryKey"`
	OrderSn    string    `gorm:"column:order_sn;type:varchar(64);not null;default:''"`
	ConsumedAt time.Time `gorm:"column:consumed_at"`
}

func (ConsumedEvent) TableName() string {
	return "consumed_event"
}

type goodsRepo struct {
	data *Data
	log  *log.Helper
}

// NewGoodsRepo .
func NewGoodsRepo(data *Data, logger log.Logger) biz.GoodsRepo {
	return &goodsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (p *Goods) ToDomain() *domain.Goods {
	return &domain.Goods{
		ID:              p.ID,
		CategoryID:      p.CategoryID,
		BrandsID:        p.BrandsID,
		TypeID:          p.TypeID,
		Name:            p.Name,
		NameAlias:       p.NameAlias,
		GoodsSn:         p.GoodsSn,
		GoodsTags:       p.GoodsTags,
		MarketPrice:     p.MarketPrice,
		GoodsBrief:      p.GoodsBrief,
		GoodsFrontImage: p.GoodsFrontImage,
		GoodsImages:     p.GoodsImages,
		OnSale:          p.OnSale,
		ShipFree:        p.ShipFree,
		ShipID:          p.ShipID,
		IsNew:           p.IsNew,
		IsHot:           p.IsHot,
		ClickNum:        p.ClickNum,
		SoldNum:         p.SoldNum,
		FavNum:          p.FavNum,
	}
}

func (g goodsRepo) CreateGoods(c context.Context, goods *domain.Goods) (*domain.Goods, error) {
	product := &Goods{
		CategoryID:      goods.CategoryID,
		BrandsID:        goods.BrandsID,
		TypeID:          goods.TypeID,
		Name:            goods.Name,
		NameAlias:       goods.NameAlias,
		GoodsSn:         goods.GoodsSn,
		GoodsTags:       goods.GoodsTags,
		MarketPrice:     goods.MarketPrice,
		GoodsBrief:      goods.GoodsBrief,
		GoodsFrontImage: goods.GoodsFrontImage,
		GoodsImages:     goods.GoodsImages,
		OnSale:          goods.OnSale,
		ShipFree:        goods.ShipFree,
		ShipID:          goods.ShipID,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
	}

	result := g.data.DB(c).Save(product)
	if result.Error != nil {
		return nil, errors.InternalServer("GOODS_CREATE_ERROR", "商品创建失败")
	}
	return product.ToDomain(), nil
}

func (g goodsRepo) GoodsListByIDs(c context.Context, ids ...int64) ([]*domain.Goods, error) {
	var l []*Goods
	if err := g.data.DB(c).Where("id IN (?)", ids).Find(&l).Error; err != nil {
		return nil, errors.NotFound("GOODS_NOT_FOUND", "商品不存在")
	}
	var res []*domain.Goods
	for _, item := range l {
		res = append(res, item.ToDomain())
	}
	return res, nil
}

func (g goodsRepo) Update(ctx context.Context, goods *domain.Goods) error {
	updates := map[string]interface{}{}
	if goods.Name != "" {
		updates["name"] = goods.Name
	}
	if goods.NameAlias != "" {
		updates["name_alias"] = goods.NameAlias
	}
	if goods.GoodsBrief != "" {
		updates["goods_brief"] = goods.GoodsBrief
	}
	if goods.GoodsFrontImage != "" {
		updates["goods_front_image"] = goods.GoodsFrontImage
	}
	if goods.MarketPrice > 0 {
		updates["market_price"] = goods.MarketPrice
	}
	updates["ship_free"] = goods.ShipFree
	updates["is_new"] = goods.IsNew
	updates["is_hot"] = goods.IsHot
	updates["update_time"] = time.Now()
	if len(updates) == 0 {
		return nil
	}
	return g.data.DB(ctx).Model(&Goods{}).Where("id = ?", goods.ID).Updates(updates).Error
}

func (g goodsRepo) Delete(ctx context.Context, id int64) error {
	return g.data.DB(ctx).Delete(&Goods{}, id).Error
}

func (g goodsRepo) UpdateStatus(ctx context.Context, id int64, onSale bool) error {
	return g.data.DB(ctx).Model(&Goods{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"on_sale":     onSale,
			"update_time": time.Now(),
		}).Error
}

func (g goodsRepo) ListPage(ctx context.Context, keywords string, categoryID, brandID int32, page, pageSize int) ([]*domain.Goods, int64, error) {
	countQuery := g.data.DB(ctx).Model(&Goods{})
	listQuery := g.data.DB(ctx).Model(&Goods{})
	if keywords != "" {
		like := "%" + keywords + "%"
		countQuery = countQuery.Where("name ILIKE ? OR name_alias ILIKE ? OR goods_sn ILIKE ?", like, like, like)
		listQuery = listQuery.Where("name ILIKE ? OR name_alias ILIKE ? OR goods_sn ILIKE ?", like, like, like)
	}
	if categoryID > 0 {
		countQuery = countQuery.Where("category_id = ?", categoryID)
		listQuery = listQuery.Where("category_id = ?", categoryID)
	}
	if brandID > 0 {
		countQuery = countQuery.Where("brands_id = ?", brandID)
		listQuery = listQuery.Where("brands_id = ?", brandID)
	}
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	var list []Goods
	if err := listQuery.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	res := make([]*domain.Goods, 0, len(list))
	for i := range list {
		res = append(res, list[i].ToDomain())
	}
	return res, total, nil
}

func (g goodsRepo) GetByID(ctx context.Context, id int64) (*domain.Goods, error) {
	var goods Goods
	if err := g.data.DB(ctx).Where("id = ?", id).First(&goods).Error; err != nil {
		return nil, err
	}
	return goods.ToDomain(), nil
}

func (g goodsRepo) ListAll(ctx context.Context) ([]*domain.Goods, error) {
	var list []Goods
	if err := g.data.DB(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.Goods, 0, len(list))
	for i := range list {
		res = append(res, list[i].ToDomain())
	}
	return res, nil
}

func (g goodsRepo) ListImagesByGoodsID(ctx context.Context, goodsID int64) ([]*domain.GoodsImage, error) {
	var list []GoodsImages
	if err := g.data.DB(ctx).Where("goods_id = ?", goodsID).Order("position ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.GoodsImage, 0, len(list))
	for i := range list {
		res = append(res, &domain.GoodsImage{
			ID:       list[i].ID,
			GoodsID:  list[i].GoodsID,
			Link:     list[i].Link,
			Position: list[i].Position,
			IsMaster: list[i].IsMaster,
		})
	}
	return res, nil
}

func (g goodsRepo) IncrSoldNum(ctx context.Context, skuItems map[int64]int32) error {
	for skuID, num := range skuItems {
		var sku GoodsSku
		if err := g.data.DB(ctx).Where("id = ?", skuID).First(&sku).Error; err != nil {
			if serrors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if err := g.data.DB(ctx).Model(&Goods{}).Where("id = ?", sku.GoodsID).
			UpdateColumn("sold_num", gorm.Expr("sold_num + ?", num)).Error; err != nil {
			return err
		}
	}
	return nil
}

func (g goodsRepo) IsConsumed(ctx context.Context, eventID string) (bool, error) {
	var evt ConsumedEvent
	err := g.data.DB(ctx).Where("event_id = ?", eventID).First(&evt).Error
	if serrors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (g goodsRepo) MarkConsumed(ctx context.Context, eventID, orderSn string) error {
	return g.data.DB(ctx).Create(&ConsumedEvent{
		EventID:    eventID,
		OrderSn:    orderSn,
		ConsumedAt: time.Now(),
	}).Error
}
