package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type Category struct {
	ID               int32
	Name             string
	ParentCategoryID int32
	SubCategory      []*Category
	Level            int32
	IsTab            bool
	Sort             int32
}

type CategoryList struct {
	Category    *CategoryInfo
	SubCategory []*CategoryInfo
}

type CategoryInfo struct {
	ID             int32
	Name           string
	ParentCategory int32
	Level          int32
	IsTab          bool
	Sort           int32
}

type CategoryRepo interface {
	AddCategory(context.Context, *CategoryInfo) (*CategoryInfo, error)
	UpdateCategory(context.Context, *CategoryInfo) error
	Category(context.Context) ([]*Category, error)
	GetCategoryByID(ctx context.Context, id int32) (*CategoryInfo, error)
	SubCategory(context.Context, CategoryInfo) ([]*CategoryInfo, error)
	DeleteCategory(context.Context, int32) error
	GetCategoryAll(context.Context, int32, int32) ([]interface{}, error)
}

type CategoryUsecase struct {
	repo CategoryRepo
	log  *log.Helper
}

func NewCategoryUsecase(repo CategoryRepo, logger log.Logger) *CategoryUsecase {
	return &CategoryUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (c *CategoryUsecase) DeleteCategory(ctx context.Context, r *CategoryInfo) error {
	// todo 需要验证是否是定级分类,定级分类下面还有没有二级分类
	err := c.repo.DeleteCategory(ctx, r.ID)
	if err != nil {
		return err
	}
	return nil
}

func (c *CategoryUsecase) UpdateCategory(ctx context.Context, r *CategoryInfo) error {
	err := c.repo.UpdateCategory(ctx, r)
	return err
}

func (c *CategoryUsecase) CreateCategory(ctx context.Context, r *CategoryInfo) (*CategoryInfo, error) {
	cateInfo, err := c.repo.AddCategory(ctx, r)
	if err != nil {
		return nil, err
	}
	return cateInfo, nil
}

func (c *CategoryUsecase) CategoryList(ctx context.Context) ([]*Category, error) {
	return c.repo.Category(ctx)
}

func (c *CategoryUsecase) SubCategoryList(ctx context.Context, cid int32) (*CategoryList, error) {
	cateInfo, err := c.repo.GetCategoryByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	category, err := c.repo.SubCategory(ctx, *cateInfo)
	if err != nil {
		return nil, err
	}

	return &CategoryList{
		Category:    cateInfo,
		SubCategory: category,
	}, nil
}

func (c *CategoryUsecase) GetCategoryByID(ctx context.Context, id int32) (*CategoryInfo, error) {
	return c.repo.GetCategoryByID(ctx, id)
}
