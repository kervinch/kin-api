package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type ProductVideo struct {
	ID              int64         `json:"id"`
	ProductDetail   ProductDetail `json:"product_detail"`
	ProductDetailID int64         `json:"product_detail_id"`
	VideoURL        string        `json:"video_url"`
	IsMain          bool          `json:"is_main"`
	IsActive        bool          `json:"is_active"`
	CreatedAt       time.Time     `json:"-"`
	UpdatedAt       time.Time     `json:"-"`
}

func ValidateProductVideo(v *validator.Validator, productVideo *ProductVideo) {
	v.Check(productVideo.ProductDetailID != 0, "product_detail_id", "must be provided")
	v.Check(productVideo.ProductDetailID > 0, "product_detail_id", "must be a positive integer")
	v.Check(productVideo.VideoURL != "", "video", "must be provided")
}

type ProductVideoModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m ProductVideoModel) GetAll(p Pagination) ([]*ProductVideo, Metadata, error) {
	var productVideos []*ProductVideo
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Find(&productVideos).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("product_videos").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return productVideos, metadata, nil
}

func (m ProductVideoModel) Get(id int64) (*ProductVideo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var productVideo *ProductVideo

	err := m.DB.First(&productVideo, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return productVideo, nil
}

func (m ProductVideoModel) Insert(productVideo *ProductVideo) error {
	err := m.DB.Create(&productVideo).Error

	return err
}

func (m ProductVideoModel) Update(p *ProductVideo) error {
	var productVideo *ProductVideo

	err := m.DB.First(&productVideo, p.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	productVideo.ProductDetailID = p.ProductDetailID
	productVideo.VideoURL = p.VideoURL
	productVideo.IsMain = p.IsMain
	productVideo.IsActive = p.IsActive

	err = m.DB.Save(&productVideo).Error
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

func (m ProductVideoModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Delete(&ProductVideo{}, id).Error
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
