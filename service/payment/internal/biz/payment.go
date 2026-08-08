package biz

import (
	"context"

	"payment/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewPaymentUsecase)

type PaymentRepo interface {
	Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error)
	GetByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error)
	MarkPaid(ctx context.Context, paymentNo, tradeNo string) (bool, error)
}

type PaymentUsecase struct {
	repo PaymentRepo
	log  *log.Helper
}

func NewPaymentUsecase(repo PaymentRepo, logger log.Logger) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *PaymentUsecase) Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	return uc.repo.Create(ctx, p)
}

func (uc *PaymentUsecase) GetByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	return uc.repo.GetByPaymentNo(ctx, paymentNo)
}

func (uc *PaymentUsecase) MarkPaid(ctx context.Context, paymentNo, tradeNo string) (bool, error) {
	return uc.repo.MarkPaid(ctx, paymentNo, tradeNo)
}
