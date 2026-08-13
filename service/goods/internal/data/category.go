package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
	"goods/internal/biz"
	"gorm.io/gorm"
	"time"
)

// Category 商品分类表
type Category struct {
	ID               int32          `gorm:"primarykey;type:int" json:"id"`
	Name             string         `gorm:"type:varchar(50);not null;comment:分类名称" json:"name"`
	ParentCategoryID int32          `json:"parent_id"`
	ParentCategory   *Category      `json:"-"`
	SubCategory      []*Category    `gorm:"foreignKey:ParentCategoryID;references:ID" json:"sub_category"`
	Level            int32          `gorm:"column:level;default:1;not null;type:int;comment:分类的级别" json:"level"`
	IsTab            bool           `gorm:"comment:是否显示;default:false" json:"is_tab"`
	Sort             int32          `gorm:"comment:分类排序;default:99;not null;type:int" json:"sort"`
	CreatedAt        time.Time      `gorm:"column:add_time" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:update_time" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at"`
}

// GoodsCategoryBrand  商品和分类多对对的表
type GoodsCategoryBrand struct {
	ID         int32          `gorm:"primarykey;type:int" json:"id"` // bigint
	CategoryID int32          `gorm:"type:int;index:idx_category_brand,unique;comment:商品和分类联合索引唯一"`
	BrandsID   int32          `gorm:"type:int;index:idx_category_brand,unique:comment:商品和分类联合索引唯一"`
	CreatedAt  time.Time      `gorm:"column:add_time" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:update_time" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

type CategoryRepo struct {
	data *Data
	log  *log.Helper
}

// NewCategoryRepo .
func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &CategoryRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *CategoryRepo) DeleteCategory(ctx context.Context, id int32) error {
	if res := r.data.db.Delete(&Category{}, id); res.RowsAffected == 0 {
		return errors.InternalServer("DELETE_CATGORY_ERROR", res.Error.Error())
	}
	return nil
}

func (r *CategoryRepo) UpdateCategory(ctx context.Context, req *biz.CategoryInfo) error {
	var category Category
	if result := r.data.db.First(&category, req.ID); result.RowsAffected == 0 {
		return errors.NotFound("CATEGORY_NOT_FOUND", "商品分类不存在")
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.ParentCategory != 0 {
		category.ParentCategoryID = req.ParentCategory
	}
	if req.Level != 0 {
		category.Level = req.Level
	}
	if req.IsTab {
		category.IsTab = req.IsTab
	}
	result := r.data.db.Save(&category)
	if result.Error != nil {
		return errors.InternalServer("CATEGORY_UPDATE_ERROR", "商品分类创建失败")
	}
	return nil
}

func (r *CategoryRepo) AddCategory(ctx context.Context, req *biz.CategoryInfo) (*biz.CategoryInfo, error) {
	cat := Category{
		Name:  req.Name,
		Level: req.Level,
		IsTab: req.IsTab,
		Sort:  req.Sort,
	}

	// 去查询父类目是否存在
	if req.Level != 1 {
		var categories Category
		if res := r.data.db.First(&categories, req.ParentCategory); res.RowsAffected == 0 {
			return nil, errors.NotFound("CATEGORY_NOT_FOUND", "商品分类不存在")
		}
		cat.ParentCategoryID = req.ParentCategory
	}

	result := r.data.db.Create(&cat)
	if result.Error != nil {
		return nil, errors.InternalServer("CATEGORY_CREATE_ERROR", result.Error.Error())
	}
	res := &biz.CategoryInfo{
		ID:             cat.ID,
		Name:           cat.Name,
		ParentCategory: cat.ParentCategoryID,
		Level:          cat.Level,
		IsTab:          cat.IsTab,
		Sort:           cat.Sort,
	}
	return res, nil
}

func (r *CategoryRepo) Category(ctx context.Context) ([]*biz.Category, error) {
	var cate []*Category
	result := r.data.db.Where(&Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&cate)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("CATEGORY_NOT_FOUND", "分类不存在")
	}
	if result.Error != nil {
		return nil, errors.InternalServer("CATEGORY_NOT_FOUND", result.Error.Error())
	}

	var res []*biz.Category
	err := copier.Copy(&res, &cate)
	if err != nil {
		return nil, errors.InternalServer("CATEGORY_COPY_ERROR", err.Error())
	}
	return res, nil
}

func (r *CategoryRepo) GetCategoryByID(ctx context.Context, id int32) (*biz.CategoryInfo, error) {
	var categories Category
	if res := r.data.db.First(&categories, id); res.RowsAffected == 0 {
		return nil, errors.NotFound("CATEGORY_NOT_FOUND", "分类不存在")
	}

	info := &biz.CategoryInfo{
		ID:             categories.ID,
		Name:           categories.Name,
		ParentCategory: categories.ParentCategoryID,
		Level:          categories.Level,
		IsTab:          categories.IsTab,
		Sort:           categories.Sort,
	}
	return info, nil
}

func (r *CategoryRepo) SubCategory(ctx context.Context, req biz.CategoryInfo) ([]*biz.CategoryInfo, error) {
	var subCategory []Category
	var subCategoryInfo []*biz.CategoryInfo
	preload := "SubCategory"
	if req.Level == 1 {
		preload = "SubCategory.SubCategory"
	}

	if err := r.data.db.Where(&Category{ParentCategoryID: req.ID}).Preload(preload).Find(&subCategory).Error; err != nil {
		return nil, errors.NotFound("CATEGORY_NOT_FOUND", "分类不存在")
	}
	for _, v := range subCategory {
		subCategoryInfo = append(subCategoryInfo, &biz.CategoryInfo{
			ID:             v.ID,
			Name:           v.Name,
			ParentCategory: v.ParentCategoryID,
			Level:          v.Level,
			IsTab:          v.IsTab,
			Sort:           v.Sort,
		})
	}

	return subCategoryInfo, nil
}

func (r *CategoryRepo) GetCategoryAll(ctx context.Context, level, id int32) ([]interface{}, error) {
	categoryIds := make([]interface{}, 0)
	var subQuery string
	// 把一级级分类下的所有三级分类都拿到
	if level == 1 {
		subQuery = fmt.Sprintf("SELECT id FROM categories WHERE parent_category_id IN (SELECT id FROM categories WHERE parent_category_id=%d)", id)
	} else if level == 2 {
		subQuery = fmt.Sprintf("SELECT id FROM categories WHERE parent_category_id=%d", id)
	} else if level == 3 {
		subQuery = fmt.Sprintf("SELECT id FROM categories WHERE id=%d", id)
	}

	type Result struct {
		ID int32
	}

	var results []Result
	if err := r.data.db.Table("categories").Model(Category{}).Raw(subQuery).Scan(&results).Error; err != nil {
		return nil, errors.InternalServer("CATEGORY_ERROR", err.Error())
	}
	for _, re := range results {
		categoryIds = append(categoryIds, re.ID)
	}
	return categoryIds, nil
}
