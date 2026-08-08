package data

import (
	"time"
)

type OrderGoods struct {
	ID         int64  `gorm:"primarykey;type:bigint" json:"id"`
	OrderSn    string `gorm:"type:varchar(30);not null;default:'';index"` // 订单号，我们平台自己生成的订单号
	UserId     int64  `gorm:"type:bigint;not null;default:0;index"`
	SkuId      int64  `gorm:"type:bigint;not null;default:0;index"`
	SkuName    string `gorm:"type:varchar(100);not null;default:'';index"`
	SkuPrice   int64  `gorm:"type:bigint;not null;default:0;comment:生成订单时的商品价格(单)"`
	Num        int32  `gorm:"type:int;not null;default:0;comment:商品数量"`
	TotalPrice int64  `gorm:"type:bigint;not null;default:0;comment:生成订单时的商品总价"`
	// 商品快照信息根据需求后期添加
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderGoods) TableName() string {
	return "order_goods"
}
