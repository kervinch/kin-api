package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type ProductDetail struct {
	ID           int64          `json:"id"`
	Product      Product        `json:"-"`
	ProductID    int64          `json:"product_id"`
	Color        string         `json:"color"`
	Size         string         `json:"size"`
	Price        int64          `json:"price"`
	SKU          string         `json:"sku"`
	Stock        int            `json:"stock"`
	IsActive     bool           `json:"is_active"`
	ProductImage []ProductImage `json:"product_images"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
}

func ValidateProductDetail(v *validator.Validator, productDetail *ProductDetail) {
	v.Check(productDetail.ProductID != 0, "product_id", "must be provided")
	v.Check(productDetail.ProductID > 0, "product_id", "must be a positive integer")
	v.Check(productDetail.Price > 0, "price", "must be a positive integer")
	v.Check(productDetail.Stock > 0, "stock", "must be a positive integer")
}

type ProductDetailModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m ProductDetailModel) GetAll(p Pagination) ([]*ProductDetail, Metadata, error) {
	var productDetails []*ProductDetail
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("ProductImage").Find(&productDetails).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("product_details").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return productDetails, metadata, nil
}

func (m ProductDetailModel) Get(id int64) (*ProductDetail, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var productDetail *ProductDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Preload("ProductImage").First(&productDetail, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return productDetail, nil
}

func (m ProductDetailModel) Insert(productDetail *ProductDetail) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&productDetail).Error

	productDetailID := productDetail.ID

	return productDetailID, err
}

func (m ProductDetailModel) InsertWithTx(productDetail *ProductDetail, tx *gorm.DB) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Create(&productDetail).Error

	productDetailID := productDetail.ID

	return productDetailID, err
}

func (m ProductDetailModel) Update(p *ProductDetail) error {
	var productDetail *ProductDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).First(&productDetail, p.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	productDetail.ProductID = p.ProductID
	productDetail.Color = p.Color
	productDetail.Size = p.Size
	productDetail.Price = p.Price
	productDetail.SKU = p.SKU
	productDetail.Stock = p.Stock
	productDetail.IsActive = p.IsActive

	err = m.DB.WithContext(ctx).Save(&productDetail).Error
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

func (m ProductDetailModel) UpdateWithTx(p *ProductDetail, tx *gorm.DB) error {
	var productDetail *ProductDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).First(&productDetail, p.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	productDetail.ProductID = p.ProductID
	productDetail.Color = p.Color
	productDetail.Size = p.Size
	productDetail.Price = p.Price
	productDetail.SKU = p.SKU
	productDetail.Stock = p.Stock
	productDetail.IsActive = p.IsActive

	err = tx.WithContext(ctx).Save(&productDetail).Error
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

func (m ProductDetailModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Delete(&ProductDetail{}, id).Error
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

func (m ProductDetailModel) DeleteWithTx(id int64, tx *gorm.DB) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Delete(&ProductDetail{}, id).Error
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
