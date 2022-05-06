package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Brand struct {
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

func ValidateBrand(v *validator.Validator, brand *Brand) {
	v.Check(brand.ImageURL != "", "image_url", "must be provided")
	v.Check(brand.Name != "", "name", "must be provided")
	v.Check(len(brand.Name) <= 500, "name", "must not be more than 500 bytes long")
}

type BrandModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m BrandModel) GetAll(p Pagination) ([]*Brand, Metadata, error) {
	var brands []*Brand
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Find(&brands).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("brands").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return brands, metadata, nil
}

func (m BrandModel) Get(id int64) (*Brand, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var brand *Brand

	err := m.DB.WithContext(ctx).First(&brand, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return brand, nil
}

func (m BrandModel) Insert(brand *Brand) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&brand).Error
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "brands_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return err
}

func (m BrandModel) Update(b *Brand) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var brand *Brand

	err := m.DB.WithContext(ctx).First(&brand, b.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	brand.ImageURL = b.ImageURL
	brand.Name = b.Name
	brand.OrderNumber = b.OrderNumber
	brand.IsActive = b.IsActive

	err = m.DB.WithContext(ctx).Save(&brand).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrEditConflict
		case err.Error() == `pq: duplicate key value violates unique constraint "brands_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return nil
}

func (m BrandModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Delete(&Brand{}, id).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m BrandModel) GetAPI() ([]*Brand, error) {
	var brands []*Brand

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("is_active", true).Order("order_number").Limit(8).Find(&brands).Error
	if err != nil {
		return nil, err
	}

	return brands, nil
}

func (m BrandModel) GetBySlugWithProducts(slug string) (*Brand, error) {
	if slug == "" {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var brand *Brand

	err := m.DB.WithContext(ctx).Where("is_active = ? AND slug = ?", true, slug).Preload("Product.ProductCategory").Preload("Product.Brand").Preload("Product.Storefront").Preload("Product.ProductDetail.ProductImage").First(&brand).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return brand, nil
}
