package domain

import (
	"time"
)

type Order struct {
	ID            int64
	User          int64
	OrderSn       string
	OrderAmount   int64
	GoodsAmount   int64
	OrderStatus   int
	ExpressAmount int64
	DeliveryAt    time.Time
	RefundTime    time.Time
	Post          string
	Address       string
	SignerName    string
	SingerMobile  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OrderAddress struct {
	ID              int64
	User            int64
	OrderSn         string
	RecipientName   string
	RecipientMobile string
	Province        string
	City            string
	Districts       string
	Address         string
	PostCode        string
}

type OrderGoods struct {
	ID         int64
	OrderSn    string
	UserId     int64
	SkuId      int64
	SkuName    string
	SkuPrice   int64
	Num        int32
	TotalPrice int64
}

type OutboxEvent struct {
	ID         int64
	EventID    string
	EventType  string
	OrderSn    string
	Payload    []byte
	Status     int
	RetryCount int
}

type CreateOrder struct {
	UserId    int64
	AddressId int64
	CartItem  CartItemList
}

type CartItem struct {
	CartId   int64
	SkuId    int64
	SkuPrice int64
	SkuNum   int32
}

type CartItemList []*CartItem

func (p CartItemList) FindById(id int64) *CartItem {
	for _, item := range p {
		if item.CartId == id {
			return item
		}
	}
	return nil
}

func (p CartItemList) GetSkuId() []int64 {
	var l []int64
	for _, item := range p {
		l = append(l, item.SkuId)
	}
	return l
}
