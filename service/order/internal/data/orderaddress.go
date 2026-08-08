package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"order/internal/biz"
	"order/internal/domain"
	"time"
)

type OrderAddress struct {
	ID      int64  `gorm:"primarykey"`
	User    int64  `gorm:"type:bigint;not null;default:0;index;comment:用户ID"`
	OrderSn string `gorm:"type:varchar(30);not null;default:'';index"` // 订单号，我们平台自己生成的订单号

	// 用户收货信息
	RecipientName   string `gorm:"type:varchar(20);not null;default:'';comment:收货姓名"`
	RecipientMobile string `gorm:"type:varchar(20);not null;default:'';comment:收货电话"`
	Province        string `gorm:"type:varchar(25);not null;default:'';comment:省"`
	City            string `gorm:"type:varchar(25);not null;default:'';comment:市"`
	Districts       string `gorm:"type:varchar(25);not null;default:'';comment:区/县"`
	Address         string `gorm:"type:varchar(255);not null;default:'';comment:收货详细地址"`
	PostCode        string `gorm:"type:varchar(25);not null;default:'';comment:邮编"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (OrderAddress) TableName() string {
	return "order_address"
}

type orderAddressRepo struct {
	data *Data
	log  *log.Helper
}

// NewOrderAddressRepo .
func NewOrderAddressRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (p *OrderAddress) ToDomain() *domain.Order {
	return &domain.Order{
		User:          p.User,
		OrderSn:       p.OrderSn,
		Address:       p.Address,
		SignerName:    p.RecipientName,
		SingerMobile:  p.RecipientMobile,
	}
}
