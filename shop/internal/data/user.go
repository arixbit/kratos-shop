package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	userService "shop/api/service/user/v1"
	"shop/internal/biz"
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
