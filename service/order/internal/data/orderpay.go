package data

import "time"

// OrderPay 支付信息
type OrderPay struct {
	ID        int64     `gorm:"primarykey"`
	User      int64     `gorm:"type:bigint;not null;default:0;index;comment:用户ID"`
	OrderSn   string    `gorm:"type:varchar(30);not null;default:'';index"`          // 订单号，我们平台自己生成的订单号
	TradeNo   string    `gorm:"type:varchar(100);not null;default:'';comment:交易号、流水号"` // 交易号就是支付宝的订单号 查账
	PayType   string    `gorm:"type:varchar(20);not null;default:'';comment:1alipay(支付宝)，2wechat(微信)"`
	PayStatus int       `gorm:"type:int;not null;default:0;comment:1PAYING(待支付),2TRADE_SUCCESS(成功)，3TRADE_CLOSED(超时关闭),4WAIT_BUYER_PAY(交易创建),5TRADE_FINISHED(交易结束)"`
	PayTime   time.Time `gorm:"type:timestamptz;comment:支付时间"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}
