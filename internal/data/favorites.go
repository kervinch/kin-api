package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Favorite struct {
	ID              int64         `json:"id"`
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

func (m FavoriteModel) GetAll(p Pagination, user *User) ([]*Favorite, Metadata, error) {
	var favorites []*Favorite
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("ProductDetail").Where("user_id = ?", user.ID).Order("id").Find(&favorites).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("favorites").Where("user_id = ?", user.ID).Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return favorites, metadata, nil
}

func (m FavoriteModel) Get(id int64, user *User) (*Favorite, error) {
	var favorite *Favorite

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, user.ID).First(&favorite).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return favorite, nil
}

func (m FavoriteModel) Insert(favorite *Favorite) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&favorite).Error

	return err
}

func (m FavoriteModel) Delete(id int64, user *User) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, user.ID).Delete(&Favorite{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}
