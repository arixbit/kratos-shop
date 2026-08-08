package domain

import "time"

type Payment struct {
	ID        int64
	PaymentNo string
	OrderSn   string
	UserID    int64
	Amount    int64
	Channel   string
	Status    int
	TradeNo   string
	PaidAt    *time.Time
}
