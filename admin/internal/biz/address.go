package biz

import (
	v1 "admin/api/admin/v1"
	"admin/internal/conf"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

type Address struct {
	ID        int64
	UserID    int64
	IsDefault int32
	Mobile    string
	Name      string
	Province  string
	City      string
	Districts string
	Address   string
	PostCode  string
}

type AddressRepo interface {
	CreateAddress(ctx context.Context, a *Address) (*Address, error)
	AddressListByUid(ctx context.Context, uid int64) ([]*Address, error)
	UpdateAddress(ctx context.Context, a *Address) error
	DefaultAddress(ctx context.Context, a *Address) error
	DeleteAddress(ctx context.Context, a *Address) error
}

type AddressUsecase struct {
	uRepo      UserRepo
	aRepo      AddressRepo
	log        *log.Helper
	signingKey string
}

func NewAddressUsecase(repo UserRepo, arepo AddressRepo, logger log.Logger, conf *conf.Auth) *AddressUsecase {
	helper := log.NewHelper(log.With(logger, "module", "usecase/admin"))
	return &AddressUsecase{
		uRepo:      repo,
		aRepo:      arepo,
		log:        helper,
		signingKey: conf.JwtKey}
}

func (ua *AddressUsecase) CreateAddress(ctx context.Context, r *v1.CreateAddressReq) (*v1.AddressInfo, error) {
	// 在上下文 context 中取出 claims 对象
	uId, err := getUid(ctx)
	if err != nil {
		return nil, err
	}

	req := Address{
		UserID:    uId,
		IsDefault: 0,
		Mobile:    r.Mobile,
		Name:      r.Name,
		Province:  r.Province,
		City:      r.City,
		Districts: r.Districts,
		Address:   r.Address,
		PostCode:  r.PostCode,
	}
	res, err := ua.aRepo.CreateAddress(ctx, &req)
	if err != nil {
		return nil, err
	}

	result := &v1.AddressInfo{
		Id:        res.ID,
		Name:      res.Name,
		Mobile:    res.Mobile,
		Province:  res.Province,
		City:      res.City,
		Districts: res.Districts,
		Address:   res.Address,
		PostCode:  res.PostCode,
		IsDefault: int32(res.IsDefault),
	}
	return result, nil
}

func (ua *AddressUsecase) AddressListByUid(ctx context.Context) (*v1.ListAddressReply, error) {
	// 在上下文 context 中取出 claims 对象
	uId, err := getUid(ctx)
	if err != nil {
		return nil, err
	}
	addressList, err := ua.aRepo.AddressListByUid(ctx, uId)
	var res v1.ListAddressReply
	for _, v := range addressList {
		addressInfoTmp := &v1.AddressInfo{
			Id:        v.ID,
			Name:      v.Name,
			Mobile:    v.Mobile,
			Province:  v.Province,
			City:      v.City,
			Districts: v.Districts,
			Address:   v.Address,
			PostCode:  v.PostCode,
			IsDefault: v.IsDefault,
		}
		res.Results = append(res.Results, addressInfoTmp)
	}
	return &res, err
}

func (ua *AddressUsecase) UserAddressList(ctx context.Context, uid int64) (*v1.UserAddressListReply, error) {
	addressList, err := ua.aRepo.AddressListByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	reply := &v1.UserAddressListReply{}
	for _, v := range addressList {
		reply.List = append(reply.List, &v1.AddressInfo{
			Id:        v.ID,
			Name:      v.Name,
			Mobile:    v.Mobile,
			Province:  v.Province,
			City:      v.City,
			Districts: v.Districts,
			Address:   v.Address,
			PostCode:  v.PostCode,
			IsDefault: v.IsDefault,
		})
	}
	return reply, nil
}

func (ua *AddressUsecase) UserAddressDelete(ctx context.Context, id, uid int64) error {
	return ua.aRepo.DeleteAddress(ctx, &Address{ID: id, UserID: uid})
}

func (ua *AddressUsecase) UpdateAddress(ctx context.Context, a *Address) (bool, error) {
	uId, err := getUid(ctx)
	if err != nil {
		return false, err
	}
	a.UserID = uId
	if err := ua.aRepo.UpdateAddress(ctx, a); err != nil {
		return false, err
	}
	return true, nil
}

func getUid(ctx context.Context) (int64, error) {
	// 在上下文 context 中取出 claims 对象
	var uId int64
	if claims, ok := jwt.FromContext(ctx); ok {
		c := claims.(jwt2.MapClaims)
		v, ok := c["ID"]

		if !ok {
			return 0, ErrAuthFailed
		}
		uId = int64(v.(float64))
	}
	return uId, nil
}

func (ua *AddressUsecase) DefaultAddress(ctx context.Context, a *Address) (bool, error) {
	uId, err := getUid(ctx)
	if err != nil {
		return false, err
	}
	a.UserID = uId
	if err := ua.aRepo.DefaultAddress(ctx, a); err != nil {
		return false, err
	}
	return true, nil
}

func (ua *AddressUsecase) DeleteAddress(ctx context.Context, a *Address) (bool, error) {
	uId, err := getUid(ctx)
	if err != nil {
		return false, err
	}
	a.UserID = uId
	if err := ua.aRepo.DeleteAddress(ctx, a); err != nil {
		return false, err
	}
	return true, nil
}
