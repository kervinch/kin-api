package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

//  gorm:"foreignKey:ID;references:users"
//  gorm:"foreignKey:GormUserID;references:GormUser"

type UserVoucher struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	GormUser  GormUser  `json:"user" gorm:"foreignKey:UserID"`
	VoucherID int64     `json:"voucher_id"`
	Voucher   Voucher   `json:"voucher"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func ValidateUserVoucher(v *validator.Validator, userVoucher *UserVoucher) {
	v.Check(userVoucher.UserID != 0, "user_id", "must be not be zero")
	v.Check(userVoucher.VoucherID != 0, "voucher_id", "must be provided")
	v.Check(userVoucher.Quantity != 0, "quantity", "must be provided")
}

type UserVoucherModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m UserVoucherModel) GetAll(p Pagination) ([]*UserVoucher, Metadata, error) {
	var userVoucher []*UserVoucher
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("GormUser").Preload("Voucher").Order("created_at DESC").Find(&userVoucher).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("user_vouchers").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return userVoucher, metadata, nil
}

func (m UserVoucherModel) Get(id int64) (*UserVoucher, error) {
	var userVoucher *UserVoucher

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).Preload("GormUser").Preload("Voucher").First(&userVoucher).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return userVoucher, nil
}

func (m UserVoucherModel) GetAllByVoucherID(id int64) ([]*UserVoucher, error) {
	var userVoucher []*UserVoucher

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("voucher_id = ?", id).Preload("GormUser").Preload("Voucher").Find(&userVoucher).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return userVoucher, nil
}

func (m UserVoucherModel) Insert(userVoucher *UserVoucher) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&userVoucher).Error
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "idx_user_vouchers"`:
			return ErrDuplicateKeyValue
		default:
			return err
		}
	}

	return err
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m UserVoucherModel) Consume(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userVoucher *UserVoucher

	err := m.DB.WithContext(ctx).First(&userVoucher, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	if userVoucher.Quantity < 1 {
		return ErrOutOfQuantity
	}

	userVoucher.Quantity = userVoucher.Quantity - 1

	err = m.DB.WithContext(ctx).Save(&userVoucher).Error
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
