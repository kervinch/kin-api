package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type BlogCategory struct {
	ID          int64     `json:"id"`
	Image       string    `json:"image"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	OrderNumber int       `json:"order_number"`
	DeletedAt   time.Time `json:"-"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func ValidateBlogCategory(v *validator.Validator, blogCategory *BlogCategory) {
	v.Check(blogCategory.Image != "", "image", "must be provided")
	v.Check(blogCategory.Name != "", "name", "must be provided")
	v.Check(len(blogCategory.Name) <= 100, "name", "must not be more than 100 bytes long")
}

type BlogCategoryModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m BlogCategoryModel) GetAll(p Pagination) ([]*BlogCategory, Metadata, error) {
	var blogCategories []*BlogCategory
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Where("status = ?", "active").Find(&blogCategories).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("blog_categories").Where("status = ?", "active").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return blogCategories, metadata, nil
}

func (m BlogCategoryModel) Get(id int64) (*BlogCategory, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var blogCategory *BlogCategory

	err := m.DB.First(&blogCategory, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return blogCategory, nil
}

func (m BlogCategoryModel) Insert(blogCategory *BlogCategory) error {
	err := m.DB.Create(&blogCategory).Error

	return err
}

func (m BlogCategoryModel) Update(b *BlogCategory) error {
	var blogCategory *BlogCategory

	err := m.DB.First(&blogCategory, b.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	blogCategory.Image = b.Image
	blogCategory.Name = b.Name
	blogCategory.Slug = b.Slug
	blogCategory.Type = b.Type
	blogCategory.Status = b.Status
	blogCategory.OrderNumber = b.OrderNumber

	err = m.DB.Save(&blogCategory).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m BlogCategoryModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Delete(&BlogCategory{}, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m BlogCategoryModel) GetAPI() ([]*BlogCategory, error) {
	var blogCategories []*BlogCategory

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("status = ?", "active").Order("order_number").Limit(6).Find(&blogCategories).Error
	if err != nil {
		return nil, err
	}

	return blogCategories, nil
}

func (m BlogCategoryModel) GetBySlug(slug string) (*BlogCategory, error) {
	if slug == "" {
		return nil, ErrRecordNotFound
	}

	var blogCategory *BlogCategory

	err := m.DB.Where("slug = ?", slug).First(&blogCategory).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return blogCategory, nil
}
