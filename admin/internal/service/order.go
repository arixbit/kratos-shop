package service

import (
	"context"

	v1 "admin/api/admin/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminService) OrderList(ctx context.Context, req *v1.OrderListRequest) (*v1.OrderListReply, error) {
	return s.ou.List(ctx, req)
}

func (s *AdminService) OrderDetail(ctx context.Context, req *v1.OrderDetailRequest) (*v1.OrderInfo, error) {
	return s.ou.Detail(ctx, req)
}

func (s *AdminService) ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.CheckResponse, error) {
	if err := s.ou.Ship(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *AdminService) RefundOrder(ctx context.Context, req *v1.RefundOrderRequest) (*v1.CheckResponse, error) {
	if err := s.ou.Refund(ctx, req); err != nil {
		return nil, err
	}
	return &v1.CheckResponse{Success: true}, nil
}

func (s *AdminService) DashboardStats(ctx context.Context, req *emptypb.Empty) (*v1.DashboardStatsReply, error) {
	stats, err := s.ou.DashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	userCount, err := s.uc.CountUser(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = userCount
	return stats, nil
}
