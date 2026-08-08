package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	orderV1 "payment/api/service/order/v1"
	v1 "payment/api/payment/v1"
	"payment/internal/biz"
	"payment/internal/conf"
	"payment/internal/domain"
	"payment/internal/pkg/mq"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewPaymentService)

type PaymentService struct {
	v1.UnimplementedPaymentServer

	uc        *biz.PaymentUsecase
	orderRPC  orderV1.OrderClient
	publisher *mq.Publisher
	log       *log.Helper
}

func NewPaymentService(uc *biz.PaymentUsecase, orderRPC orderV1.OrderClient, mqConf *conf.Mq, logger log.Logger) *PaymentService {
	s := &PaymentService{
		uc:       uc,
		orderRPC: orderRPC,
		log:      log.NewHelper(logger),
	}
	if mqConf != nil && mqConf.Addr != "" {
		s.publisher = mq.NewPublisher(mqConf.Addr)
	}
	return s
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *v1.CreatePaymentRequest) (*v1.PaymentInfoReply, error) {
	order, err := s.orderRPC.GetOrder(ctx, &orderV1.GetOrderRequest{
		UserId:  req.UserId,
		OrderSn: req.OrderSn,
	})
	if err != nil {
		return nil, err
	}
	if order.Total != req.Amount {
		return nil, kerrors.New(400, "PAYMENT_AMOUNT_MISMATCH", "支付金额与订单金额不一致")
	}
	pay, err := s.uc.Create(ctx, &domain.Payment{
		PaymentNo: generatePaymentNo(),
		OrderSn:   req.OrderSn,
		UserID:    req.UserId,
		Amount:    order.Total,
		Channel:   req.Channel,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PaymentInfoReply{
		Id:        pay.ID,
		PaymentNo: pay.PaymentNo,
		OrderSn:   pay.OrderSn,
		Amount:    pay.Amount,
		Channel:   pay.Channel,
		Status:    int32(pay.Status),
	}, nil
}

func (s *PaymentService) PaymentCallback(ctx context.Context, req *v1.PaymentCallbackRequest) (*v1.CheckReply, error) {
	pay, err := s.uc.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		return &v1.CheckReply{Success: false, Message: err.Error()}, nil
	}
	if !req.Success {
		return &v1.CheckReply{Success: false, Message: "支付失败"}, nil
	}
	order, err := s.orderRPC.GetOrder(ctx, &orderV1.GetOrderRequest{
		UserId:  pay.UserID,
		OrderSn: pay.OrderSn,
	})
	if err != nil {
		return &v1.CheckReply{Success: false, Message: "订单不存在"}, nil
	}
	if order.Status != "待支付" {
		return &v1.CheckReply{Success: false, Message: "订单状态不允许支付"}, nil
	}
	changed, err := s.uc.MarkPaid(ctx, req.PaymentNo, req.TradeNo)
	if err != nil {
		return &v1.CheckReply{Success: false, Message: err.Error()}, nil
	}
	if !changed {
		return &v1.CheckReply{Success: true, Message: "支付成功（重复回调已忽略）"}, nil
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"payment_no": pay.PaymentNo,
		"order_sn":   pay.OrderSn,
		"user_id":    pay.UserID,
		"amount":     pay.Amount,
		"trade_no":   req.TradeNo,
		"paid_at":    time.Now().Format(time.RFC3339),
	})
	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, "payment.exchange", "payment.success", payload); err != nil {
			s.log.Errorf("publish payment.success failed: %v", err)
			return &v1.CheckReply{Success: false, Message: "支付成功但事件发布失败"}, nil
		}
	}
	return &v1.CheckReply{Success: true, Message: "支付成功"}, nil
}

func generatePaymentNo() string {
	return fmt.Sprintf("PAY-%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}
