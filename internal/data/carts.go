package data

import (
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Cart struct {
	ID              int64         `json:"id"`
	User            User          `json:"user"`
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

func (m CartModel) Insert(cart *Cart) error {
	err := m.DB.Create(&cart).Error

	return err
}

func (m CartModel) Update(c *Cart) error {
	var cart *Cart

	err := m.DB.First(&cart, c.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	cart.Quantity = c.Quantity

	err = m.DB.Save(&cart).Error
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

func (m CartModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Delete(&Cart{}, id).Error
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
