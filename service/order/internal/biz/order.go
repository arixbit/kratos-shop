package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	cartV1 "order/api/cart/v1"
	goodsV1 "order/api/goods/v1"
	userV1 "order/api/user/v1"
	"order/internal/domain"
	"order/internal/pkg/mq"
)

//go:generate mockgen -destination=../mocks/mrepo/order.go -package=mrepo . OrderRepo
type OrderRepo interface {
	Create(ctx context.Context, order *domain.Order, address *domain.OrderAddress, items []*domain.OrderGoods, outbox *domain.OutboxEvent) error
	ListPendingOutbox(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)
	MarkOutboxPublished(ctx context.Context, id int64) error
	CreateOutbox(ctx context.Context, outbox *domain.OutboxEvent) error
	UpdateStatusIf(ctx context.Context, orderSn string, from, to int) (bool, error)
	ListItemsByOrderSn(ctx context.Context, orderSn string) ([]*domain.OrderGoods, error)
	ListPendingTimeout(ctx context.Context, minutes int) ([]*domain.Order, error)
	GetDetail(ctx context.Context, userId int64, orderSn string) (*domain.Order, error)
	ListByUser(ctx context.Context, userId int64, page, pageSize int) ([]*domain.Order, int64, error)
	IncrementOutboxRetry(ctx context.Context, id int64) error
	MarkOutboxFailed(ctx context.Context, id int64) error
}

type OrderUsecase struct {
	repo     OrderRepo
	userRPC  userV1.UserClient
	cartRPC  cartV1.CartClient
	goodsRPC goodsV1.GoodsClient
	publisher *mq.Publisher
	log      *log.Helper
}

func NewOrderUsecase(repo OrderRepo, userRPC userV1.UserClient, cartRPC cartV1.CartClient, goodsRPC goodsV1.GoodsClient, publisher *mq.Publisher,
	logger log.Logger) *OrderUsecase {

	uc := &OrderUsecase{
		repo:     repo,
		userRPC:  userRPC,
		cartRPC:  cartRPC,
		goodsRPC: goodsRPC,
		publisher: publisher,
		log:      log.NewHelper(logger),
	}
	if publisher != nil {
		go uc.relayLoop()
	}
	go uc.timeoutLoop()
	return uc
}

func (oc *OrderUsecase) CreateOrder(ctx context.Context, order *domain.CreateOrder) (*domain.Order, error) {
	if order == nil || len(order.CartItem) == 0 {
		return nil, kerrors.New(400, "ORDER_ITEM_EMPTY", "订单商品不能为空")
	}

	// 跨服务查询购物车，校验本次下单条目
	cartList, err := oc.cartRPC.ListCart(ctx, &cartV1.ListCartRequest{UserId: order.UserId})
	if err != nil {
		return nil, err
	}
	cartMap := make(map[int64]*cartV1.CartInfoReply, len(cartList.Results))
	for _, cart := range cartList.Results {
		cartMap[cart.Id] = cart
	}
	for _, item := range order.CartItem {
		cart := cartMap[item.CartId]
		if cart == nil || cart.SkuId != item.SkuId {
			return nil, kerrors.New(400, "CART_ITEM_INVALID", "购物车条目与订单不一致")
		}
		if item.SkuNum <= 0 {
			return nil, kerrors.New(400, "SKU_NUM_INVALID", "商品数量必须大于 0")
		}
	}

	// 跨服务查询 SKU，使用服务端价格计算金额
	skuResp, err := oc.goodsRPC.SkuList(ctx, &goodsV1.SkuListRequest{Id: order.CartItem.GetSkuId()})
	if err != nil {
		return nil, err
	}
	skuMap := make(map[int64]*goodsV1.SkuInfo, len(skuResp.List))
	for _, sku := range skuResp.List {
		skuMap[sku.Id] = sku
	}

	var (
		goodsAmount int64
		orderGoods  []*domain.OrderGoods
	)
	for _, item := range order.CartItem {
		sku := skuMap[item.SkuId]
		if sku == nil || !sku.OnSale {
			return nil, kerrors.New(400, "SKU_NOT_FOUND", "SKU 不存在或已下架")
		}
		price := sku.Price
		if sku.PromotionPrice > 0 {
			price = sku.PromotionPrice
		}
		amount := price * int64(item.SkuNum)
		goodsAmount += amount
		orderGoods = append(orderGoods, &domain.OrderGoods{
			UserId:     order.UserId,
			SkuId:      sku.Id,
			SkuName:    sku.SkuName,
			SkuPrice:   price,
			Num:        item.SkuNum,
			TotalPrice: amount,
		})
	}

	// 跨服务查询收货地址
	address, err := oc.userRPC.GetAddress(ctx, &userV1.AddressReq{
		Id:  order.AddressId,
		Uid: order.UserId,
	})
	if err != nil {
		return nil, err
	}

	orderSn := generateOrderSn()
	od := &domain.Order{
		User:          order.UserId,
		OrderSn:       orderSn,
		GoodsAmount:   goodsAmount,
		OrderAmount:   goodsAmount,
		ExpressAmount: 0,
		OrderStatus:   1, // 待支付
		Address:       address.Address,
		SignerName:    address.Name,
		SingerMobile:  address.Mobile,
	}
	orderAddress := &domain.OrderAddress{
		User:            order.UserId,
		OrderSn:         orderSn,
		RecipientName:   address.Name,
		RecipientMobile: address.Mobile,
		Province:        address.Province,
		City:            address.City,
		Districts:       address.Districts,
		Address:         address.Address,
		PostCode:        address.PostCode,
	}
	for _, item := range orderGoods {
		item.OrderSn = orderSn
	}

	outbox := &domain.OutboxEvent{
		EventID:   generateEventID(),
		EventType: "order.created",
		OrderSn:   orderSn,
		Payload:   buildOrderEventPayload("order.created", orderSn, order.UserId, orderGoods),
	}
	if err := oc.repo.Create(ctx, od, orderAddress, orderGoods, outbox); err != nil {
		return nil, err
	}
	// 下单成功后删除已购买的购物车条目（尽力而为，不阻塞订单）
	for _, item := range order.CartItem {
		if _, err := oc.cartRPC.DeleteCart(ctx, &cartV1.DeleteCartRequest{Id: item.CartId, UserId: order.UserId}); err != nil {
			oc.log.Errorf("delete cart after order failed: cart=%d err=%v", item.CartId, err)
		}
	}
	fmt.Printf("order created: %s amount=%d\n", od.OrderSn, od.OrderAmount)
	return od, nil
}

func generateOrderSn() string {
	return fmt.Sprintf("%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

func generateEventID() string {
	return fmt.Sprintf("evt-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))
}

func buildOrderEventPayload(eventType, orderSn string, userId int64, items []*domain.OrderGoods) []byte {
	type skuItem struct {
		SkuID int64 `json:"sku_id"`
		Num   int32 `json:"num"`
	}
	type payload struct {
		Skus []skuItem `json:"skus"`
	}
	type event struct {
		EventID   string  `json:"event_id"`
		EventType string  `json:"event_type"`
		OrderSn   string  `json:"order_sn"`
		UserID    int64   `json:"user_id"`
		Payload   payload `json:"payload"`
	}
	e := event{
		EventID:   generateEventID(),
		EventType: eventType,
		OrderSn:   orderSn,
		UserID:    userId,
		Payload:   payload{},
	}
	for _, item := range items {
		e.Payload.Skus = append(e.Payload.Skus, skuItem{SkuID: item.SkuId, Num: item.Num})
	}
	b, _ := json.Marshal(e)
	return b
}

func (oc *OrderUsecase) CancelOrder(ctx context.Context, userId int64, orderSn string) error {
	items, err := oc.repo.ListItemsByOrderSn(ctx, orderSn)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return kerrors.New(400, "ORDER_NOT_FOUND", "订单不存在")
	}
	changed, err := oc.repo.UpdateStatusIf(ctx, orderSn, 1, 5)
	if err != nil {
		return err
	}
	if !changed {
		return nil // 已经是取消状态，幂等
	}
	outbox := &domain.OutboxEvent{
		EventID:   generateEventID(),
		EventType: "order.cancelled",
		OrderSn:   orderSn,
		Payload:   buildOrderEventPayload("order.cancelled", orderSn, userId, items),
	}
	return oc.repo.CreateOutbox(ctx, outbox)
}

func (oc *OrderUsecase) MarkPaid(ctx context.Context, orderSn string) error {
	items, err := oc.repo.ListItemsByOrderSn(ctx, orderSn)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return kerrors.New(400, "ORDER_NOT_FOUND", "订单不存在")
	}
	changed, err := oc.repo.UpdateStatusIf(ctx, orderSn, 1, 2)
	if err != nil {
		return err
	}
	if !changed {
		return nil // 已经是已支付状态，幂等
	}
	outbox := &domain.OutboxEvent{
		EventID:   generateEventID(),
		EventType: "order.paid",
		OrderSn:   orderSn,
		Payload:   buildOrderEventPayload("order.paid", orderSn, items[0].UserId, items),
	}
	return oc.repo.CreateOutbox(ctx, outbox)
}

func (oc *OrderUsecase) GetOrder(ctx context.Context, userId int64, orderSn string) (*domain.Order, error) {
	return oc.repo.GetDetail(ctx, userId, orderSn)
}

func (oc *OrderUsecase) ListOrders(ctx context.Context, userId int64, page, pageSize int) ([]*domain.Order, int64, error) {
	return oc.repo.ListByUser(ctx, userId, page, pageSize)
}

func (oc *OrderUsecase) relayLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := oc.RelayOutbox(context.Background()); err != nil {
			oc.log.Errorf("outbox relay error: %v", err)
		}
	}
}

func (oc *OrderUsecase) RelayOutbox(ctx context.Context) error {
	events, err := oc.repo.ListPendingOutbox(ctx, 50)
	if err != nil {
		return err
	}
	for _, evt := range events {
		if err := oc.publisher.Publish(ctx, "order.exchange", evt.EventType, evt.Payload); err != nil {
			oc.log.Errorf("publish outbox event failed: id=%d event=%s err=%v", evt.ID, evt.EventType, err)
			if err := oc.repo.IncrementOutboxRetry(ctx, evt.ID); err != nil {
				oc.log.Errorf("increment outbox retry failed: id=%d err=%v", evt.ID, err)
			}
			if evt.RetryCount+1 >= 5 {
				if err := oc.repo.MarkOutboxFailed(ctx, evt.ID); err != nil {
					oc.log.Errorf("mark outbox failed failed: id=%d err=%v", evt.ID, err)
				}
			}
			continue
		}
		if err := oc.repo.MarkOutboxPublished(ctx, evt.ID); err != nil {
			oc.log.Errorf("mark outbox published failed: id=%d err=%v", evt.ID, err)
		}
	}
	return nil
}

func (oc *OrderUsecase) timeoutLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := oc.CancelExpiredOrders(context.Background()); err != nil {
			oc.log.Errorf("cancel expired orders error: %v", err)
		}
	}
}

func (oc *OrderUsecase) CancelExpiredOrders(ctx context.Context) error {
	orders, err := oc.repo.ListPendingTimeout(ctx, 15)
	if err != nil {
		return err
	}
	for _, od := range orders {
		if err := oc.CancelOrder(ctx, od.User, od.OrderSn); err != nil {
			oc.log.Errorf("cancel expired order failed: %s err=%v", od.OrderSn, err)
		}
	}
	return nil
}
