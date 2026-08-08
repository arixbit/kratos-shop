package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"goods/internal/biz"
	"goods/internal/domain"
)

// GoodsImages goods images
type GoodsImages struct {
	BaseFields
	GoodsID int64 `gorm:"index:goods_id;type:int;comment:商品ID;not null"`
	Goods   Goods

	Link     string `gorm:"type:varchar(200);comment:商品图片URL地址;not null"`
	Position int32  `gorm:"type:smallint;comment:商品图片位置;not null"`
	IsMaster bool   `gorm:"comment:是否主图: 1是,0否;default:false;not null"`
}

type goodsImagesRepo struct {
	data *Data
	log  *log.Helper
}

// NewGoodsImagesRepo .
func NewGoodsImagesRepo(data *Data, logger log.Logger) biz.GoodsImagesRepo {
	return &goodsImagesRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *goodsImagesRepo) CreateGreeter(ctx context.Context, g *domain.Goods) error {
	return nil
}

func (r *goodsImagesRepo) UpdateGreeter(ctx context.Context, g *domain.Goods) error {
	return nil
}
