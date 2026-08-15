package data

import (
	"context"
	serrors "errors"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"order/internal/biz"
	"order/internal/domain"
	"time"
)

type Order struct {
	ID            int64     `gorm:"primarykey"`
	User          int64     `gorm:"type:bigint;not null;default:0;index"`
	OrderSn       string    `gorm:"type:varchar(30);not null;default:'';index"` // 订单号，我们平台自己生成的订单号
	OrderAmount   int64     `gorm:"type:bigint;not null;default:0;comment:订单金额"`
	GoodsAmount   int64     `gorm:"type:bigint;not null;default:0;comment:商品总金额"`
	OrderStatus   int       `gorm:"type:int;not null;default:0;comment:0库存处理中,1待支付,2已支付,3已发货,4已签收,5已取消,6交易完成,7已退款"`
	ExpressAmount int64     `gorm:"type:bigint;not null;default:0;comment:运费"`
	DeliveryAt    time.Time `gorm:"column:delivery_at;type:timestamptz;comment:发货时间"`
	RefundTime    time.Time `gorm:"type:timestamptz;comment:退款时间"`
	Post          string    `gorm:"type:varchar(200);not null;default:'';comment:订单备注信息"`

	// 优惠信息、赠品、买反、优惠卷
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt
}

type OrderEventOutbox struct {
	ID          int64      `gorm:"primarykey;type:bigint"`
	EventID     string     `gorm:"column:event_id;type:varchar(64);not null;uniqueIndex"`
	EventType   string     `gorm:"column:event_type;type:varchar(32);not null"`
	OrderSn     string     `gorm:"column:order_sn;type:varchar(30);not null;default:''"`
	Payload     string     `gorm:"column:payload;type:text;not null"`
	Status      int        `gorm:"column:status;type:smallint;not null;default:0"`
	RetryCount  int        `gorm:"column:retry_count;type:int;not null;default:0"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	PublishedAt *time.Time `gorm:"column:published_at"`
}

func (OrderEventOutbox) TableName() string {
	return "order_event_outbox"
}

func (Order) TableName() string {
	return "orders"
}

type orderRepo struct {
	data *Data
	log  *log.Helper
}

// NewOrderRepo .
func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (p *Order) ToDomain() *domain.Order {
	return &domain.Order{
		ID:            p.ID,
		User:          p.User,
		OrderSn:       p.OrderSn,
		OrderAmount:   p.OrderAmount,
		GoodsAmount:   p.GoodsAmount,
		OrderStatus:   p.OrderStatus,
		ExpressAmount: p.ExpressAmount,
		DeliveryAt:    p.DeliveryAt,
		RefundTime:    p.RefundTime,
		Post:          p.Post,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func (o *orderRepo) Create(ctx context.Context, order *domain.Order, address *domain.OrderAddress, items []*domain.OrderGoods, outbox *domain.OutboxEvent) error {
	return o.data.ExecTx(ctx, func(ctx context.Context) error {
		od := Order{
			User:          order.User,
			OrderSn:       order.OrderSn,
			OrderAmount:   order.OrderAmount,
			GoodsAmount:   order.GoodsAmount,
			OrderStatus:   order.OrderStatus,
			ExpressAmount: order.ExpressAmount,
			Post:          order.Post,
		}
		if err := o.data.DB(ctx).Create(&od).Error; err != nil {
			return err
		}
		order.ID = od.ID
		order.CreatedAt = od.CreatedAt
		order.UpdatedAt = od.UpdatedAt

		if address != nil {
			addr := OrderAddress{
				User:            address.User,
				OrderSn:         address.OrderSn,
				RecipientName:   address.RecipientName,
				RecipientMobile: address.RecipientMobile,
				Province:        address.Province,
				City:            address.City,
				Districts:       address.Districts,
				Address:         address.Address,
				PostCode:        address.PostCode,
			}
			if err := o.data.DB(ctx).Create(&addr).Error; err != nil {
				return err
			}
		}

		for _, item := range items {
			og := OrderGoods{
				OrderSn:    item.OrderSn,
				UserId:     item.UserId,
				SkuId:      item.SkuId,
				SkuName:    item.SkuName,
				SkuPrice:   item.SkuPrice,
				Num:        item.Num,
				TotalPrice: item.TotalPrice,
			}
			if err := o.data.DB(ctx).Create(&og).Error; err != nil {
				return err
			}
		}

		if outbox != nil {
			oe := OrderEventOutbox{
				EventID:    outbox.EventID,
				EventType:  outbox.EventType,
				OrderSn:    outbox.OrderSn,
				Payload:    string(outbox.Payload),
				Status:     0,
				RetryCount: 0,
			}
			if err := o.data.DB(ctx).Create(&oe).Error; err != nil {
				return err
			}
			outbox.ID = oe.ID
		}
		return nil
	})
}

func (o *orderRepo) ListPendingOutbox(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var list []OrderEventOutbox
	if err := o.data.db.WithContext(ctx).
		Where("status = 0").
		Order("id ASC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.OutboxEvent, 0, len(list))
	for i := range list {
		res = append(res, &domain.OutboxEvent{
			ID:         list[i].ID,
			EventID:    list[i].EventID,
			EventType:  list[i].EventType,
			OrderSn:    list[i].OrderSn,
			Payload:    []byte(list[i].Payload),
			Status:     list[i].Status,
			RetryCount: list[i].RetryCount,
		})
	}
	return res, nil
}

func (o *orderRepo) MarkOutboxPublished(ctx context.Context, id int64) error {
	return o.data.db.WithContext(ctx).Model(&OrderEventOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       1,
			"published_at": time.Now(),
		}).Error
}

func (o *orderRepo) CreateOutbox(ctx context.Context, outbox *domain.OutboxEvent) error {
	oe := OrderEventOutbox{
		EventID:    outbox.EventID,
		EventType:  outbox.EventType,
		OrderSn:    outbox.OrderSn,
		Payload:    string(outbox.Payload),
		Status:     0,
		RetryCount: 0,
	}
	if err := o.data.db.WithContext(ctx).Create(&oe).Error; err != nil {
		return err
	}
	outbox.ID = oe.ID
	return nil
}

func (o *orderRepo) UpdateStatusIf(ctx context.Context, orderSn string, from, to int) (bool, error) {
	result := o.data.db.WithContext(ctx).Model(&Order{}).
		Where("order_sn = ? AND order_status = ?", orderSn, from).
		Updates(map[string]interface{}{
			"order_status": to,
			"updated_at":   time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		var od Order
		if err := o.data.db.WithContext(ctx).Where("order_sn = ?", orderSn).First(&od).Error; err != nil {
			return false, err
		}
		if od.OrderStatus == to {
			return false, nil // 幂等：已经是目标状态
		}
		return false, kerrors.New(400, "ORDER_STATUS_INVALID", "订单状态不允许此操作")
	}
	return true, nil
}

func (o *orderRepo) ListItemsByOrderSn(ctx context.Context, orderSn string) ([]*domain.OrderGoods, error) {
	var list []OrderGoods
	if err := o.data.db.WithContext(ctx).Where("order_sn = ?", orderSn).Find(&list).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.OrderGoods, 0, len(list))
	for i := range list {
		res = append(res, &domain.OrderGoods{
			ID:         list[i].ID,
			OrderSn:    list[i].OrderSn,
			UserId:     list[i].UserId,
			SkuId:      list[i].SkuId,
			SkuName:    list[i].SkuName,
			SkuPrice:   list[i].SkuPrice,
			Num:        list[i].Num,
			TotalPrice: list[i].TotalPrice,
		})
	}
	return res, nil
}

func (o *orderRepo) ListPendingTimeout(ctx context.Context, minutes int) ([]*domain.Order, error) {
	var list []Order
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)
	if err := o.data.db.WithContext(ctx).
		Where("order_status = ? AND created_at < ?", domain.OrderStatusPendingPayment, cutoff).
		Find(&list).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.Order, 0, len(list))
	for i := range list {
		res = append(res, list[i].ToDomain())
	}
	return res, nil
}

func (o *orderRepo) GetDetail(ctx context.Context, userId int64, orderSn string) (*domain.Order, error) {
	var od Order
	if err := o.data.db.WithContext(ctx).
		Where("order_sn = ? AND \"user\" = ?", orderSn, userId).
		First(&od).Error; err != nil {
		if serrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, kerrors.New(400, "ORDER_NOT_FOUND", "订单不存在")
		}
		return nil, err
	}
	dom := od.ToDomain()
	var addr OrderAddress
	if err := o.data.db.WithContext(ctx).Where("order_sn = ?", orderSn).First(&addr).Error; err == nil {
		dom.Address = addr.Address
		dom.SignerName = addr.RecipientName
		dom.SingerMobile = addr.RecipientMobile
	}
	items, err := o.ListItemsByOrderSn(ctx, orderSn)
	if err != nil {
		return nil, err
	}
	dom.Items = items
	return dom, nil
}

func (o *orderRepo) ListByUser(ctx context.Context, userId int64, page, pageSize int) ([]*domain.Order, int64, error) {
	var total int64
	if err := o.data.db.WithContext(ctx).Model(&Order{}).Where("\"user\" = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	var list []Order
	if err := o.data.db.WithContext(ctx).
		Where("\"user\" = ?", userId).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	res := make([]*domain.Order, 0, len(list))
	for i := range list {
		res = append(res, list[i].ToDomain())
	}
	return res, total, nil
}

func (o *orderRepo) AdminList(ctx context.Context, page, pageSize, status int) ([]*domain.Order, int64, error) {
	query := o.data.db.WithContext(ctx).Model(&Order{})
	if status > 0 {
		query = query.Where("order_status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	var list []Order
	if err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	res := make([]*domain.Order, 0, len(list))
	for i := range list {
		dom := list[i].ToDomain()
		var addr OrderAddress
		if err := o.data.db.WithContext(ctx).Where("order_sn = ?", dom.OrderSn).First(&addr).Error; err == nil {
			dom.Address = addr.Address
			dom.SignerName = addr.RecipientName
			dom.SingerMobile = addr.RecipientMobile
		}
		if items, err := o.ListItemsByOrderSn(ctx, dom.OrderSn); err == nil {
			dom.Items = items
		}
		res = append(res, dom)
	}
	return res, total, nil
}

func (o *orderRepo) GetByOrderSn(ctx context.Context, orderSn string) (*domain.Order, error) {
	var od Order
	if err := o.data.db.WithContext(ctx).
		Where("order_sn = ?", orderSn).
		First(&od).Error; err != nil {
		if serrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, kerrors.New(400, "ORDER_NOT_FOUND", "订单不存在")
		}
		return nil, err
	}
	dom := od.ToDomain()
	var addr OrderAddress
	if err := o.data.db.WithContext(ctx).Where("order_sn = ?", orderSn).First(&addr).Error; err == nil {
		dom.Address = addr.Address
		dom.SignerName = addr.RecipientName
		dom.SingerMobile = addr.RecipientMobile
	}
	items, err := o.ListItemsByOrderSn(ctx, orderSn)
	if err != nil {
		return nil, err
	}
	dom.Items = items
	return dom, nil
}

func (o *orderRepo) Ship(ctx context.Context, orderSn, post string) (bool, error) {
	result := o.data.db.WithContext(ctx).Model(&Order{}).
		Where("order_sn = ? AND order_status = 2", orderSn).
		Updates(map[string]interface{}{
			"order_status": 3,
			"post":         post,
			"delivery_at":  time.Now(),
			"updated_at":   time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (o *orderRepo) Refund(ctx context.Context, orderSn string) (bool, error) {
	result := o.data.db.WithContext(ctx).Model(&Order{}).
		Where("order_sn = ? AND order_status IN (2,3,4)", orderSn).
		Updates(map[string]interface{}{
			"order_status": 7,
			"refund_time":  time.Now(),
			"updated_at":   time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (o *orderRepo) DashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{}

	if err := o.data.db.WithContext(ctx).Model(&Order{}).Count(&stats.TotalOrders).Error; err != nil {
		return nil, err
	}
	if err := o.data.db.WithContext(ctx).Model(&Order{}).
		Where("order_status IN (2,3,4,6)").
		Select("COALESCE(SUM(order_amount),0)").
		Row().Scan(&stats.TotalSales); err != nil {
		return nil, err
	}
	if err := o.data.db.WithContext(ctx).Model(&Order{}).
		Where("created_at >= date_trunc('day', now())").
		Count(&stats.TodayOrders).Error; err != nil {
		return nil, err
	}
	if err := o.data.db.WithContext(ctx).Model(&Order{}).
		Where("created_at >= date_trunc('day', now()) AND order_status IN (2,3,4,6)").
		Select("COALESCE(SUM(order_amount),0)").
		Row().Scan(&stats.TodaySales); err != nil {
		return nil, err
	}

	var statusRows []struct {
		Status int32
		Count  int64
	}
	if err := o.data.db.WithContext(ctx).Model(&Order{}).
		Select("order_status AS status, COUNT(*) AS count").
		Group("order_status").
		Scan(&statusRows).Error; err != nil {
		return nil, err
	}
	for _, row := range statusRows {
		stats.StatusCounts = append(stats.StatusCounts, &domain.StatusCount{
			Status: row.Status,
			Count:  row.Count,
		})
	}

	var dailyRows []struct {
		Date       string
		OrderCount int64
		Amount     int64
	}
	if err := o.data.db.WithContext(ctx).Raw(`
		SELECT to_char(created_at, 'YYYY-MM-DD') AS date,
		       COUNT(*) AS order_count,
		       COALESCE(SUM(CASE WHEN order_status IN (2,3,4,6) THEN order_amount ELSE 0 END), 0) AS amount
		FROM orders
		WHERE created_at >= date_trunc('day', now()) - interval '29 days'
		GROUP BY date
		ORDER BY date
	`).Scan(&dailyRows).Error; err != nil {
		return nil, err
	}
	for _, row := range dailyRows {
		stats.Last30Days = append(stats.Last30Days, &domain.DailySales{
			Date:       row.Date,
			OrderCount: row.OrderCount,
			Amount:     row.Amount,
		})
	}

	var topRows []struct {
		SkuID   int64
		SkuName string
		Num     int64
		Amount  int64
	}
	if err := o.data.db.WithContext(ctx).Raw(`
		SELECT og.sku_id, og.sku_name, SUM(og.num) AS num, SUM(og.total_price) AS amount
		FROM order_goods og
		JOIN orders o ON o.order_sn = og.order_sn
		WHERE o.order_status IN (2,3,4,6)
		GROUP BY og.sku_id, og.sku_name
		ORDER BY num DESC
		LIMIT 5
	`).Scan(&topRows).Error; err != nil {
		return nil, err
	}
	for _, row := range topRows {
		stats.TopGoods = append(stats.TopGoods, &domain.TopGoods{
			SkuID:   row.SkuID,
			SkuName: row.SkuName,
			Num:     row.Num,
			Amount:  row.Amount,
		})
	}

	return stats, nil
}

func (o *orderRepo) IncrementOutboxRetry(ctx context.Context, id int64) error {
	return o.data.db.WithContext(ctx).Model(&OrderEventOutbox{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

func (o *orderRepo) MarkOutboxFailed(ctx context.Context, id int64) error {
	return o.data.db.WithContext(ctx).Model(&OrderEventOutbox{}).
		Where("id = ?", id).
		Update("status", 2).Error
}
