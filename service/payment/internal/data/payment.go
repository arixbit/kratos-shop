package data

import (
	"context"
	"errors"
	"time"

	"payment/internal/biz"
	"payment/internal/domain"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type Payment struct {
	ID        int64      `gorm:"primarykey;type:bigint"`
	PaymentNo string     `gorm:"column:payment_no;type:varchar(64);not null;uniqueIndex"`
	OrderSn   string     `gorm:"column:order_sn;type:varchar(64);not null;index"`
	UserID    int64      `gorm:"column:user_id;type:bigint;not null;default:0"`
	Amount    int64      `gorm:"column:amount;type:bigint;not null;default:0"`
	Channel   string     `gorm:"column:channel;type:varchar(32);not null;default:'mock'"`
	Status    int        `gorm:"column:status;type:int;not null;default:1"`
	TradeNo   string     `gorm:"column:trade_no;type:varchar(100);not null;default:''"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	PaidAt    *time.Time `gorm:"column:paid_at"`
}

func (Payment) TableName() string {
	return "payments"
}

func (p *Payment) ToDomain() *domain.Payment {
	return &domain.Payment{
		ID:        p.ID,
		PaymentNo: p.PaymentNo,
		OrderSn:   p.OrderSn,
		UserID:    p.UserID,
		Amount:    p.Amount,
		Channel:   p.Channel,
		Status:    p.Status,
		TradeNo:   p.TradeNo,
		PaidAt:    p.PaidAt,
	}
}

type paymentRepo struct {
	data *Data
	log  *log.Helper
}

func NewPaymentRepo(data *Data, logger log.Logger) biz.PaymentRepo {
	return &paymentRepo{data: data, log: log.NewHelper(logger)}
}

func (r *paymentRepo) Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	pay := Payment{
		PaymentNo: p.PaymentNo,
		OrderSn:   p.OrderSn,
		UserID:    p.UserID,
		Amount:    p.Amount,
		Channel:   p.Channel,
		Status:    1,
	}
	if err := r.data.db.WithContext(ctx).Create(&pay).Error; err != nil {
		return nil, err
	}
	return pay.ToDomain(), nil
}

func (r *paymentRepo) GetByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	var pay Payment
	if err := r.data.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&pay).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, kerrors.New(400, "PAYMENT_NOT_FOUND", "支付单不存在")
		}
		return nil, err
	}
	return pay.ToDomain(), nil
}

func (r *paymentRepo) MarkPaid(ctx context.Context, paymentNo, tradeNo string) (bool, error) {
	now := time.Now()
	result := r.data.db.WithContext(ctx).Model(&Payment{}).
		Where("payment_no = ? AND status = 1", paymentNo).
		Updates(map[string]interface{}{
			"status":   2,
			"trade_no": tradeNo,
			"paid_at":  now,
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		var pay Payment
		if err := r.data.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&pay).Error; err != nil {
			return false, err
		}
		if pay.Status == 2 {
			return false, nil // 幂等：已支付
		}
		return false, kerrors.New(400, "PAYMENT_NOT_FOUND", "支付单不存在")
	}
	return true, nil
}
