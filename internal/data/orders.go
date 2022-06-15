package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Order struct {
	ID          int64         `json:"id"`
	UserID      int64         `json:"user_id"`
	GormUser    GormUser      `json:"user" gorm:"foreignKey:UserID"`
	Receiver    string        `json:"receiver"`
	PhoneNumber string        `json:"phone_number"`
	City        string        `json:"city"`
	PostalCode  string        `json:"postal_code"`
	Address     string        `json:"address"`
	Subtotal    int64         `json:"subtotal"`
	VoucherID   sql.NullInt64 `json:"voucher_id"`
	Voucher     Voucher       `json:"voucher"`
	Total       int64         `json:"total"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"-"`
	UpdatedAt   time.Time     `json:"-"`
	OrderDetail []OrderDetail `json:"order_details"`
}

func ValidateOrder(v *validator.Validator, order *Order) {
	v.Check(order.UserID != 0, "user_id", "must be not be zero")
	v.Check(order.Receiver != "", "receiver", "must be provided")
	v.Check(order.PhoneNumber != "", "phone_number", "must be provided")
	v.Check(order.City != "", "city", "must be provided")
	v.Check(order.PostalCode != "", "postal_code", "must be provided")
	v.Check(order.Address != "", "address", "must be provided")
	v.Check(validator.In(order.Status, "awaiting_payment", "expired", "paid", "pending", "processing", "delivery", "completed", "refund_requested", "refund_rejected", "refund_completed"), "status", "must be valid to enum defined")
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

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("Voucher").Preload("GormUser").Order("created_at DESC").Find(&order).Error
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

	err := m.DB.WithContext(ctx).Where("id = ?", id).Preload("Voucher").Preload("GormUser").First(&order).Error
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

func (m OrderModel) Update(o *Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var order *Order

	err := m.DB.WithContext(ctx).First(&order, o.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	order.Status = o.Status

	err = m.DB.Save(&order).Error
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

// ====================================================================================
// Business Functions
// ====================================================================================

func (m OrderModel) GetAPI(p Pagination, userID int64, statusType string) ([]*Order, Metadata, error) {
	var orders []*Order
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if statusType == "" {
		err := m.DB.WithContext(ctx).Preload("Voucher").Preload("GormUser").Scopes(Paginate(p)).Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error
		if err != nil {
			return nil, Metadata{}, err
		}

		err = m.DB.Table("orders").Where("user_id = ?", userID).Count(&count).Error
		if err != nil {
			return nil, Metadata{}, err
		}
	} else {
		err := m.DB.WithContext(ctx).Preload("Voucher").Scopes(Paginate(p)).Where("user_id = ?", userID).Where("status = ?", statusType).Order("created_at DESC").Find(&orders).Error
		if err != nil {
			return nil, Metadata{}, err
		}

		err = m.DB.Table("orders").Where("user_id = ?", userID).Where("status = ?", statusType).Count(&count).Error
		if err != nil {
			return nil, Metadata{}, err
		}
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return orders, metadata, nil
}

func (m OrderModel) GetWithTx(id int64, tx *gorm.DB) (*Order, error) {
	var order *Order

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Preload("OrderDetail").Where("id = ?", id).First(&order).Error
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

func (m OrderModel) InsertWithTx(order *Order, tx *gorm.DB) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.WithContext(ctx).Create(&order).Error
	if err != nil {
		return 0, err
	}

	orderID := order.ID

	return orderID, err
}

func (m OrderModel) SetTotalWithTx(id int64, subtotal int64, total int64, tx *gorm.DB) error {
	var order *Order

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).First(&order, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	order.Subtotal = subtotal
	order.Total = total

	err = tx.WithContext(ctx).Save(&order).Error
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

func (m OrderModel) SetTotalWithVoucherAndTx(id int64, subtotal int64, voucherID int64, total int64, tx *gorm.DB) error {
	var order *Order

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).First(&order, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	order.Subtotal = subtotal
	order.VoucherID = sql.NullInt64{Int64: voucherID, Valid: true}
	order.Total = total

	err = tx.WithContext(ctx).Save(&order).Error
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

func (m OrderModel) UpdateStatus(id int64, status string) error {
	var order *Order

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).First(&order, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	order.Status = status

	err = m.DB.WithContext(ctx).Save(&order).Error
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
