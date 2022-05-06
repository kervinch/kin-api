package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type ProductCategory struct {
	ID          int64     `json:"id"`
	ImageURL    string    `json:"image_url"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	OrderNumber int       `json:"order_number"`
	IsActive    bool      `json:"is_active"`
	Product     []Product `json:"products"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func ValidateProductCategory(v *validator.Validator, productCategory *ProductCategory) {
	v.Check(productCategory.Name != "", "name", "must be provided")
	v.Check(len(productCategory.Name) <= 100, "name", "must not be more than 100 bytes long")
}

type ProductCategoryModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m ProductCategoryModel) GetAll(p Pagination) ([]*ProductCategory, Metadata, error) {
	var productCategories []*ProductCategory
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Find(&productCategories).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("product_categories").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return productCategories, metadata, nil
}

func (m ProductCategoryModel) Get(id int64) (*ProductCategory, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var productCategory *ProductCategory

	err := m.DB.WithContext(ctx).First(&productCategory, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return productCategory, nil
}

func (m ProductCategoryModel) Insert(productCategory *ProductCategory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&productCategory).Error
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "product_categories_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return err
}

func (m ProductCategoryModel) Update(p *ProductCategory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var productCategory *ProductCategory

	err := m.DB.WithContext(ctx).First(&productCategory, p.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	productCategory.ImageURL = p.ImageURL
	productCategory.Name = p.Name
	productCategory.Slug = p.Slug
	productCategory.IsActive = p.IsActive
	productCategory.OrderNumber = p.OrderNumber

	err = m.DB.Save(&productCategory).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrEditConflict
		case err.Error() == `pq: duplicate key value violates unique constraint "product_categories_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return nil
}

func (m ProductCategoryModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Delete(&ProductCategory{}, id).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m ProductCategoryModel) GetAPI() ([]*ProductCategory, error) {
	var productCategories []*ProductCategory

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("is_active = ?", true).Order("order_number").Limit(6).Find(&productCategories).Error
	if err != nil {
		return nil, err
	}

	return productCategories, nil
}

func (m ProductCategoryModel) GetBySlug(slug string) (*ProductCategory, error) {
	if slug == "" {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var productCategory *ProductCategory

	err := m.DB.WithContext(ctx).Where("is_active = ? AND slug = ?", true, slug).First(&productCategory).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return productCategory, nil
}

func (m ProductCategoryModel) GetBySlugWithProducts(slug string) (*ProductCategory, error) {
	if slug == "" {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var productCategory *ProductCategory

	err := m.DB.WithContext(ctx).Where("is_active = ? AND slug = ?", true, slug).Preload("Product.ProductCategory").Preload("Product.Brand").Preload("Product.Storefront").Preload("Product.ProductDetail.ProductImage").First(&productCategory).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return productCategory, nil
}
