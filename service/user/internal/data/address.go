package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"time"
	"user/internal/biz"
	"user/internal/domain"
)

type Address struct {
	ID        int64     `gorm:"primarykey"`
	UserID    int64     `gorm:"index:idx_user_id;column:user_id;default:1;type:bigint;comment:关联用户ID"`
	IsDefault int       `gorm:"column:is_default;default:0;type:smallint;comment:是否是默认"`
	Mobile    string    `gorm:"index:idx_mobile;type:varchar(11);comment:手机号码;not null"`
	Name      string    `gorm:"type:varchar(25);comment:收货用户名称"`
	Province  string    `gorm:"type:varchar(25);comment:省"`
	City      string    `gorm:"type:varchar(25);comment:市"`
	Districts string    `gorm:"type:varchar(25);comment:区县"`
	Address   string    `gorm:"type:varchar(255);comment:收货详细地址"`
	PostCode  string    `gorm:"type:varchar(25);comment:邮编"`
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt
}

func (Address) TableName() string {
	return "user_addresses"
}

type adderessRepo struct {
	data *Data
	log  *log.Helper
}

// NewAddressRepo .
func NewAddressRepo(data *Data, logger log.Logger) biz.AddressRepo {
	return &adderessRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (p *Address) ToDomain() *domain.Address {
	return &domain.Address{
		ID:        p.ID,
		UserID:    p.UserID,
		IsDefault: p.IsDefault,
		Mobile:    p.Mobile,
		Name:      p.Name,
		Province:  p.Province,
		City:      p.City,
		Districts: p.Districts,
		Address:   p.Address,
		PostCode:  p.PostCode,
	}
}

func (a *adderessRepo) DeleteAddress(ctx context.Context, r *domain.Address) error {
	var address Address
	result := a.data.db.Where(&Address{ID: r.ID}).First(&address)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.NotFound("ADDRESS_NOT_FOUND", "address not found")
	}

	if result.Error != nil {
		return result.Error
	}

	if address.UserID != r.UserID {
		return errors.NotFound("NOT_USER_ADDRESS_FOR_FOUND", "Is not the address of this user")
	}

	if err := a.data.db.Delete(&address).Error; err != nil {
		return errors.New(500, "DELETE_USER_ADDRESS_ERROR", "用户地址删除失败")
	}
	return nil
}

func (a *adderessRepo) DefaultAddress(ctx context.Context, r *domain.Address) error {
	var address, addressOld Address
	resCurrDefAdr := a.data.db.Where(&Address{UserID: r.UserID, IsDefault: 1}).First(&addressOld)
	if errors.Is(resCurrDefAdr.Error, gorm.ErrRecordNotFound) {
		return errors.NotFound("USER_ADDRESS_NOT_FOUND", "user address not found")
	}

	if resCurrDefAdr.Error == nil {
		addressOld.IsDefault = 0
		addressOld.UpdatedAt = time.Now()
		if err := a.data.db.Save(&addressOld).Error; err != nil {
			return errors.NotFound("SET_ADDRESS_ERROR", "用户地址修改失败")
		}
	}

	result := a.data.db.Where(&Address{ID: r.ID, UserID: r.UserID}).First(&address)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.NotFound("USER_ADDRESS_NOT_FOUND", "user address not found")
	}

	address.IsDefault = 1
	address.UpdatedAt = time.Now()
	if err := a.data.db.Save(&address).Error; err != nil {
		return errors.NotFound("SET_ADDRESS_ERROR", "用户地址修改失败")
	}
	return nil
}

func (a *adderessRepo) UpdateAddress(ctx context.Context, r *domain.Address) error {
	var address Address
	//fmt.Println(r)
	result := a.data.db.Where(&Address{ID: r.ID}).Find(&address)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.NotFound("USER_ADDRESS_NOT_FOUND", "user address not found")
	}

	if result.Error != nil {
		return errors.New(500, "UPDATE_USER_ADDRESS_ERROR", "用户地址更新失败")
	}

	if address.UserID != r.UserID {
		return errors.New(500, "UPDATE_USER_ADDRESS_ERROR", "用户 ID 参数有误")
	}

	address.Address = r.Address
	address.City = r.City
	address.Mobile = r.Mobile
	address.Name = r.Name
	address.IsDefault = r.IsDefault
	address.PostCode = r.PostCode
	address.Districts = r.Districts
	address.Province = r.Province
	address.UpdatedAt = time.Now()

	if err := a.data.db.Save(&address).Error; err != nil {
		return errors.New(500, "UPDATE_USER_ADDRESS_ERROR", "用户地址更新失败")
	}
	return nil
}

func (a *adderessRepo) AddressListByUid(ctx context.Context, uid int64) ([]*domain.Address, error) {
	var address []Address
	result := a.data.db.Where(&Address{UserID: uid}).Find(&address)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("USER_ADDRESS_NOT_FOUND", "user address not found")
	}

	if result.Error != nil {
		return nil, errors.New(500, "USER_ADDRESS_LIST_ERROR", "查询用户地址列表失败")
	}

	if result.RowsAffected == 0 {
		return nil, errors.NotFound("USER_ADDRESS_NOT_FOUND", "user address not found")
	}
	var addressList []*domain.Address
	for _, v := range address {
		addressTmp := modelToBizResponse(v)
		addressList = append(addressList, addressTmp)
	}
	return addressList, nil
}

func (a *adderessRepo) CreateAddress(c context.Context, r *domain.Address) (*domain.Address, error) {
	addInfo := Address{
		UserID:    r.UserID,
		IsDefault: r.IsDefault,
		Mobile:    r.Mobile,
		Name:      r.Name,
		Province:  r.Province,
		City:      r.City,
		Districts: r.Districts,
		Address:   r.Address,
		PostCode:  r.PostCode,
	}

	result := a.data.db.Save(&addInfo)

	if result.Error != nil {
		return nil, errors.NotFound("CREATE_ADDRESS_NOT_FOUND", "创建用户地址失败")
	}

	return modelToBizResponse(addInfo), nil
}

func modelToBizResponse(addInfo Address) *domain.Address {
	return &domain.Address{
		ID:        addInfo.ID,
		UserID:    addInfo.UserID,
		IsDefault: addInfo.IsDefault,
		Mobile:    addInfo.Mobile,
		Name:      addInfo.Name,
		Province:  addInfo.Province,
		City:      addInfo.City,
		Districts: addInfo.Districts,
		Address:   addInfo.Address,
		PostCode:  addInfo.PostCode,
	}
}

func (a *adderessRepo) GetAddress(ctx context.Context, r *domain.Address) (*domain.Address, error) {
	var address Address
	result := a.data.db.Where(&Address{ID: r.ID, UserID: r.UserID}).First(&address)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("USER_ADDRESS_NOT_FOUND", "user address not found")
	}
	return address.ToDomain(), nil
}
