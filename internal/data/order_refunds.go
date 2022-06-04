package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type OrderRefund struct {
	ID            int64       `json:"id"`
	UserID        int64       `json:"user_id"`
	GormUser      GormUser    `json:"user" gorm:"foreignKey:UserID"`
	OrderDetailID int64       `json:"order_detail_id"`
	OrderDetail   OrderDetail `json:"order_detail"`
	BrandID       int64       `json:"brand_id"`
	Brand         Brand       `json:"brand"`
	Image1        string      `json:"image_1" gorm:"column:image_1"`
	Image2        string      `json:"image_2" gorm:"column:image_2"`
	Image3        string      `json:"image_3" gorm:"column:image_1"`
	Video         string      `json:"video"`
	Explanation   string      `json:"explanation"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
}

func ValidateOrderRefund(v *validator.Validator, orderRefund *OrderRefund) {
	v.Check(orderRefund.OrderDetailID != 0, "order_detail_id", "must be provided")
	v.Check(orderRefund.BrandID != 0, "brand_id", "must be provided")
	v.Check(orderRefund.Explanation != "", "explanation", "must be provided")
}

type OrderRefundModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m OrderRefundModel) GetAll(p Pagination) ([]*OrderRefund, Metadata, error) {
	var orderRefund []*OrderRefund
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("OrderDetail.Order.Voucher").Preload("Brand").Preload("GormUser").Order("created_at DESC").Find(&orderRefund).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("order_refunds").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return orderRefund, metadata, nil
}

func (m OrderRefundModel) Get(id int64) (*OrderRefund, error) {
	var orderRefund *OrderRefund

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Preload("OrderDetail.Order.Voucher").Preload("Brand").Preload("GormUser").Where("id = ?", id).First(&orderRefund).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return orderRefund, nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m OrderRefundModel) Insert(orderRefund *OrderRefund) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&orderRefund).Error
	if err != nil {
		return err
	}

	return err
}

func (m OrderRefundModel) GetAPI(p Pagination, user *User) ([]*OrderRefund, Metadata, error) {
	var orderRefund []*OrderRefund
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("OrderDetail.Order.Voucher").Preload("Brand").Preload("GormUser").Where("user_id = ?", user.ID).Order("created_at DESC").Find(&orderRefund).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("order_refunds").Where("user_id = ?", user.ID).Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return orderRefund, metadata, nil
}
