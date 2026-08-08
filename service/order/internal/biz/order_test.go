package biz

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"google.golang.org/grpc"

	cartV1 "order/api/cart/v1"
	goodsV1 "order/api/goods/v1"
	userV1 "order/api/user/v1"
	"order/internal/domain"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type fakeCartClient struct {
	cartV1.CartClient
	list       *cartV1.CartListReply
	listErr    error
	deleteCalls int
}

func (f *fakeCartClient) ListCart(_ context.Context, _ *cartV1.ListCartRequest, _ ...grpc.CallOption) (*cartV1.CartListReply, error) {
	return f.list, f.listErr
}

func (f *fakeCartClient) DeleteCart(_ context.Context, _ *cartV1.DeleteCartRequest, _ ...grpc.CallOption) (*cartV1.CheckResponse, error) {
	f.deleteCalls++
	return &cartV1.CheckResponse{Success: true}, nil
}

type fakeGoodsClient struct {
	goodsV1.GoodsClient
	list    *goodsV1.SkuListResponse
	listErr error
}

func (f *fakeGoodsClient) SkuList(_ context.Context, _ *goodsV1.SkuListRequest, _ ...grpc.CallOption) (*goodsV1.SkuListResponse, error) {
	return f.list, f.listErr
}

type fakeUserClient struct {
	userV1.UserClient
	address *userV1.AddressInfo
	addrErr error
}

func (f *fakeUserClient) GetAddress(_ context.Context, _ *userV1.AddressReq, _ ...grpc.CallOption) (*userV1.AddressInfo, error) {
	return f.address, f.addrErr
}

type fakeOrderRepo struct {
	OrderRepo

	createErr     error
	createdOrder  *domain.Order
	createdItems  []*domain.OrderGoods
	createdOutbox *domain.OutboxEvent

	items         []*domain.OrderGoods
	listItemsErr  error
	updateChanged bool
	updateErr     error
	outboxCreated []*domain.OutboxEvent
}

func (f *fakeOrderRepo) Create(_ context.Context, order *domain.Order, _ *domain.OrderAddress, items []*domain.OrderGoods, outbox *domain.OutboxEvent) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdOrder = order
	f.createdItems = items
	f.createdOutbox = outbox
	return nil
}

func (f *fakeOrderRepo) ListItemsByOrderSn(_ context.Context, _ string) ([]*domain.OrderGoods, error) {
	return f.items, f.listItemsErr
}

func (f *fakeOrderRepo) UpdateStatusIf(_ context.Context, _ string, _, _ int) (bool, error) {
	return f.updateChanged, f.updateErr
}

func (f *fakeOrderRepo) CreateOutbox(_ context.Context, outbox *domain.OutboxEvent) error {
	f.outboxCreated = append(f.outboxCreated, outbox)
	return nil
}

func newTestOrderUsecase(repo *fakeOrderRepo, cart *fakeCartClient, goods *fakeGoodsClient, user *fakeUserClient) *OrderUsecase {
	return &OrderUsecase{
		repo:     repo,
		userRPC:  user,
		cartRPC:  cart,
		goodsRPC: goods,
		log:      log.NewHelper(log.DefaultLogger),
	}
}

func reasonOf(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	return kerrors.FromError(err).Reason
}

func TestGenerateOrderSn(t *testing.T) {
	sn := generateOrderSn()
	if len(sn) != 20 {
		t.Fatalf("order sn length = %d, want 20", len(sn))
	}
	for _, c := range sn {
		if c < '0' || c > '9' {
			t.Fatalf("order sn should contain only digits: %q", sn)
		}
	}
}

func TestBuildOrderEventPayload(t *testing.T) {
	payload := buildOrderEventPayload("order.created", "SN-001", 7, []*domain.OrderGoods{
		{SkuId: 11, Num: 2},
		{SkuId: 12, Num: 1},
	})
	var evt struct {
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
		OrderSn   string `json:"order_sn"`
		UserID    int64  `json:"user_id"`
		Payload   struct {
			Skus []struct {
				SkuID int64 `json:"sku_id"`
				Num   int32 `json:"num"`
			} `json:"skus"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(payload, &evt); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if evt.EventID == "" || evt.EventType != "order.created" || evt.OrderSn != "SN-001" || evt.UserID != 7 {
		t.Fatalf("unexpected event: %+v", evt)
	}
	if len(evt.Payload.Skus) != 2 || evt.Payload.Skus[0].SkuID != 11 || evt.Payload.Skus[0].Num != 2 {
		t.Fatalf("unexpected skus: %+v", evt.Payload.Skus)
	}
}

func TestCreateOrderValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("nil order", func(t *testing.T) {
		uc := newTestOrderUsecase(&fakeOrderRepo{}, &fakeCartClient{}, &fakeGoodsClient{}, &fakeUserClient{})
		_, err := uc.CreateOrder(ctx, nil)
		if reason := reasonOf(t, err); reason != "ORDER_ITEM_EMPTY" {
			t.Fatalf("reason = %q, want ORDER_ITEM_EMPTY", reason)
		}
	})

	t.Run("empty cart items", func(t *testing.T) {
		uc := newTestOrderUsecase(&fakeOrderRepo{}, &fakeCartClient{}, &fakeGoodsClient{}, &fakeUserClient{})
		_, err := uc.CreateOrder(ctx, &domain.CreateOrder{UserId: 1, CartItem: nil})
		if reason := reasonOf(t, err); reason != "ORDER_ITEM_EMPTY" {
			t.Fatalf("reason = %q, want ORDER_ITEM_EMPTY", reason)
		}
	})

	cart := &fakeCartClient{list: &cartV1.CartListReply{
		Results: []*cartV1.CartInfoReply{{Id: 1, SkuId: 11}},
	}}
	goods := &fakeGoodsClient{list: &goodsV1.SkuListResponse{
		List: []*goodsV1.SkuInfo{{Id: 11, Price: 100, OnSale: true}},
	}}
	user := &fakeUserClient{address: &userV1.AddressInfo{Address: "addr", Name: "n", Mobile: "13800138000"}}

	t.Run("invalid sku num", func(t *testing.T) {
		uc := newTestOrderUsecase(&fakeOrderRepo{}, cart, goods, user)
		_, err := uc.CreateOrder(ctx, &domain.CreateOrder{
			UserId:    1,
			AddressId: 1,
			CartItem:  domain.CartItemList{{CartId: 1, SkuId: 11, SkuNum: 0}},
		})
		if reason := reasonOf(t, err); reason != "SKU_NUM_INVALID" {
			t.Fatalf("reason = %q, want SKU_NUM_INVALID", reason)
		}
	})

	t.Run("cart item mismatch", func(t *testing.T) {
		uc := newTestOrderUsecase(&fakeOrderRepo{}, cart, goods, user)
		_, err := uc.CreateOrder(ctx, &domain.CreateOrder{
			UserId:    1,
			AddressId: 1,
			CartItem:  domain.CartItemList{{CartId: 1, SkuId: 22, SkuNum: 1}},
		})
		if reason := reasonOf(t, err); reason != "CART_ITEM_INVALID" {
			t.Fatalf("reason = %q, want CART_ITEM_INVALID", reason)
		}
	})

	t.Run("sku not found", func(t *testing.T) {
		uc := newTestOrderUsecase(&fakeOrderRepo{}, cart, &fakeGoodsClient{list: &goodsV1.SkuListResponse{}}, user)
		_, err := uc.CreateOrder(ctx, &domain.CreateOrder{
			UserId:    1,
			AddressId: 1,
			CartItem:  domain.CartItemList{{CartId: 1, SkuId: 11, SkuNum: 1}},
		})
		if reason := reasonOf(t, err); reason != "SKU_NOT_FOUND" {
			t.Fatalf("reason = %q, want SKU_NOT_FOUND", reason)
		}
	})

	t.Run("sku off sale", func(t *testing.T) {
		offSale := &fakeGoodsClient{list: &goodsV1.SkuListResponse{
			List: []*goodsV1.SkuInfo{{Id: 11, Price: 100, OnSale: false}},
		}}
		uc := newTestOrderUsecase(&fakeOrderRepo{}, cart, offSale, user)
		_, err := uc.CreateOrder(ctx, &domain.CreateOrder{
			UserId:    1,
			AddressId: 1,
			CartItem:  domain.CartItemList{{CartId: 1, SkuId: 11, SkuNum: 1}},
		})
		if reason := reasonOf(t, err); reason != "SKU_NOT_FOUND" {
			t.Fatalf("reason = %q, want SKU_NOT_FOUND", reason)
		}
	})
}

func TestCreateOrderSuccess(t *testing.T) {
	ctx := context.Background()
	repo := &fakeOrderRepo{}
	cart := &fakeCartClient{list: &cartV1.CartListReply{
		Results: []*cartV1.CartInfoReply{{Id: 1, SkuId: 11}},
	}}
	goods := &fakeGoodsClient{list: &goodsV1.SkuListResponse{
		List: []*goodsV1.SkuInfo{{Id: 11, SkuName: "手机", Price: 100, PromotionPrice: 90, OnSale: true}},
	}}
	user := &fakeUserClient{address: &userV1.AddressInfo{
		Address: "上海市", Name: "张三", Mobile: "13800138000",
		Province: "上海", City: "上海", Districts: "浦东", PostCode: "200000",
	}}
	uc := newTestOrderUsecase(repo, cart, goods, user)

	order, err := uc.CreateOrder(ctx, &domain.CreateOrder{
		UserId:    7,
		AddressId: 3,
		CartItem:  domain.CartItemList{{CartId: 1, SkuId: 11, SkuNum: 2}},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.OrderAmount != 180 || order.GoodsAmount != 180 || order.OrderStatus != 1 {
		t.Fatalf("unexpected order: %+v", order)
	}
	if order.Address != "上海市" || order.SignerName != "张三" || order.SingerMobile != "13800138000" {
		t.Fatalf("unexpected address fields: %+v", order)
	}
	if repo.createdOrder == nil || repo.createdOrder.OrderSn != order.OrderSn {
		t.Fatal("repo.Create should be called with created order")
	}
	if len(repo.createdItems) != 1 || repo.createdItems[0].SkuPrice != 90 || repo.createdItems[0].TotalPrice != 180 {
		t.Fatalf("unexpected order items: %+v", repo.createdItems)
	}
	if repo.createdOutbox == nil || repo.createdOutbox.EventType != "order.created" {
		t.Fatalf("unexpected outbox: %+v", repo.createdOutbox)
	}
	if cart.deleteCalls != 1 {
		t.Fatalf("delete cart calls = %d, want 1", cart.deleteCalls)
	}
}

func TestCancelOrderIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := &fakeOrderRepo{
		items:         []*domain.OrderGoods{{SkuId: 11, Num: 1}},
		updateChanged: false,
	}
	uc := newTestOrderUsecase(repo, &fakeCartClient{}, &fakeGoodsClient{}, &fakeUserClient{})
	if err := uc.CancelOrder(ctx, 1, "SN-001"); err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if len(repo.outboxCreated) != 0 {
		t.Fatalf("idempotent cancel should not create outbox, got %d", len(repo.outboxCreated))
	}
}

func TestMarkPaidSuccess(t *testing.T) {
	ctx := context.Background()
	repo := &fakeOrderRepo{
		items:         []*domain.OrderGoods{{UserId: 7, SkuId: 11, Num: 1}},
		updateChanged: true,
	}
	uc := newTestOrderUsecase(repo, &fakeCartClient{}, &fakeGoodsClient{}, &fakeUserClient{})
	if err := uc.MarkPaid(ctx, "SN-001"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if len(repo.outboxCreated) != 1 || repo.outboxCreated[0].EventType != "order.paid" {
		t.Fatalf("unexpected outbox: %+v", repo.outboxCreated)
	}
	if !strings.Contains(string(repo.outboxCreated[0].Payload), "SN-001") {
		t.Fatalf("outbox payload should contain order sn: %s", repo.outboxCreated[0].Payload)
	}
}

func TestListItemsError(t *testing.T) {
	ctx := context.Background()
	repo := &fakeOrderRepo{listItemsErr: errors.New("db down")}
	uc := newTestOrderUsecase(repo, &fakeCartClient{}, &fakeGoodsClient{}, &fakeUserClient{})
	if err := uc.CancelOrder(ctx, 1, "SN-001"); err == nil {
		t.Fatal("expected error from repo")
	}
}
