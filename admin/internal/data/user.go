package data

import (
	userService "admin/api/service/user/v1"
	"admin/internal/biz"
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo .
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "repo/user")),
	}
}

func (u *userRepo) CreateUser(c context.Context, user *biz.User) (*biz.User, error) {
	createUser, err := u.data.uc.CreateUser(c, &userService.CreateUserInfo{
		NickName: user.NickName,
		Password: user.Password,
		Mobile:   user.Mobile,
	})
	if err != nil {
		return nil, err
	}
	return &biz.User{
		ID:       createUser.Id,
		Mobile:   createUser.Mobile,
		NickName: createUser.NickName,
	}, nil
}

func (u *userRepo) UserByMobile(c context.Context, mobile string) (*biz.User, error) {
	byMobile, err := u.data.uc.GetUserByMobile(c, &userService.MobileRequest{Mobile: mobile})
	if err != nil {
		return nil, err
	}
	return &biz.User{
		Mobile:   byMobile.Mobile,
		ID:       byMobile.Id,
		Password: byMobile.Password,
		NickName: byMobile.NickName,
		Role:     int(byMobile.Role),
	}, nil
}

func (u *userRepo) CheckPassword(c context.Context, password, encryptedPassword string) (bool, error) {
	if byMobile, err := u.data.uc.CheckPassword(c, &userService.PasswordCheckInfo{Password: password, EncryptedPassword: encryptedPassword}); err != nil {
		return false, err
	} else {
		return byMobile.Success, nil
	}
}

func (u *userRepo) UserById(c context.Context, id int64) (*biz.User, error) {
	user, err := u.data.uc.GetUserById(c, &userService.IdRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &biz.User{
		ID:       user.Id,
		Mobile:   user.Mobile,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int(user.Role),
		Birthday: int64(user.Birthday),
	}, nil
}

func (u *userRepo) ListUser(ctx context.Context, pageNum, pageSize int) ([]*biz.User, int, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	resp, err := u.data.uc.GetUserList(ctx, &userService.PageInfo{
		Pn:    uint32(pageNum),
		PSize: uint32(pageSize),
	})
	if err != nil {
		return nil, 0, err
	}
	list := make([]*biz.User, 0, len(resp.Data))
	for _, item := range resp.Data {
		list = append(list, &biz.User{
			ID:       item.Id,
			Mobile:   item.Mobile,
			NickName: item.NickName,
			Birthday: int64(item.Birthday),
			Gender:   item.Gender,
			Role:     int(item.Role),
		})
	}
	return list, int(resp.Total), nil
}
