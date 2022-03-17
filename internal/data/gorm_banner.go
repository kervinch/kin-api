package data

import (
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type BannerTemp struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	ImageURL    string    `json:"image_url"`
	Title       string    `json:"title"`
	Deeplink    string    `json:"deeplink"`
	OutboundURL string    `json:"outbound_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func GormValidateBanner(v *validator.Validator, banner *Banner) {
	v.Check(banner.ImageURL != "", "image_url", "must be provided")
	v.Check(banner.Title != "", "title", "must be provided")
	v.Check(len(banner.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(len(banner.Deeplink) <= 500, "deeplink", "must not be more than 500 bytes long")
	v.Check(len(banner.OutboundURL) <= 500, "outbound_url", "must not be more than 500 bytes long")
}

type GormBannerModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (g GormBannerModel) GetAll() ([]*Banner, error) {
	var banners []*Banner

	err := g.DB.Find(&banners).Error

	if err != nil {
		return nil, err
	}

	return banners, nil
}

// ====================================================================================
// Business Functions
// ====================================================================================
