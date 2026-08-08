package data

import (
	"cart/internal/domain"
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"cart/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type ShopCart struct {
	ID         int64          `gorm:"primarykey;type:bigint" json:"id"`
	UserId     int64          `gorm:"type:bigint;not null;comment:用户id" json:"user_id"`
	GoodsId    int64          `gorm:"type:bigint;not null;comment:商品id" json:"goods_id"`
	SkuId      int64          `gorm:"type:bigint;not null;comment:sku_id" json:"sku_id"`
	GoodsPrice int64          `gorm:"type:bigint;not null;comment:商品价格" json:"goods_price"`
	GoodsNum   int32          `gorm:"type:int;not null;comment:商品数量" json:"goods_num"`
	GoodsSn    string         `gorm:"type:varchar(500);default:;comment:商品编号"`
	GoodsName  string         `gorm:"type:varchar(500);default:;comment:商品名称"`
	IsSelect   bool           `gorm:"type:boolean;comment:是否选中;default:false" json:"is_select"`
	CreatedAt  time.Time      `gorm:"column:add_time" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:update_time" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

type cartRepo struct {
	data *Data
	log  *log.Helper
}

func NewCartRepo(data *Data, logger log.Logger) biz.CartRepo {
	return &cartRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (p *ShopCart) ToDomain() *domain.ShopCart {
	return &domain.ShopCart{
		ID:         p.ID,
		UserId:     p.UserId,
		GoodsId:    p.GoodsId,
		SkuId:      p.SkuId,
		GoodsPrice: p.GoodsPrice,
		GoodsNum:   p.GoodsNum,
		GoodsSn:    p.GoodsSn,
		GoodsName:  p.GoodsName,
		IsSelect:   p.IsSelect,
	}
}

func (r *cartRepo) Create(ctx context.Context, c *domain.ShopCart) (*domain.ShopCart, error) {
	var shopCart ShopCart
	if result := r.data.db.Where(&ShopCart{UserId: c.UserId, SkuId: c.SkuId}).First(&shopCart); result.RowsAffected == 1 {
		shopCart.GoodsNum += c.GoodsNum
	} else {
		shopCart.UserId = c.UserId
		shopCart.GoodsId = c.GoodsId
		shopCart.SkuId = c.SkuId
		shopCart.GoodsPrice = c.GoodsPrice
		shopCart.GoodsNum = c.GoodsNum
		shopCart.GoodsSn = c.GoodsSn
		shopCart.GoodsName = c.GoodsName
		shopCart.IsSelect = c.IsSelect
	}

	if result := r.data.db.Save(&shopCart); result.Error != nil {
		return nil, errors.InternalServer("CREATE_CART_NOT_FOUND", "创建购物车失败")
	}

	return shopCart.ToDomain(), nil
}

func (r *cartRepo) List(ctx context.Context, userId int64) (domain.ShopCartList, error) {
	var shopCarts []ShopCart
	if result := r.data.db.Where(&ShopCart{UserId: userId}).Find(&shopCarts); result.Error != nil {
		return nil, errors.InternalServer("SELECT_CART_ERROR", "用户购物车列表查询失败")
	} else {
		var rsp domain.ShopCartList

		for _, cart := range shopCarts {
			rsp = append(rsp, cart.ToDomain())
		}
		return rsp, nil
	}
}

func (r *cartRepo) Update(ctx context.Context, c *domain.ShopCart) error {
	var cart ShopCart
	if err := r.data.db.Where("id = ? AND user_id = ?", c.ID, c.UserId).First(&cart).Error; err != nil {
		return errors.NotFound("CART_NOT_FOUND", "购物车记录不存在")
	}
	if c.GoodsNum <= 0 {
		return errors.New(400, "CART_NUM_INVALID", "商品数量必须大于 0")
	}
	cart.GoodsNum = c.GoodsNum
	if err := r.data.db.Save(&cart).Error; err != nil {
		return errors.InternalServer("UPDATE_CART_ERROR", "购物车更新失败")
	}
	return nil
}

func (r *cartRepo) Delete(ctx context.Context, id, userId int64) error {
	result := r.data.db.Where("id = ? AND user_id = ?", id, userId).Delete(&ShopCart{})
	if result.Error != nil {
		return errors.InternalServer("DELETE_CART_ERROR", "购物车删除失败")
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("CART_NOT_FOUND", "购物车记录不存在")
	}
	return nil
}

func (r *cartRepo) Get(ctx context.Context, id, userId int64) (*domain.ShopCart, error) {
	var cart ShopCart
	if err := r.data.db.Where("id = ? AND user_id = ?", id, userId).First(&cart).Error; err != nil {
		return nil, errors.NotFound("CART_NOT_FOUND", "购物车记录不存在")
	}
	return cart.ToDomain(), nil
}
