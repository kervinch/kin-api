package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type OrderDetail struct {
	ID            int64     `json:"id"`
	OrderID       int64     `json:"order_id"`
	Order         Order     `json:"order"`
	BrandID       int64     `json:"brand_id"`
	Brand         Brand     `json:"brand"`
	InvoiceNumber string    `json:"invoice_number"`
	Subtotal      int       `json:"subtotal"`
	VoucherID     int64     `json:"voucher_id"`
	Voucher       Voucher   `json:"voucher"`
	Total         int       `json:"total"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

func ValidateOrderDetail(v *validator.Validator, orderDetail *OrderDetail) {
	v.Check(orderDetail.OrderID != 0, "order_id", "must be not be zero")
	v.Check(orderDetail.BrandID != 0, "brand_id", "must be provided")
	v.Check(orderDetail.InvoiceNumber != "", "invoice_number", "must be provided")
}

type OrderDetailModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m OrderDetailModel) GetAll(p Pagination) ([]*OrderDetail, Metadata, error) {
	var orderDetail []*OrderDetail
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Order("created_at DESC").Find(&orderDetail).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("order_details").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return orderDetail, metadata, nil
}

func (m OrderDetailModel) Get(id int64) (*OrderDetail, error) {
	var orderDetail *OrderDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).First(&orderDetail).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return orderDetail, nil
}

func (m OrderDetailModel) GetAllByBrandID(id int64) ([]*OrderDetail, error) {
	var orderDetail []*OrderDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("brand_id = ?", id).Find(&orderDetail).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return orderDetail, nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m OrderDetailModel) InsertWithTx(orderDetail *OrderDetail, tx *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Create(&orderDetail).Error
	if err != nil {
		return err
	}

	return err
}
