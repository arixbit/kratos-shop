package service

import (
	"context"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/emptypb"
	"shop/internal/biz"

	v1 "shop/api/shop/v1"
)

func (s *ShopService) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	//  add trace
	tr := otel.Tracer("service")
	ctx, span := tr.Start(ctx, "get user info by mobile")
	span.SpanContext()
	defer span.End()

	return s.uc.CreateUser(ctx, req)
}

func (s *ShopService) Login(ctx context.Context, req *v1.LoginReq) (*v1.RegisterReply, error) {
	return s.uc.PassWordLogin(ctx, req)
}

func (s *ShopService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RegisterReply, error) {
	return s.uc.RefreshToken(ctx, req.Token)
}

func (s *ShopService) Captcha(ctx context.Context, r *emptypb.Empty) (*v1.CaptchaReply, error) {
	return s.uc.GetCaptcha(ctx)
}

func (s *ShopService) Detail(ctx context.Context, r *emptypb.Empty) (*v1.UserDetailResponse, error) {
	return s.uc.UserDetailByID(ctx)
}

func (s *ShopService) CreateAddress(ctx context.Context, r *v1.CreateAddressReq) (*v1.AddressInfo, error) {
	return s.ua.CreateAddress(ctx, r)
}

func (s *ShopService) AddressListByUid(ctx context.Context, empty *emptypb.Empty) (*v1.ListAddressReply, error) {
	return s.ua.AddressListByUid(ctx)
}

func (s *ShopService) UpdateAddress(ctx context.Context, r *v1.UpdateAddressReq) (*v1.CheckResponse, error) {
	req := toBizAddress(r)
	address, err := s.ua.UpdateAddress(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: address}, nil
}

func toBizAddress(r *v1.UpdateAddressReq) *biz.Address {
	return &biz.Address{
		ID:        r.Id,
		IsDefault: r.IsDefault,
		Mobile:    r.Mobile,
		Name:      r.Name,
		Province:  r.Province,
		City:      r.City,
		Districts: r.Districts,
		Address:   r.Address,
		PostCode:  r.PostCode,
	}
}

func (s *ShopService) DefaultAddress(ctx context.Context, r *v1.AddressReq) (*v1.CheckResponse, error) {

	address, err := s.ua.DefaultAddress(ctx, &biz.Address{
		ID: r.Id,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: address}, nil
}

func (s *ShopService) DeleteAddress(ctx context.Context, r *v1.AddressReq) (*v1.CheckResponse, error) {
	address, err := s.ua.DeleteAddress(ctx, &biz.Address{
		ID: r.Id,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: address}, nil
}
