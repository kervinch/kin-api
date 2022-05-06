package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Storefront struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	Slug        string     `json:"slug"`
	Product     []*Product `gorm:"many2many:product_storefront_subscriptions"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"-"`
	UpdatedAt   time.Time  `json:"-"`
}

func ValidateStorefront(v *validator.Validator, storefront *Storefront) {
	v.Check(storefront.ImageURL != "", "image_url", "must be provided")
	v.Check(storefront.Name != "", "name", "must be provided")
	v.Check(len(storefront.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(storefront.Description != "", "description", "must be provided")
	v.Check(len(storefront.Description) <= 500, "description", "must not be more than 500 bytes long")
}

type StorefrontModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m StorefrontModel) GetAll(p Pagination) ([]*Storefront, Metadata, error) {
	var storefronts []*Storefront
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Find(&storefronts).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("storefronts").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return storefronts, metadata, nil
}

func (m StorefrontModel) Get(id int64) (*Storefront, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var storefront *Storefront

	err := m.DB.WithContext(ctx).First(&storefront, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return storefront, nil
}

func (m StorefrontModel) Insert(storefront *Storefront) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&storefront).Error
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "storefronts_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return err
}

func (m StorefrontModel) Update(s *Storefront) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var storefront *Storefront

	err := m.DB.WithContext(ctx).First(&storefront, s.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	storefront.Name = s.Name
	storefront.Description = s.Description
	storefront.ImageURL = s.ImageURL
	storefront.Slug = s.Slug
	storefront.IsActive = s.IsActive

	err = m.DB.Save(&storefront).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrEditConflict
		case err.Error() == `pq: duplicate key value violates unique constraint "storefronts_slug_key"`:
			return ErrDuplicateSlug
		default:
			return err
		}
	}

	return nil
}

func (m StorefrontModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Delete(&Storefront{}, id).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m StorefrontModel) GetBySlugWithProducts(slug string) (*Storefront, error) {
	if slug == "" {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var storefront *Storefront

	err := m.DB.WithContext(ctx).Where("is_active = ? AND slug = ?", true, slug).Preload("Product.ProductCategory").Preload("Product.Brand").Preload("Product.Storefront").Preload("Product.ProductDetail.ProductImage").First(&storefront).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return storefront, nil
}
