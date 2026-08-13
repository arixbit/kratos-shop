package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "admin/api/admin/v1"
	orderV1 "admin/api/service/order/v1"
)

type OrderUsecase struct {
	oc  orderV1.OrderClient
	log *log.Helper
}

func NewOrderUsecase(oc orderV1.OrderClient, logger log.Logger) *OrderUsecase {
	return &OrderUsecase{oc: oc, log: log.NewHelper(log.With(logger, "module", "usecase/order"))}
}

func (uc *OrderUsecase) List(ctx context.Context, req *v1.OrderListRequest) (*v1.OrderListReply, error) {
	resp, err := uc.oc.AdminListOrders(ctx, &orderV1.AdminListOrderRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	reply := &v1.OrderListReply{Total: resp.Total}
	for _, item := range resp.List {
		reply.List = append(reply.List, toOrderInfo(item))
	}
	return reply, nil
}

func (uc *OrderUsecase) Detail(ctx context.Context, req *v1.OrderDetailRequest) (*v1.OrderInfo, error) {
	resp, err := uc.oc.AdminGetOrder(ctx, &orderV1.AdminGetOrderRequest{OrderSn: req.OrderSn})
	if err != nil {
		return nil, err
	}
	return toOrderInfo(resp), nil
}

func (uc *OrderUsecase) Ship(ctx context.Context, req *v1.ShipOrderRequest) error {
	_, err := uc.oc.ShipOrder(ctx, &orderV1.ShipOrderRequest{
		OrderSn: req.OrderSn,
		Post:    req.Post,
	})
	return err
}

func (uc *OrderUsecase) Refund(ctx context.Context, req *v1.RefundOrderRequest) error {
	_, err := uc.oc.RefundOrder(ctx, &orderV1.RefundOrderRequest{OrderSn: req.OrderSn})
	return err
}

func (uc *OrderUsecase) DashboardStats(ctx context.Context) (*v1.DashboardStatsReply, error) {
	resp, err := uc.oc.DashboardStats(ctx, &orderV1.DashboardStatsRequest{})
	if err != nil {
		return nil, err
	}
	reply := &v1.DashboardStatsReply{
		TotalOrders: resp.TotalOrders,
		TotalSales:  resp.TotalSales,
		TodayOrders: resp.TodayOrders,
		TodaySales:  resp.TodaySales,
	}
	for _, item := range resp.StatusCounts {
		reply.StatusCounts = append(reply.StatusCounts, &v1.StatusCount{
			Status: item.Status,
			Count:  item.Count,
		})
	}
	for _, item := range resp.Last30Days {
		reply.Last30Days = append(reply.Last30Days, &v1.DailySales{
			Date:       item.Date,
			OrderCount: item.OrderCount,
			Amount:     item.Amount,
		})
	}
	for _, item := range resp.TopGoods {
		reply.TopGoods = append(reply.TopGoods, &v1.TopGoods{
			SkuId:   item.SkuId,
			SkuName: item.SkuName,
			Num:     item.Num,
			Amount:  item.Amount,
		})
	}
	return reply, nil
}

func toOrderInfo(item *orderV1.OrderInfoResponse) *v1.OrderInfo {
	if item == nil {
		return nil
	}
	reply := &v1.OrderInfo{
		Id:      item.Id,
		UserId:  item.UserId,
		OrderSn: item.OrderSn,
		Status:  item.Status,
		Post:    item.Post,
		Total:   item.Total,
		Address: item.Address,
		Name:    item.Name,
		Mobile:  item.Mobile,
		AddTime: item.AddTime,
	}
	for _, goods := range item.Goods {
		reply.Goods = append(reply.Goods, &v1.OrderGoodsInfo{
			Id:         goods.Id,
			SkuId:      goods.SkuId,
			SkuName:    goods.SkuName,
			SkuPrice:   goods.SkuPrice,
			Num:        goods.Num,
			TotalPrice: goods.TotalPrice,
		})
	}
	return reply
}
