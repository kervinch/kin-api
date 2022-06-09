package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type OrderDetail struct {
	ID            int64           `json:"id"`
	OrderID       int64           `json:"order_id"`
	Order         Order           `json:"order"`
	BrandID       int64           `json:"brand_id"`
	Brand         Brand           `json:"brand"`
	InvoiceNumber string          `json:"invoice_number"`
	Subtotal      int64           `json:"subtotal"`
	VoucherID     sql.NullInt64   `json:"voucher_id"`
	Voucher       Voucher         `json:"voucher"`
	Total         int64           `json:"total"`
	Status        string          `json:"status"`
	CreatedAt     time.Time       `json:"-"`
	UpdatedAt     time.Time       `json:"-"`
	InvoiceDetail []InvoiceDetail `json:"invoice_details"`
}

func ValidateOrderDetail(v *validator.Validator, orderDetail *OrderDetail) {
	v.Check(orderDetail.OrderID != 0, "order_id", "must be not be zero")
	v.Check(orderDetail.BrandID != 0, "brand_id", "must be provided")
	v.Check(orderDetail.InvoiceNumber != "", "invoice_number", "must be provided")
	v.Check(validator.In(orderDetail.Status, "awaiting_payment", "expired", "paid", "pending", "processing", "delivery", "completed", "refund_requested", "refund_rejected", "refund_completed"), "status", "must be valid to enum defined")
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

func (m OrderDetailModel) GetWithTx(id int64, tx *gorm.DB) (*OrderDetail, error) {
	var orderDetail *OrderDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Preload("InvoiceDetail").Where("id = ?", id).First(&orderDetail).Error
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

func (m OrderDetailModel) InsertWithTx(orderDetail *OrderDetail, tx *gorm.DB) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Create(&orderDetail).Error
	if err != nil {
		return 0, err
	}

	orderDetailID := orderDetail.ID

	return orderDetailID, err
}

func (m OrderDetailModel) SetTotalWithTx(id int64, subtotal int64, total int64, tx *gorm.DB) error {
	var orderDetail *OrderDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).First(&orderDetail, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	orderDetail.Subtotal = subtotal
	orderDetail.Total = total

	err = tx.WithContext(ctx).Save(&orderDetail).Error
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

func (m OrderDetailModel) SetTotalWithVoucherAndTx(id int64, subtotal int64, voucherID int64, total int64, tx *gorm.DB) error {
	var orderDetail *OrderDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).First(&orderDetail, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	orderDetail.Subtotal = subtotal
	orderDetail.VoucherID = sql.NullInt64{Int64: voucherID, Valid: true}
	orderDetail.Total = total

	err = tx.WithContext(ctx).Save(&orderDetail).Error
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

func (m OrderDetailModel) UpdateStatusByOrderID(orderID int64, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Model(&OrderDetail{}).Where("order_id = ?", orderID).Update("status", status).Error
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
