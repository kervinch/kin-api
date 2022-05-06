package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Cart struct {
	ID              int64         `json:"id"`
	UserID          int64         `json:"user_id"`
	ProductDetail   ProductDetail `json:"product_detail"`
	ProductDetailID int64         `json:"product_detail_id"`
	Quantity        int           `json:"quantity"`
	CreatedAt       time.Time     `json:"-"`
	UpdatedAt       time.Time     `json:"-"`
}

func ValidateCart(v *validator.Validator, cart *Cart) {
	v.Check(cart.UserID != 0, "user_id", "must be provided")
	v.Check(cart.UserID > 0, "user_id", "must be a positive integer")
	v.Check(cart.ProductDetailID != 0, "product_detail_id", "must be provided")
	v.Check(cart.ProductDetailID > 0, "product_detail_id", "must be a positive integer")
	v.Check(cart.Quantity != 0, "quantity", "must be provided")
	v.Check(cart.Quantity > 0, "quantity", "must be a positive integer")
}

type CartModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m CartModel) GetAll(p Pagination, user *User) ([]*Cart, Metadata, error) {
	var carts []*Cart
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("ProductDetail").Where("user_id = ?", user.ID).Order("id").Find(&carts).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("carts").Where("user_id = ?", user.ID).Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return carts, metadata, nil
}

func (m CartModel) Get(id int64, user *User) (*Cart, error) {
	var cart *Cart

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, user.ID).First(&cart).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return cart, nil
}

func (m CartModel) Insert(cart *Cart) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&cart).Error

	return err
}

func (m CartModel) Update(c *Cart) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var cart *Cart

	err := m.DB.WithContext(ctx).First(&cart, c.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	cart.Quantity = c.Quantity

	err = m.DB.WithContext(ctx).Save(&cart).Error
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

func (m CartModel) Delete(id int64, user *User) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, user.ID).Delete(&Cart{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}
