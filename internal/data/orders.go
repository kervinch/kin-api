package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Order struct {
	ID          int64         `json:"id"`
	UserID      int64         `json:"user_id"`
	User        User          `json:"user"`
	Receiver    string        `json:"receiver"`
	PhoneNumber string        `json:"phone_number"`
	City        string        `json:"city"`
	PostalCode  string        `json:"postal_code"`
	Address     string        `json:"address"`
	Subtotal    int           `json:"subtotal"`
	VoucherID   int64         `json:"voucher_id"`
	Voucher     Voucher       `json:"voucher"`
	Total       int           `json:"total"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"-"`
	UpdatedAt   time.Time     `json:"-"`
	OrderDetail []OrderDetail `json:"order_details"`
}

func ValidateOrder(v *validator.Validator, order *Order) {
	v.Check(order.UserID != 0, "user_id", "must be not be zero")
	v.Check(order.PhoneNumber != "", "phone_number", "must be provided")
	v.Check(order.City != "", "city", "must be provided")
	v.Check(order.PostalCode != "", "postal_code", "must be provided")
	v.Check(order.Address != "", "address", "must be provided")
}

type OrderModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m OrderModel) GetAll(p Pagination) ([]*Order, Metadata, error) {
	var order []*Order
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Order("created_at DESC").Find(&order).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("orders").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return order, metadata, nil
}

func (m OrderModel) Get(id int64) (*Order, error) {
	var order *Order

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return order, nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m OrderModel) InsertWithTx(order *Order, tx *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Create(&order).Error
	if err != nil {
		return err
	}

	return err
}
