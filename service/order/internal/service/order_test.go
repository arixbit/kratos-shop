package service

import (
	"testing"
	"time"

	"order/internal/domain"
)

func TestOrderStatusText(t *testing.T) {
	cases := map[int]string{
		0:  "库存处理中",
		1:  "待支付",
		2:  "已支付",
		3:  "已发货",
		4:  "已签收",
		5:  "已取消",
		6:  "交易完成",
		99: "未知",
	}
	for status, want := range cases {
		if got := orderStatusText(status); got != want {
			t.Fatalf("orderStatusText(%d)=%q want %q", status, got, want)
		}
	}
}

func TestToOrderInfo(t *testing.T) {
	created := time.Date(2026, 8, 8, 12, 0, 0, 0, time.Local)
	order := &domain.Order{
		ID:           1,
		User:         2,
		OrderSn:      "SN-001",
		OrderAmount:  100,
		OrderStatus:  1,
		Address:      "地址",
		SignerName:   "张三",
		SingerMobile: "13800138000",
		CreatedAt:    created,
	}
	info := toOrderInfo(order)
	if info.Id != 1 || info.UserId != 2 || info.OrderSn != "SN-001" || info.Total != 100 {
		t.Fatalf("unexpected order info: %+v", info)
	}
	if info.Status != "待支付" || info.Address != "地址" || info.Name != "张三" || info.Mobile != "13800138000" {
		t.Fatalf("unexpected order info fields: %+v", info)
	}
}
