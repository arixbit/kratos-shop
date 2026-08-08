package service

import (
	"context"
	"encoding/json"
	"time"

	v1 "inventory/api/inventory/v1"
	"inventory/internal/biz"
	"inventory/internal/conf"
	"inventory/internal/domain"
	"inventory/internal/pkg/mq"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewInventoryService)

type InventoryService struct {
	v1.UnimplementedInventoryServer

	uc     *biz.InventoryUsecase
	log    *log.Helper
	mqAddr string
}

func NewInventoryService(uc *biz.InventoryUsecase, mqConf *conf.Mq, logger log.Logger) *InventoryService {
	s := &InventoryService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
	if mqConf != nil && mqConf.Addr != "" {
		s.mqAddr = mqConf.Addr
		go s.startConsumer()
	}
	return s
}

func (s *InventoryService) QueryStock(ctx context.Context, req *v1.StockQueryRequest) (*v1.StockQueryReply, error) {
	invs, err := s.uc.Query(ctx, req.SkuIds)
	if err != nil {
		return nil, err
	}
	reply := &v1.StockQueryReply{}
	for _, inv := range invs {
		reply.Items = append(reply.Items, &v1.StockItem{
			SkuId:     inv.SkuID,
			Inventory: inv.Inventory,
			Locked:    inv.Locked,
			Available: inv.Inventory - inv.Locked,
		})
	}
	return reply, nil
}

func (s *InventoryService) TryLock(ctx context.Context, req *v1.TryLockRequest) (*v1.TryLockReply, error) {
	if err := s.uc.TryLock(ctx, req.OrderSn, toDomainItems(req.Items)); err != nil {
		return &v1.TryLockReply{Success: false, Reason: err.Error()}, nil
	}
	return &v1.TryLockReply{Success: true}, nil
}

func (s *InventoryService) ConfirmDeduct(ctx context.Context, req *v1.ConfirmDeductRequest) (*v1.CheckReply, error) {
	if err := s.uc.ConfirmDeduct(ctx, req.OrderSn, toDomainItems(req.Items)); err != nil {
		return &v1.CheckReply{Success: false, Reason: err.Error()}, nil
	}
	return &v1.CheckReply{Success: true}, nil
}

func (s *InventoryService) Release(ctx context.Context, req *v1.ReleaseRequest) (*v1.CheckReply, error) {
	if err := s.uc.Release(ctx, req.OrderSn, toDomainItems(req.Items)); err != nil {
		return &v1.CheckReply{Success: false, Reason: err.Error()}, nil
	}
	return &v1.CheckReply{Success: true}, nil
}

func toDomainItems(items []*v1.SkuItem) []*domain.SkuItem {
	res := make([]*domain.SkuItem, 0, len(items))
	for _, item := range items {
		res = append(res, &domain.SkuItem{SkuID: item.SkuId, Num: item.Num})
	}
	return res
}

type orderCreatedEvent struct {
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

func (s *InventoryService) startConsumer() {
	for {
		var consumer *mq.Consumer
		consumer, err := mq.NewConsumer(
			s.mqAddr,
			"order.exchange",
			"q.order.created",
			[]string{"order.created", "order.cancelled", "order.paid"},
			func(ctx context.Context, body []byte) error {
				return s.handleOrderCreated(ctx, body, func(routingKey string, payload []byte) error {
					return consumer.Publish("inventory.exchange", routingKey, payload)
				})
			},
		)
		if err != nil {
			s.log.Errorf("mq consumer init failed: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if err := consumer.Run(context.Background()); err != nil {
			s.log.Errorf("mq consumer stopped: %v", err)
		}
		consumer.Close()
		time.Sleep(5 * time.Second)
	}
}

func (s *InventoryService) handleOrderCreated(ctx context.Context, body []byte, publish func(string, []byte) error) error {
	var evt orderCreatedEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}
	consumed, err := s.uc.IsConsumed(ctx, evt.EventID)
	if err != nil {
		return err
	}
	if consumed {
		return nil
	}
	items := make([]*domain.SkuItem, 0, len(evt.Payload.Skus))
	for _, sku := range evt.Payload.Skus {
		items = append(items, &domain.SkuItem{SkuID: sku.SkuID, Num: sku.Num})
	}

	routingKey := "inventory.locked"
	reason := ""
	var opErr error
	if evt.EventType == "order.cancelled" {
		routingKey = "inventory.released"
		opErr = s.uc.Release(ctx, evt.OrderSn, items)
		if opErr != nil {
			routingKey = "inventory.release.failed"
			reason = opErr.Error()
			s.log.Errorf("release failed: order=%s err=%v", evt.OrderSn, opErr)
		}
	} else {
		if evt.EventType == "order.paid" {
			routingKey = "inventory.deducted"
			opErr = s.uc.ConfirmDeduct(ctx, evt.OrderSn, items)
			if opErr != nil {
				routingKey = "inventory.deduct.failed"
				reason = opErr.Error()
				s.log.Errorf("confirm deduct failed: order=%s err=%v", evt.OrderSn, opErr)
			}
		} else {
			opErr = s.uc.TryLock(ctx, evt.OrderSn, items)
			if opErr != nil {
				routingKey = "inventory.lock.failed"
				reason = opErr.Error()
				s.log.Errorf("try lock failed: order=%s err=%v", evt.OrderSn, opErr)
			}
		}
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"order_sn": evt.OrderSn,
		"success":  reason == "",
		"reason":   reason,
	})
	if err := publish(routingKey, payload); err != nil {
		return err
	}
	// 业务失败（库存不足等）也视为终态并幂等记录，避免无限重试；DB/网络错误不记录，交给 MQ 重试
	if opErr == nil || kerrors.FromError(opErr) != nil {
		if err := s.uc.MarkConsumed(ctx, evt.EventID, evt.OrderSn); err != nil {
			return err
		}
	}
	return nil
}
