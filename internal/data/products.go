package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Product struct {
	ID                int64           `json:"id"`
	ProductCategory   ProductCategory `json:"product_category"`
	ProductCategoryID int64           `json:"product_category_id"`
	Brand             Brand           `json:"brand"`
	BrandID           int64           `json:"brand_id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Weight            int             `json:"weight"`
	MinimumOrder      int             `json:"minimum_order"`
	Storefront        Storefront      `json:"storefront"`
	StorefrontID      int64           `json:"storefront_id"`
	PreorderDays      int             `json:"preorder_days"`
	Condition         string          `json:"condition"`
	Slug              string          `json:"slug"`
	InsuranceRequired bool            `json:"insurance_required"`
	IsActive          bool            `json:"is_active"`
	CreatedAt         time.Time       `json:"-"`
	UpdatedAt         time.Time       `json:"-"`
}

func ValidateProduct(v *validator.Validator, product *Product) {
	v.Check(product.ProductCategoryID != 0, "product_category_id", "must be provided")
	v.Check(product.ProductCategoryID > 0, "product_category_id", "must be a positive integer")
	v.Check(product.BrandID != 0, "brand_id", "must be provided")
	v.Check(product.BrandID > 0, "brand_id", "must be a positive integer")
	v.Check(product.Name != "", "name", "must be provided")
	v.Check(len(product.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(product.Description != "", "description", "must be provided")
	v.Check(len(product.Description) <= 500, "description", "must not be more than 500 bytes long")
	v.Check(product.Weight != 0, "weight", "must be provided")
	v.Check(product.Weight > 0, "weight", "must be a positive integer")
	v.Check(product.MinimumOrder != 0, "minimum_order", "must be provided")
	v.Check(product.MinimumOrder > 0, "minimum_order", "must be a positive integer")
	v.Check(validator.In(product.Condition, "new", "used"), "condition", "must be either new or used")
}

type ProductModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m ProductModel) GetAll(p Pagination) ([]*Product, Metadata, error) {
	var products []*Product
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Find(&products).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("products").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return products, metadata, nil
}

func (m ProductModel) Get(id int64) (*Product, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var product *Product

	err := m.DB.First(&product, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return product, nil
}

func (m ProductModel) Insert(product *Product) error {
	err := m.DB.Create(&product).Error

	return err
}

func (m ProductModel) Update(p *Product) error {
	var product *Product

	err := m.DB.First(&product, p.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	product.ProductCategoryID = p.ProductCategoryID
	product.BrandID = p.BrandID
	product.Name = p.Name
	product.Description = p.Description
	product.Weight = p.Weight
	product.MinimumOrder = p.MinimumOrder
	product.StorefrontID = p.StorefrontID
	product.PreorderDays = p.PreorderDays
	product.Condition = p.Condition
	product.Slug = p.Slug
	product.InsuranceRequired = p.InsuranceRequired
	product.IsActive = p.IsActive

	err = m.DB.Save(&product).Error
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

func (m ProductModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Delete(&Product{}, id).Error
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
