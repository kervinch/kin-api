package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type ProductImage struct {
	ID              int64         `json:"id"`
	ProductDetail   ProductDetail `json:"product_detail"`
	ProductDetailID int64         `json:"product_detail_id"`
	ImageURL        string        `json:"image_url"`
	IsMain          bool          `json:"is_main"`
	IsActive        bool          `json:"is_active"`
	CreatedAt       time.Time     `json:"-"`
	UpdatedAt       time.Time     `json:"-"`
}

func ValidateProductImage(v *validator.Validator, productImage *ProductImage) {
	v.Check(productImage.ProductDetailID != 0, "product_detail_id", "must be provided")
	v.Check(productImage.ProductDetailID > 0, "product_detail_id", "must be a positive integer")
	v.Check(productImage.ImageURL != "", "image", "must be provided")
}

type ProductImageModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m ProductImageModel) GetAll(p Pagination) ([]*ProductImage, Metadata, error) {
	var productImages []*ProductImage
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Find(&productImages).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("product_images").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return productImages, metadata, nil
}

func (m ProductImageModel) Get(id int64) (*ProductImage, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var productImage *ProductImage

	err := m.DB.First(&productImage, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return productImage, nil
}

func (m ProductImageModel) Insert(productImage *ProductImage) error {
	err := m.DB.Create(&productImage).Error

	return err
}

func (m ProductImageModel) Update(p *ProductImage) error {
	var productImage *ProductImage

	err := m.DB.First(&productImage, p.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	productImage.ProductDetailID = p.ProductDetailID
	productImage.ImageURL = p.ImageURL
	productImage.IsMain = p.IsMain
	productImage.IsActive = p.IsActive

	err = m.DB.Save(&productImage).Error
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

func (m ProductImageModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Delete(&ProductImage{}, id).Error
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
