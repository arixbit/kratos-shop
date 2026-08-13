package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	cartV1 "shop/api/service/cart/v1"
	v1 "shop/api/shop/v1"
)

type CartUsecase struct {
	cc  cartV1.CartClient
	log *log.Helper
}

func NewCartUsecase(cc cartV1.CartClient, logger log.Logger) *CartUsecase {
	return &CartUsecase{cc: cc, log: log.NewHelper(log.With(logger, "module", "usecase/cart"))}
}

func (uc *CartUsecase) List(ctx context.Context) (*v1.CartListReply, error) {
	uid, err := getUid(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := uc.cc.ListCart(ctx, &cartV1.ListCartRequest{UserId: uid})
	if err != nil {
		return nil, err
	}
	reply := &v1.CartListReply{}
	for _, item := range resp.Results {
		reply.List = append(reply.List, toCartItemReply(item))
	}
	return reply, nil
}

func (uc *CartUsecase) Add(ctx context.Context, req *v1.CartAddRequest) (*v1.CartItemReply, error) {
	uid, err := getUid(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := uc.cc.CreateCart(ctx, &cartV1.CreateCartRequest{
		UserId:     uid,
		GoodsId:    req.GoodsId,
		GoodsSn:    req.GoodsSn,
		GoodsName:  req.GoodsName,
		SkuId:      req.SkuId,
		GoodsPrice: req.GoodsPrice,
		GoodsNum:   req.GoodsNum,
		IsSelect:   req.IsSelect,
	})
	if err != nil {
		return nil, err
	}
	return toCartItemReply(resp), nil
}

func (uc *CartUsecase) Update(ctx context.Context, req *v1.CartUpdateRequest) error {
	uid, err := getUid(ctx)
	if err != nil {
		return err
	}
	_, err = uc.cc.UpdateCart(ctx, &cartV1.UpdateCartRequest{
		Id:       req.Id,
		UserId:   uid,
		GoodsNum: req.GoodsNum,
	})
	return err
}

func (uc *CartUsecase) Delete(ctx context.Context, req *v1.CartDeleteRequest) error {
	uid, err := getUid(ctx)
	if err != nil {
		return err
	}
	_, err = uc.cc.DeleteCart(ctx, &cartV1.DeleteCartRequest{
		Id:     req.Id,
		UserId: uid,
	})
	return err
}

func toCartItemReply(item *cartV1.CartInfoReply) *v1.CartItemReply {
	if item == nil {
		return nil
	}
	return &v1.CartItemReply{
		Id:         item.Id,
		UserId:     item.UserId,
		GoodsId:    item.GoodsId,
		GoodsSn:    item.GoodsSn,
		GoodsName:  item.GoodsName,
		SkuId:      item.SkuId,
		GoodsPrice: item.GoodsPrice,
		GoodsNum:   item.GoodsNum,
		IsSelect:   item.IsSelect,
	}
}
