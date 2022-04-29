package data

import (
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Favorite struct {
	ID              int64         `json:"id"`
	User            User          `json:"user"`
	UserID          int64         `json:"user_id"`
	ProductDetail   ProductDetail `json:"product_detail"`
	ProductDetailID int64         `json:"product_detail_id"`
	CreatedAt       time.Time     `json:"-"`
	UpdatedAt       time.Time     `json:"-"`
}

func ValidateFavorite(v *validator.Validator, favorite *Favorite) {
	v.Check(favorite.UserID != 0, "user_id", "must be provided")
	v.Check(favorite.UserID > 0, "user_id", "must be a positive integer")
	v.Check(favorite.ProductDetailID != 0, "product_detail_id", "must be provided")
	v.Check(favorite.ProductDetailID > 0, "product_detail_id", "must be a positive integer")
}

type FavoriteModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m FavoriteModel) Insert(favorite *Favorite) error {
	err := m.DB.Create(&favorite).Error

	return err
}

func (m FavoriteModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Delete(&Favorite{}, id).Error
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
