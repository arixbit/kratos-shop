package service

import (
	"cart/internal/biz"
	"cart/internal/domain"
	"context"

	v1 "cart/api/cart/v1"
)

type CartService struct {
	v1.UnimplementedCartServer
	cart *biz.CartUsecase
}

func NewCartService(cart *biz.CartUsecase) *CartService {
	return &CartService{cart: cart}
}

func (s *CartService) CreateCart(ctx context.Context, req *v1.CreateCartRequest) (*v1.CartInfoReply, error) {

	rv, err := s.cart.CreateCart(ctx, &domain.ShopCart{
		UserId:     req.UserId,
		GoodsId:    req.GoodsId,
		SkuId:      req.SkuId,
		GoodsPrice: req.GoodsPrice,
		GoodsNum:   req.GoodsNum,
		GoodsSn:    req.GoodsSn,
		GoodsName:  req.GoodsName,
		IsSelect:   req.IsSelect,
	})

	if err != nil {
		return nil, err
	}

	return &v1.CartInfoReply{
		Id:         rv.ID,
		UserId:     rv.UserId,
		GoodsId:    rv.GoodsId,
		GoodsSn:    rv.GoodsSn,
		GoodsName:  rv.GoodsName,
		SkuId:      rv.SkuId,
		GoodsPrice: rv.GoodsPrice,
		GoodsNum:   rv.GoodsNum,
		IsSelect:   rv.IsSelect,
	}, nil
}

func (s *CartService) ListCart(ctx context.Context, req *v1.ListCartRequest) (*v1.CartListReply, error) {
	res, err := s.cart.List(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	var rsp v1.CartListReply

	for _, cart := range res {
		rsp.Results = append(rsp.Results, toCartReply(cart))
	}

	return &rsp, nil
}

func (s *CartService) UpdateCart(ctx context.Context, req *v1.UpdateCartRequest) (*v1.CheckResponse, error) {
	if err := s.cart.UpdateCart(ctx, &domain.ShopCart{
		ID:       req.Id,
		UserId:   req.UserId,
		GoodsNum: req.GoodsNum,
	}); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *CartService) DeleteCart(ctx context.Context, req *v1.DeleteCartRequest) (*v1.CheckResponse, error) {
	if err := s.cart.DeleteCart(ctx, req.Id, req.UserId); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.GetCartReply, error) {
	cart, err := s.cart.GetCart(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.GetCartReply{Info: toCartReply(cart)}, nil
}

func toCartReply(cart *domain.ShopCart) *v1.CartInfoReply {
	return &v1.CartInfoReply{
		Id:         cart.ID,
		UserId:     cart.UserId,
		GoodsId:    cart.GoodsId,
		GoodsSn:    cart.GoodsSn,
		GoodsName:  cart.GoodsName,
		SkuId:      cart.SkuId,
		GoodsPrice: cart.GoodsPrice,
		GoodsNum:   cart.GoodsNum,
		IsSelect:   cart.IsSelect,
	}
}
