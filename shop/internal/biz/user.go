package biz

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwt2 "github.com/golang-jwt/jwt/v5"
	v1 "shop/api/shop/v1"
	"shop/internal/conf"
	"shop/internal/pkg/captcha"
	"shop/internal/pkg/middleware/auth"
	"time"
)

var (
	ErrPasswordInvalid     = errors.New("password invalid")
	ErrUsernameInvalid     = errors.New("username invalid")
	ErrCaptchaInvalid      = errors.New("verification code error")
	ErrMobileInvalid       = errors.New("mobile invalid")
	ErrUserNotFound        = errors.New("user not found")
	ErrLoginFailed         = errors.New("login failed")
	ErrGenerateTokenFailed = errors.New("generate token failed")
	ErrAuthFailed          = errors.New("authentication failed")
)

type User struct {
	ID        int64
	Mobile    string
	Password  string
	NickName  string
	Birthday  int64
	Gender    string
	Role      int
	CreatedAt time.Time
}

type UserRepo interface {
	CreateUser(c context.Context, u *User) (*User, error)
	UserByMobile(ctx context.Context, mobile string) (*User, error)
	UserById(ctx context.Context, Id int64) (*User, error)
	CheckPassword(ctx context.Context, password, encryptedPassword string) (bool, error)
	//ListUser(ctx context.Context, pageNum, pageSize int) ([]*User, int, error)
	//UpdateUser(context.Context, *User) (bool, error)
}

type UserUsecase struct {
	uRepo      UserRepo
	log        *log.Helper
	signingKey string
}

func NewUserUsecase(repo UserRepo, logger log.Logger, conf *conf.Auth) *UserUsecase {
	helper := log.NewHelper(log.With(logger, "module", "usecase/shop"))
	return &UserUsecase{uRepo: repo, log: helper, signingKey: conf.JwtKey}
}

// GetCaptcha 验证码
func (uc *UserUsecase) GetCaptcha(ctx context.Context) (*v1.CaptchaReply, error) {
	captchaInfo, err := captcha.GetCaptcha(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.CaptchaReply{
		CaptchaId: captchaInfo.CaptchaId,
		PicPath:   captchaInfo.PicPath,
	}, nil
}

func (uc *UserUsecase) UserDetailByID(ctx context.Context) (*v1.UserDetailResponse, error) {
	// 在上下文 context 中取出 claims 对象
	var uId int64
	if claims, ok := jwt.FromContext(ctx); ok {
		c := claims.(jwt2.MapClaims)
		i, ok := c["ID"].(float64)
		if !ok {
			return nil, ErrAuthFailed
		}
		uId = int64(i)
	}

	user, err := uc.uRepo.UserById(ctx, uId)
	if err != nil {
		return nil, err
	}
	return &v1.UserDetailResponse{
		Id:       user.ID,
		NickName: user.NickName,
		Mobile:   user.Mobile,
	}, nil
}

func (uc *UserUsecase) PassWordLogin(ctx context.Context, req *v1.LoginReq) (*v1.RegisterReply, error) {
	// 表单验证
	if len(req.Mobile) <= 0 {
		return nil, ErrMobileInvalid
	}
	if len(req.Password) <= 0 {
		return nil, ErrUsernameInvalid
	}
	// 验证验证码是否正确
	if !captcha.Store.Verify(req.CaptchaId, req.Captcha, true) {
		return nil, ErrCaptchaInvalid
	}

	if user, err := uc.uRepo.UserByMobile(ctx, req.Mobile); err != nil {
		return nil, ErrUserNotFound
	} else {
		// 用户存在检查密码
		if passRsp, pasErr := uc.uRepo.CheckPassword(ctx, req.Password, user.Password); pasErr != nil {
			return nil, ErrPasswordInvalid
		} else {
			if passRsp {
				claims := auth.CustomClaims{
					ID:          user.ID,
					NickName:    user.NickName,
					AuthorityId: user.Role,
					RegisteredClaims: jwt2.RegisteredClaims{
						NotBefore: jwt2.NewNumericDate(time.Now()),                            // 签名的生效时间
						ExpiresAt: jwt2.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),   // 30天过期
						Issuer:    "Gyl",
					},
				}

				token, err := auth.CreateToken(claims, uc.signingKey)
				if err != nil {
					return nil, ErrGenerateTokenFailed
				}
				return &v1.RegisterReply{
					Id:        user.ID,
					Mobile:    user.Mobile,
					Username:  user.NickName,
					Token:     token,
					ExpiredAt: time.Now().Unix() + 60*60*24*30,
				}, nil
			} else {
				return nil, ErrLoginFailed
			}
		}
	}
}

func (uc *UserUsecase) CreateUser(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	newUser, err := NewUser(req.Mobile, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	createUser, err := uc.uRepo.CreateUser(ctx, &newUser)
	claims := auth.CustomClaims{
		ID:          createUser.ID,
		NickName:    createUser.NickName,
		AuthorityId: createUser.Role,
		RegisteredClaims: jwt2.RegisteredClaims{
			NotBefore: jwt2.NewNumericDate(time.Now()),                          // 签名的生效时间
			ExpiresAt: jwt2.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30天过期
			Issuer:    "Gyl",
		},
	}
	token, err := auth.CreateToken(claims, uc.signingKey)
	if err != nil {
		return nil, err
	}

	return &v1.RegisterReply{
		Id:        createUser.ID,
		Mobile:    createUser.Mobile,
		Username:  createUser.NickName,
		Token:     token,
		ExpiredAt: time.Now().Unix() + 60*60*24*30,
	}, nil
}

func (uc *UserUsecase) RefreshToken(ctx context.Context, token string) (*v1.RegisterReply, error) {
	claims, err := auth.ParseToken(token, uc.signingKey)
	if err != nil {
		return nil, ErrAuthFailed
	}
	newToken, err := auth.CreateToken(auth.CustomClaims{
		ID:          claims.ID,
		NickName:    claims.NickName,
		AuthorityId: claims.AuthorityId,
		RegisteredClaims: jwt2.RegisteredClaims{
			NotBefore: jwt2.NewNumericDate(time.Now()),
			ExpiresAt: jwt2.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			Issuer:    "Gyl",
		},
	}, uc.signingKey)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterReply{
		Id:        claims.ID,
		Username:  claims.NickName,
		Token:     newToken,
		ExpiredAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
	}, nil
}

func NewUser(mobile, username, password string) (User, error) {
	// check mobile
	if len(mobile) <= 0 {
		return User{}, ErrMobileInvalid
	}
	// check username
	if len(username) <= 0 {
		return User{}, ErrUsernameInvalid
	}
	// check password
	if len(password) <= 0 {
		return User{}, ErrPasswordInvalid
	}
	return User{
		Mobile:   mobile,
		NickName: username,
		Password: password,
	}, nil
}
