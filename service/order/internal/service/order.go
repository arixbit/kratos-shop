package service

import (
	"context"
	"encoding/json"
	"time"

	kerrors "github.com/go-kratos/kratos/v2/errors"
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
		go s.startInventoryConsumer()
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

func (o *OrderService) AdminListOrders(ctx context.Context, r *v1.AdminListOrderRequest) (*v1.ListOrderReply, error) {
	list, total, err := o.oc.AdminListOrders(ctx, int(r.Page), int(r.PageSize), int(r.Status))
	if err != nil {
		return nil, err
	}
	reply := &v1.ListOrderReply{Total: int32(total)}
	for _, order := range list {
		reply.List = append(reply.List, toOrderInfo(order))
	}
	return reply, nil
}

func (o *OrderService) AdminGetOrder(ctx context.Context, r *v1.AdminGetOrderRequest) (*v1.OrderInfoResponse, error) {
	order, err := o.oc.AdminGetOrder(ctx, r.OrderSn)
	if err != nil {
		return nil, err
	}
	return toOrderInfo(order), nil
}

func (o *OrderService) ShipOrder(ctx context.Context, r *v1.ShipOrderRequest) (*v1.CheckResponse, error) {
	if err := o.oc.ShipOrder(ctx, r.OrderSn, r.Post); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (o *OrderService) RefundOrder(ctx context.Context, r *v1.RefundOrderRequest) (*v1.CheckResponse, error) {
	if err := o.oc.RefundOrder(ctx, r.OrderSn); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (o *OrderService) DashboardStats(ctx context.Context, r *v1.DashboardStatsRequest) (*v1.DashboardStatsReply, error) {
	stats, err := o.oc.DashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	reply := &v1.DashboardStatsReply{
		TotalOrders: stats.TotalOrders,
		TotalSales:  stats.TotalSales,
		TodayOrders: stats.TodayOrders,
		TodaySales:  stats.TodaySales,
	}
	for _, item := range stats.StatusCounts {
		reply.StatusCounts = append(reply.StatusCounts, &v1.StatusCount{
			Status: item.Status,
			Count:  item.Count,
		})
	}
	for _, item := range stats.Last30Days {
		reply.Last30Days = append(reply.Last30Days, &v1.DailySales{
			Date:       item.Date,
			OrderCount: item.OrderCount,
			Amount:     item.Amount,
		})
	}
	for _, item := range stats.TopGoods {
		reply.TopGoods = append(reply.TopGoods, &v1.TopGoods{
			SkuId:   item.SkuID,
			SkuName: item.SkuName,
			Num:     item.Num,
			Amount:  item.Amount,
		})
	}
	return reply, nil
}

func toOrderInfo(order *domain.Order) *v1.OrderInfoResponse {
	resp := &v1.OrderInfoResponse{
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
	for _, item := range order.Items {
		resp.Goods = append(resp.Goods, &v1.OrderGoodsInfo{
			Id:         item.ID,
			SkuId:      item.SkuId,
			SkuName:    item.SkuName,
			SkuPrice:   item.SkuPrice,
			Num:        item.Num,
			TotalPrice: item.TotalPrice,
		})
	}
	return resp
}

func orderStatusText(status int) string {
	switch status {
	case domain.OrderStatusInventoryPending:
		return "库存处理中"
	case domain.OrderStatusPendingPayment:
		return "待支付"
	case domain.OrderStatusPaid:
		return "已支付"
	case domain.OrderStatusShipped:
		return "已发货"
	case domain.OrderStatusSigned:
		return "已签收"
	case domain.OrderStatusCancelled:
		return "已取消"
	case domain.OrderStatusCompleted:
		return "交易完成"
	case domain.OrderStatusRefunded:
		return "已退款"
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

type inventoryResultEvent struct {
	OrderSn string `json:"order_sn"`
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

func (o *OrderService) startInventoryConsumer() {
	for {
		consumer, err := mq.NewConsumer(
			o.mqAddr,
			"inventory.exchange",
			"q.inventory.result",
			[]string{"inventory.locked", "inventory.lock.failed"},
			o.handleInventoryResult,
		)
		if err != nil {
			o.log.Errorf("inventory consumer init failed: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if err := consumer.Run(context.Background()); err != nil {
			o.log.Errorf("inventory consumer stopped: %v", err)
		}
		consumer.Close()
		time.Sleep(5 * time.Second)
	}
}

func (o *OrderService) handleInventoryResult(ctx context.Context, body []byte) error {
	var evt inventoryResultEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}
	if evt.OrderSn == "" {
		return kerrors.New(400, "INVENTORY_RESULT_INVALID", "库存结果缺少订单号")
	}
	if evt.Success {
		if err := o.oc.MarkInventoryLocked(ctx, evt.OrderSn); err != nil {
			return err
		}
		o.log.Infof("inventory locked: order=%s", evt.OrderSn)
		return nil
	}
	if err := o.oc.MarkInventoryFailed(ctx, evt.OrderSn); err != nil {
		return err
	}
	o.log.Warnf("inventory lock failed: order=%s reason=%s", evt.OrderSn, evt.Reason)
	return nil
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
