package service

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"

	orderV1 "payment/api/service/order/v1"
	v1 "payment/api/payment/v1"
	"payment/internal/biz"
	"payment/internal/domain"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type fakeOrderClient struct {
	orderV1.OrderClient
	order *orderV1.OrderInfoResponse
	err   error
}

func (f *fakeOrderClient) GetOrder(_ context.Context, _ *orderV1.GetOrderRequest, _ ...grpc.CallOption) (*orderV1.OrderInfoResponse, error) {
	return f.order, f.err
}

type fakePaymentRepo struct {
	biz.PaymentRepo

	payment    *domain.Payment
	getErr     error
	createErr  error
	changed    bool
	markErr    error
	created    *domain.Payment
}

func (f *fakePaymentRepo) Create(_ context.Context, p *domain.Payment) (*domain.Payment, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = p
	return p, nil
}

func (f *fakePaymentRepo) GetByPaymentNo(_ context.Context, _ string) (*domain.Payment, error) {
	return f.payment, f.getErr
}

func (f *fakePaymentRepo) MarkPaid(_ context.Context, _, _ string) (bool, error) {
	return f.changed, f.markErr
}

func newTestPaymentService(repo *fakePaymentRepo, order *fakeOrderClient) *PaymentService {
	return &PaymentService{
		uc:       biz.NewPaymentUsecase(repo, log.DefaultLogger),
		orderRPC: order,
		log:      log.NewHelper(log.DefaultLogger),
	}
}

func paymentReason(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	return kerrors.FromError(err).Reason
}

func TestCreatePaymentAmountMismatch(t *testing.T) {
	repo := &fakePaymentRepo{}
	order := &fakeOrderClient{order: &orderV1.OrderInfoResponse{Total: 100, Status: "待支付"}}
	s := newTestPaymentService(repo, order)

	_, err := s.CreatePayment(context.Background(), &v1.CreatePaymentRequest{
		UserId:  1,
		OrderSn: "SN-001",
		Amount:  99,
	})
	if reason := paymentReason(t, err); reason != "PAYMENT_AMOUNT_MISMATCH" {
		t.Fatalf("reason = %q, want PAYMENT_AMOUNT_MISMATCH", reason)
	}
	if repo.created != nil {
		t.Fatal("repo.Create should not be called on mismatch")
	}
}

func TestCreatePaymentSuccess(t *testing.T) {
	repo := &fakePaymentRepo{}
	order := &fakeOrderClient{order: &orderV1.OrderInfoResponse{Total: 100, Status: "待支付"}}
	s := newTestPaymentService(repo, order)

	reply, err := s.CreatePayment(context.Background(), &v1.CreatePaymentRequest{
		UserId:  1,
		OrderSn: "SN-001",
		Amount:  100,
		Channel: "mock",
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if reply.Amount != 100 || reply.OrderSn != "SN-001" || reply.Channel != "mock" {
		t.Fatalf("unexpected reply: %+v", reply)
	}
	if repo.created == nil || repo.created.PaymentNo == "" {
		t.Fatal("repo.Create should persist payment with generated no")
	}
}

func TestPaymentCallbackPaymentNotFound(t *testing.T) {
	repo := &fakePaymentRepo{getErr: errors.New("not found")}
	s := newTestPaymentService(repo, &fakeOrderClient{})

	reply, err := s.PaymentCallback(context.Background(), &v1.PaymentCallbackRequest{
		PaymentNo: "PAY-1",
		Success:   true,
	})
	if err != nil {
		t.Fatalf("payment callback: %v", err)
	}
	if reply.Success {
		t.Fatal("callback should fail when payment not found")
	}
}

func TestPaymentCallbackFailure(t *testing.T) {
	repo := &fakePaymentRepo{payment: &domain.Payment{PaymentNo: "PAY-1", OrderSn: "SN-001", UserID: 1}}
	s := newTestPaymentService(repo, &fakeOrderClient{})

	reply, err := s.PaymentCallback(context.Background(), &v1.PaymentCallbackRequest{
		PaymentNo: "PAY-1",
		Success:   false,
	})
	if err != nil {
		t.Fatalf("payment callback: %v", err)
	}
	if reply.Success || reply.Message != "支付失败" {
		t.Fatalf("unexpected reply: %+v", reply)
	}
}

func TestPaymentCallbackOrderNotPending(t *testing.T) {
	repo := &fakePaymentRepo{payment: &domain.Payment{PaymentNo: "PAY-1", OrderSn: "SN-001", UserID: 1}}
	order := &fakeOrderClient{order: &orderV1.OrderInfoResponse{Status: "已支付"}}
	s := newTestPaymentService(repo, order)

	reply, err := s.PaymentCallback(context.Background(), &v1.PaymentCallbackRequest{
		PaymentNo: "PAY-1",
		Success:   true,
		TradeNo:   "TRADE-1",
	})
	if err != nil {
		t.Fatalf("payment callback: %v", err)
	}
	if reply.Success || reply.Message != "订单状态不允许支付" {
		t.Fatalf("unexpected reply: %+v", reply)
	}
}

func TestPaymentCallbackSuccess(t *testing.T) {
	repo := &fakePaymentRepo{
		payment: &domain.Payment{PaymentNo: "PAY-1", OrderSn: "SN-001", UserID: 1, Amount: 100},
		changed: true,
	}
	order := &fakeOrderClient{order: &orderV1.OrderInfoResponse{Status: "待支付"}}
	s := newTestPaymentService(repo, order)

	reply, err := s.PaymentCallback(context.Background(), &v1.PaymentCallbackRequest{
		PaymentNo: "PAY-1",
		Success:   true,
		TradeNo:   "TRADE-1",
	})
	if err != nil {
		t.Fatalf("payment callback: %v", err)
	}
	if !reply.Success || reply.Message != "支付成功" {
		t.Fatalf("unexpected reply: %+v", reply)
	}
}

func TestPaymentCallbackDuplicateIgnored(t *testing.T) {
	repo := &fakePaymentRepo{
		payment: &domain.Payment{PaymentNo: "PAY-1", OrderSn: "SN-001", UserID: 1},
		changed: false,
	}
	order := &fakeOrderClient{order: &orderV1.OrderInfoResponse{Status: "待支付"}}
	s := newTestPaymentService(repo, order)

	reply, err := s.PaymentCallback(context.Background(), &v1.PaymentCallbackRequest{
		PaymentNo: "PAY-1",
		Success:   true,
		TradeNo:   "TRADE-1",
	})
	if err != nil {
		t.Fatalf("payment callback: %v", err)
	}
	if !reply.Success || reply.Message != "支付成功（重复回调已忽略）" {
		t.Fatalf("unexpected reply: %+v", reply)
	}
}
