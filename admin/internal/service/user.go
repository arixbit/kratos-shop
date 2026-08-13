package service

import (
	"admin/internal/biz"
	"context"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/emptypb"

	v1 "admin/api/admin/v1"
)

func (s *AdminService) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	//  add trace
	tr := otel.Tracer("service")
	ctx, span := tr.Start(ctx, "get user info by mobile")
	span.SpanContext()
	defer span.End()

	return s.uc.CreateUser(ctx, req)
}

func (s *AdminService) Login(ctx context.Context, req *v1.LoginReq) (*v1.RegisterReply, error) {
	return s.uc.PassWordLogin(ctx, req)
}

func (s *AdminService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RegisterReply, error) {
	return s.uc.RefreshToken(ctx, req.Token)
}

func (s *AdminService) Captcha(ctx context.Context, r *emptypb.Empty) (*v1.CaptchaReply, error) {
	return s.uc.GetCaptcha(ctx)
}

func (s *AdminService) Detail(ctx context.Context, r *emptypb.Empty) (*v1.UserDetailResponse, error) {
	reply, err := s.uc.UserDetailByID(ctx)
	if err != nil {
		return nil, err
	}
	perms, err := s.pu.CodesForRole(ctx, reply.Role)
	if err != nil {
		return nil, err
	}
	reply.Permissions = perms
	return reply, nil
}

func (s *AdminService) UserList(ctx context.Context, req *v1.UserListRequest) (*v1.UserListReply, error) {
	return s.uc.UserList(ctx, req)
}

func (s *AdminService) UserAddressList(ctx context.Context, req *v1.UserAddressListRequest) (*v1.UserAddressListReply, error) {
	return s.ua.UserAddressList(ctx, req.Uid)
}

func (s *AdminService) UserAddressDelete(ctx context.Context, req *v1.UserAddressDeleteRequest) (*v1.CheckResponse, error) {
	if err := s.ua.UserAddressDelete(ctx, req.Id, req.Uid); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *AdminService) CreateAddress(ctx context.Context, r *v1.CreateAddressReq) (*v1.AddressInfo, error) {
	return s.ua.CreateAddress(ctx, r)
}

func (s *AdminService) AddressListByUid(ctx context.Context, empty *emptypb.Empty) (*v1.ListAddressReply, error) {
	return s.ua.AddressListByUid(ctx)
}

func (s *AdminService) UpdateAddress(ctx context.Context, r *v1.UpdateAddressReq) (*v1.CheckResponse, error) {
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

func (s *AdminService) DefaultAddress(ctx context.Context, r *v1.AddressReq) (*v1.CheckResponse, error) {

	address, err := s.ua.DefaultAddress(ctx, &biz.Address{
		ID: r.Id,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: address}, nil
}

func (s *AdminService) DeleteAddress(ctx context.Context, r *v1.AddressReq) (*v1.CheckResponse, error) {
	address, err := s.ua.DeleteAddress(ctx, &biz.Address{
		ID: r.Id,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: address}, nil
}
