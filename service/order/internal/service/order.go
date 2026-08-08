package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	v1 "order/api/order/v1"
	"order/internal/biz"
	"order/internal/conf"
	"order/internal/domain"
	"order/internal/pkg/mq"
)

type OrderService struct {
	v1.UnimplementedOrderServer

	oc  *biz.OrderUsecase
	log *log.Helper

	mqAddr string
}

func NewOrderService(o *biz.OrderUsecase, mqConf *conf.Mq, logger log.Logger) *OrderService {
	s := &OrderService{oc: o, log: log.NewHelper(logger)}
	if mqConf != nil && mqConf.Addr != "" {
		s.mqAddr = mqConf.Addr
		go s.startPaymentConsumer()
	}
	return s
}

func (o *OrderService) CreateOrder(ctx context.Context, r *v1.OrderRequest) (*v1.OrderInfoResponse, error) {
	var cartItem []*domain.CartItem
	for _, cart := range r.CartItem {
		res := &domain.CartItem{
			CartId:   cart.CartId,
			SkuId:    cart.SkuId,
			SkuPrice: cart.SkuPrice,
			SkuNum:   cart.SkuNum,
		}
		cartItem = append(cartItem, res)
	}

	order, err := o.oc.CreateOrder(ctx, &domain.CreateOrder{
		UserId:    r.UserId,
		AddressId: r.Address,
		CartItem:  cartItem,
	})
	if err != nil {
		return nil, err
	}
	return toOrderInfo(order), nil
}

func (o *OrderService) CancelOrder(ctx context.Context, r *v1.CancelOrderRequest) (*v1.CheckResponse, error) {
	if err := o.oc.CancelOrder(ctx, r.UserId, r.OrderSn); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (o *OrderService) GetOrder(ctx context.Context, r *v1.GetOrderRequest) (*v1.OrderInfoResponse, error) {
	order, err := o.oc.GetOrder(ctx, r.UserId, r.OrderSn)
	if err != nil {
		return nil, err
	}
	return toOrderInfo(order), nil
}

func (o *OrderService) ListOrders(ctx context.Context, r *v1.ListOrderRequest) (*v1.ListOrderReply, error) {
	list, total, err := o.oc.ListOrders(ctx, r.UserId, int(r.Page), int(r.PageSize))
	if err != nil {
		return nil, err
	}
	reply := &v1.ListOrderReply{Total: int32(total)}
	for _, order := range list {
		reply.List = append(reply.List, toOrderInfo(order))
	}
	return reply, nil
}

func toOrderInfo(order *domain.Order) *v1.OrderInfoResponse {
	return &v1.OrderInfoResponse{
		Id:      order.ID,
		UserId:  order.User,
		OrderSn: order.OrderSn,
		Status:  orderStatusText(order.OrderStatus),
		Post:    order.Post,
		Total:   order.OrderAmount,
		Address: order.Address,
		Name:    order.SignerName,
		Mobile:  order.SingerMobile,
		AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func orderStatusText(status int) string {
	switch status {
	case 1:
		return "待支付"
	case 2:
		return "已支付"
	case 3:
		return "已发货"
	case 4:
		return "已签收"
	case 5:
		return "已取消"
	case 6:
		return "交易完成"
	default:
		return "未知"
	}
}

type paymentSuccessEvent struct {
	PaymentNo string `json:"payment_no"`
	OrderSn   string `json:"order_sn"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
	TradeNo   string `json:"trade_no"`
}

func (o *OrderService) startPaymentConsumer() {
	for {
		consumer, err := mq.NewConsumer(
			o.mqAddr,
			"payment.exchange",
			"q.payment.success",
			[]string{"payment.success"},
			o.handlePaymentSuccess,
		)
		if err != nil {
			o.log.Errorf("payment consumer init failed: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if err := consumer.Run(context.Background()); err != nil {
			o.log.Errorf("payment consumer stopped: %v", err)
		}
		consumer.Close()
		time.Sleep(5 * time.Second)
	}
}

func (o *OrderService) handlePaymentSuccess(ctx context.Context, body []byte) error {
	var evt paymentSuccessEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}
	if err := o.oc.MarkPaid(ctx, evt.OrderSn); err != nil {
		o.log.Errorf("mark paid failed: order=%s err=%v", evt.OrderSn, err)
		return err
	}
	o.log.Infof("order paid: %s trade=%s", evt.OrderSn, evt.TradeNo)
	return nil
}
