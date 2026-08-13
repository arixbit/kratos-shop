package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"

	paymentV1 "shop/api/service/payment/v1"
	v1 "shop/api/shop/v1"
)

type PaymentUsecase struct {
	pc  paymentV1.PaymentClient
	log *log.Helper
}

func NewPaymentUsecase(pc paymentV1.PaymentClient, logger log.Logger) *PaymentUsecase {
	return &PaymentUsecase{pc: pc, log: log.NewHelper(log.With(logger, "module", "usecase/payment"))}
}

func (uc *PaymentUsecase) Create(ctx context.Context, req *v1.PaymentCreateRequest) (*v1.PaymentInfoReply, error) {
	uid, err := getUid(ctx)
	if err != nil {
		return nil, err
	}
	channel := req.Channel
	if channel == "" {
		channel = "mock"
	}
	resp, err := uc.pc.CreatePayment(ctx, &paymentV1.CreatePaymentRequest{
		UserId:  uid,
		OrderSn: req.OrderSn,
		Amount:  req.Amount,
		Channel: channel,
	})
	if err != nil {
		return nil, err
	}
	return toPaymentInfoReply(resp), nil
}

func (uc *PaymentUsecase) MockPay(ctx context.Context, req *v1.PaymentMockPayRequest) error {
	tradeNo := fmt.Sprintf("MOCK-%d", time.Now().UnixNano())
	resp, err := uc.pc.PaymentCallback(ctx, &paymentV1.PaymentCallbackRequest{
		PaymentNo: req.PaymentNo,
		TradeNo:   tradeNo,
		Success:   true,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return errors.New(400, "PAYMENT_FAILED", resp.Message)
	}
	return nil
}

func toPaymentInfoReply(item *paymentV1.PaymentInfoReply) *v1.PaymentInfoReply {
	if item == nil {
		return nil
	}
	return &v1.PaymentInfoReply{
		Id:        item.Id,
		PaymentNo: item.PaymentNo,
		OrderSn:   item.OrderSn,
		Amount:    item.Amount,
		Channel:   item.Channel,
		Status:    item.Status,
	}
}
