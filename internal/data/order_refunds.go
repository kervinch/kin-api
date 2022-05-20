package data

import (
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type OrderRefund struct {
	ID            int64       `json:"id"`
	OrderDetailID int64       `json:"order_detail_id"`
	OrderDetail   OrderDetail `json:"order_detail"`
	BrandID       int64       `json:"brand_id"`
	Brand         Brand       `json:"brand"`
	Image1        string      `json:"image_1"`
	Image2        string      `json:"image_2"`
	Image3        string      `json:"image_3"`
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
