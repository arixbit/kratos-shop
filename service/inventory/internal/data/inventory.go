package data

import (
	"context"
	"errors"
	"strconv"
	"time"

	"inventory/internal/biz"
	"inventory/internal/domain"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const stockKeyPrefix = "stock:sku:"

func stockKey(skuID int64) string {
	return stockKeyPrefix + strconv.FormatInt(skuID, 10)
}

type Inventory struct {
	ID        int64     `gorm:"primarykey;type:bigint"`
	SkuID     int64     `gorm:"column:sku_id;type:bigint;not null;uniqueIndex"`
	Inventory int64     `gorm:"column:inventory;type:bigint;not null;default:0"`
	Locked    int64     `gorm:"column:locked;type:bigint;not null;default:0"`
	Version   int64     `gorm:"column:version;type:bigint;not null;default:0"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Inventory) TableName() string {
	return "inventories"
}

type InventoryLock struct {
	ID        int64     `gorm:"primarykey;type:bigint"`
	OrderSn   string    `gorm:"column:order_sn;type:varchar(64);not null;index"`
	SkuID     int64     `gorm:"column:sku_id;type:bigint;not null;index"`
	Num       int32     `gorm:"column:num;type:int;not null"`
	Status    int       `gorm:"column:status;type:int;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (InventoryLock) TableName() string {
	return "inventory_locks"
}

type InventoryFlow struct {
	ID        int64     `gorm:"primarykey;type:bigint"`
	OrderSn   string    `gorm:"column:order_sn;type:varchar(64);not null;index"`
	SkuID     int64     `gorm:"column:sku_id;type:bigint;not null;index"`
	Change    int64     `gorm:"column:change;type:bigint;not null"`
	Type      string    `gorm:"column:type;type:varchar(32);not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (InventoryFlow) TableName() string {
	return "inventory_flows"
}

type ConsumedEvent struct {
	EventID    string    `gorm:"column:event_id;type:varchar(64);primaryKey"`
	OrderSn    string    `gorm:"column:order_sn;type:varchar(64);not null;default:''"`
	ConsumedAt time.Time `gorm:"column:consumed_at"`
}

func (ConsumedEvent) TableName() string {
	return "consumed_event"
}

func (p *Inventory) ToDomain() *domain.Inventory {
	return &domain.Inventory{
		ID:        p.ID,
		SkuID:     p.SkuID,
		Inventory: p.Inventory,
		Locked:    p.Locked,
		Version:   p.Version,
	}
}

type inventoryRepo struct {
	data *Data
	log  *log.Helper
}

func NewInventoryRepo(data *Data, logger log.Logger) biz.InventoryRepo {
	return &inventoryRepo{data: data, log: log.NewHelper(logger)}
}

func (r *inventoryRepo) Query(ctx context.Context, skuIds []int64) ([]*domain.Inventory, error) {
	if len(skuIds) == 0 {
		return nil, nil
	}
	res := make([]*domain.Inventory, 0, len(skuIds))
	var missing []int64
	for _, id := range skuIds {
		inv, ok, err := r.queryCache(ctx, id)
		if err != nil {
			return nil, err
		}
		if ok {
			res = append(res, inv)
			continue
		}
		missing = append(missing, id)
	}
	if len(missing) > 0 {
		var list []Inventory
		if err := r.data.db.WithContext(ctx).Where("sku_id IN (?)", missing).Find(&list).Error; err != nil {
			return nil, err
		}
		for i := range list {
			res = append(res, list[i].ToDomain())
			_ = r.cacheSet(ctx, &list[i])
		}
	}
	return res, nil
}

func (r *inventoryRepo) TryLock(ctx context.Context, orderSn string, items []*domain.SkuItem) error {
	for _, item := range items {
		ok, err := r.tryLockRedis(ctx, item)
		if err != nil {
			return err
		}
		if !ok {
			return kerrors.New(400, "STOCK_NOT_ENOUGH", "库存不足")
		}
	}
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var inv Inventory
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("sku_id = ?", item.SkuID).First(&inv).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return kerrors.New(400, "STOCK_NOT_FOUND", "SKU 库存不存在")
				}
				return err
			}
			if inv.Inventory-inv.Locked < int64(item.Num) {
				return kerrors.New(400, "STOCK_NOT_ENOUGH", "库存不足")
			}

			var lock InventoryLock
			err := tx.Where("order_sn = ? AND sku_id = ?", orderSn, item.SkuID).First(&lock).Error
			if err == nil {
				continue // 幂等：同一订单同一 SKU 已预占
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			if err := tx.Model(&Inventory{}).Where("id = ?", inv.ID).
				Updates(map[string]interface{}{
					"locked":    gorm.Expr("locked + ?", item.Num),
					"version":   gorm.Expr("version + 1"),
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
			lock = InventoryLock{
				OrderSn: orderSn,
				SkuID:   item.SkuID,
				Num:     item.Num,
				Status:  1,
			}
			if err := tx.Create(&lock).Error; err != nil {
				return err
			}
			if err := tx.Create(&InventoryFlow{
				OrderSn: orderSn,
				SkuID:   item.SkuID,
				Change:  int64(item.Num),
				Type:    "lock",
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		for _, item := range items {
			r.rollbackRedisLock(ctx, item)
		}
		return err
	}
	for _, item := range items {
		_ = r.cacheInvalidate(ctx, item.SkuID)
	}
	return nil
}

func (r *inventoryRepo) ConfirmDeduct(ctx context.Context, orderSn string, items []*domain.SkuItem) error {
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var lock InventoryLock
			if err := tx.Where("order_sn = ? AND sku_id = ? AND status = 1", orderSn, item.SkuID).
				First(&lock).Error; err != nil {
				return kerrors.New(400, "LOCK_NOT_FOUND", "预占记录不存在")
			}
			if err := tx.Model(&Inventory{}).Where("sku_id = ?", item.SkuID).
				Updates(map[string]interface{}{
					"inventory":  gorm.Expr("inventory - ?", item.Num),
					"locked":     gorm.Expr("locked - ?", item.Num),
					"version":    gorm.Expr("version + 1"),
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
			if err := tx.Model(&InventoryLock{}).Where("id = ?", lock.ID).
				Update("status", 2).Error; err != nil {
				return err
			}
			if err := tx.Create(&InventoryFlow{
				OrderSn: orderSn,
				SkuID:   item.SkuID,
				Change:  -int64(item.Num),
				Type:    "deduct",
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, item := range items {
		_ = r.cacheInvalidate(ctx, item.SkuID)
	}
	return nil
}

func (r *inventoryRepo) Release(ctx context.Context, orderSn string, items []*domain.SkuItem) error {
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var lock InventoryLock
			if err := tx.Where("order_sn = ? AND sku_id = ? AND status = 1", orderSn, item.SkuID).
				First(&lock).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			if err := tx.Model(&Inventory{}).Where("sku_id = ?", item.SkuID).
				Updates(map[string]interface{}{
					"locked":     gorm.Expr("locked - ?", item.Num),
					"version":    gorm.Expr("version + 1"),
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
			if err := tx.Model(&InventoryLock{}).Where("id = ?", lock.ID).
				Update("status", 3).Error; err != nil {
				return err
			}
			if err := tx.Create(&InventoryFlow{
				OrderSn: orderSn,
				SkuID:   item.SkuID,
				Change:  int64(item.Num),
				Type:    "release",
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, item := range items {
		_ = r.cacheInvalidate(ctx, item.SkuID)
	}
	return nil
}

func (r *inventoryRepo) IsConsumed(ctx context.Context, eventID string) (bool, error) {
	var evt ConsumedEvent
	err := r.data.db.WithContext(ctx).Where("event_id = ?", eventID).First(&evt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *inventoryRepo) MarkConsumed(ctx context.Context, eventID, orderSn string) error {
	return r.data.db.WithContext(ctx).Create(&ConsumedEvent{
		EventID:    eventID,
		OrderSn:    orderSn,
		ConsumedAt: time.Now(),
	}).Error
}

func (r *inventoryRepo) queryCache(ctx context.Context, skuID int64) (*domain.Inventory, bool, error) {
	key := stockKey(skuID)
	values, err := r.data.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if len(values) == 0 {
		return nil, false, nil
	}
	total, err1 := strconv.ParseInt(values["total"], 10, 64)
	locked, err2 := strconv.ParseInt(values["locked"], 10, 64)
	if err1 != nil || err2 != nil {
		return nil, false, nil
	}
	return &domain.Inventory{SkuID: skuID, Inventory: total, Locked: locked}, true, nil
}

func (r *inventoryRepo) cacheSet(ctx context.Context, inv *Inventory) error {
	return r.data.rdb.HSet(ctx, stockKey(inv.SkuID), "total", inv.Inventory, "locked", inv.Locked).Err()
}

func (r *inventoryRepo) cacheInvalidate(ctx context.Context, skuID int64) error {
	return r.data.rdb.Del(ctx, stockKey(skuID)).Err()
}

func (r *inventoryRepo) tryLockRedis(ctx context.Context, item *domain.SkuItem) (bool, error) {
	script := redis.NewScript(`
local total = tonumber(redis.call('HGET', KEYS[1], 'total') or '-1')
local locked = tonumber(redis.call('HGET', KEYS[1], 'locked') or '0')
local num = tonumber(ARGV[1])
if total >= 0 and total - locked >= num then
  redis.call('HSET', KEYS[1], 'locked', locked + num)
  return 1
end
return 0
`)
	for attempt := 0; attempt < 2; attempt++ {
		res, err := script.Run(ctx, r.data.rdb, []string{stockKey(item.SkuID)}, item.Num).Int()
		if err != nil {
			return false, err
		}
		if res == 1 {
			return true, nil
		}
		var inv Inventory
		if err := r.data.db.WithContext(ctx).Where("sku_id = ?", item.SkuID).First(&inv).Error; err != nil {
			return false, err
		}
		if err := r.cacheSet(ctx, &inv); err != nil {
			return false, err
		}
	}
	return false, nil
}

func (r *inventoryRepo) rollbackRedisLock(ctx context.Context, item *domain.SkuItem) {
	_ = r.data.rdb.HIncrBy(ctx, stockKey(item.SkuID), "locked", -int64(item.Num)).Err()
	_ = r.cacheInvalidate(ctx, item.SkuID)
}
