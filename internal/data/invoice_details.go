package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type InvoiceDetail struct {
	ID              int64         `json:"id"`
	OrderDetailID   int64         `json:"order_detail_id"`
	OrderDetail     OrderDetail   `json:"order_detail"`
	ProductDetailID int64         `json:"product_detail_id"`
	ProductDetail   ProductDetail `json:"product_detail"`
	ProductName     string        `json:"product_name"`
	Quantity        int           `json:"quantity"`
	Price           int           `json:"price"`
	Status          string        `json:"status"`
	CreatedAt       time.Time     `json:"-"`
	UpdatedAt       time.Time     `json:"-"`
}

func ValidateInvoiceDetail(v *validator.Validator, invoiceDetail *InvoiceDetail) {
	v.Check(invoiceDetail.OrderDetailID != 0, "order_detail_id", "must be not be zero")
	v.Check(invoiceDetail.ProductDetailID != 0, "product_detail_id", "must be provided")
	v.Check(invoiceDetail.ProductName != "", "product_name", "must be provided")
	v.Check(invoiceDetail.Quantity != 0, "quantity", "must not be 0")
}

type InvoiceDetailModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m InvoiceDetailModel) GetAll(p Pagination) ([]*InvoiceDetail, Metadata, error) {
	var invoiceDetail []*InvoiceDetail
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Order("created_at DESC").Find(&invoiceDetail).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("invoice_details").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return invoiceDetail, metadata, nil
}

func (m InvoiceDetailModel) Get(id int64) (*InvoiceDetail, error) {
	var invoiceDetail *InvoiceDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).First(&invoiceDetail).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return invoiceDetail, nil
}

func (m InvoiceDetailModel) GetAllByOrderDetailID(id int64) ([]*InvoiceDetail, error) {
	var invoiceDetail []*InvoiceDetail

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("order_detail_id = ?", id).Find(&invoiceDetail).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return invoiceDetail, nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m InvoiceDetailModel) InsertWithTx(invoiceDetail *InvoiceDetail, tx *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Create(&invoiceDetail).Error
	if err != nil {
		return err
	}

	return err
}
