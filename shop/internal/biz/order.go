package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	orderV1 "shop/api/service/order/v1"
	v1 "shop/api/shop/v1"
)

type OrderUsecase struct {
	oc  orderV1.OrderClient
	log *log.Helper
}

func NewOrderUsecase(oc orderV1.OrderClient, logger log.Logger) *OrderUsecase {
	return &OrderUsecase{oc: oc, log: log.NewHelper(log.With(logger, "module", "usecase/order"))}
}

func (uc *OrderUsecase) Create(ctx context.Context, req *v1.OrderCreateRequest) (*v1.OrderInfoReply, error) {
	uid, err := getUid(ctx)
	if err != nil {
		return nil, err
	}
	var items []*orderV1.CartItem
	for _, item := range req.CartItem {
		items = append(items, &orderV1.CartItem{
			CartId: item.CartId,
			SkuId:  item.SkuId,
			SkuNum: item.SkuNum,
		})
	}
	resp, err := uc.oc.CreateOrder(ctx, &orderV1.OrderRequest{
		UserId:   uid,
		Address:  req.Address,
		CartItem: items,
	})
	if err != nil {
		return nil, err
	}
	return toOrderInfoReply(resp), nil
}

func (uc *OrderUsecase) Cancel(ctx context.Context, req *v1.OrderCancelRequest) error {
	uid, err := getUid(ctx)
	if err != nil {
		return err
	}
	_, err = uc.oc.CancelOrder(ctx, &orderV1.CancelOrderRequest{
		UserId:  uid,
		OrderSn: req.OrderSn,
	})
	return err
}

func (uc *OrderUsecase) List(ctx context.Context, req *v1.OrderListRequest) (*v1.OrderListReply, error) {
	uid, err := getUid(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := uc.oc.ListOrders(ctx, &orderV1.ListOrderRequest{
		UserId:   uid,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	reply := &v1.OrderListReply{Total: resp.Total}
	for _, item := range resp.List {
		reply.List = append(reply.List, toOrderInfoReply(item))
	}
	return reply, nil
}

func (uc *OrderUsecase) Detail(ctx context.Context, req *v1.OrderDetailRequest) (*v1.OrderInfoReply, error) {
	uid, err := getUid(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := uc.oc.GetOrder(ctx, &orderV1.GetOrderRequest{
		UserId:  uid,
		OrderSn: req.OrderSn,
	})
	if err != nil {
		return nil, err
	}
	return toOrderInfoReply(resp), nil
}

func toOrderInfoReply(item *orderV1.OrderInfoResponse) *v1.OrderInfoReply {
	if item == nil {
		return nil
	}
	reply := &v1.OrderInfoReply{
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
		reply.Goods = append(reply.Goods, &v1.OrderGoodsReply{
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
