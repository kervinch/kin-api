package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Voucher struct {
	ID                int64         `json:"id"`
	Type              string        `json:"type"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	TermsAndCondition string        `json:"terms_and_condition"`
	ImageURL          string        `json:"image_url"`
	Slug              string        `json:"slug"`
	BrandID           sql.NullInt64 `json:"brand_id"`
	Brand             Brand         `json:"brand"`
	LogisticID        sql.NullInt64 `json:"logistic_id"`
	Logistic          Logistic      `json:"logistic"`
	Code              string        `json:"code"`
	IsPercent         bool          `json:"is_percent"`
	Value             int           `json:"value"`
	Stock             int           `json:"stock"`
	IsActive          bool          `json:"is_active"`
	EffectiveAt       time.Time     `json:"effective_at"`
	ExpiredAt         time.Time     `json:"expired_at"`
	CreatedAt         time.Time     `json:"-"`
	UpdatedAt         time.Time     `json:"-"`
	CreatedBy         int64         `json:"created_by"`
	UpdatedBy         int64         `json:"updated_by"`
}

func ValidateVoucher(v *validator.Validator, voucher *Voucher) {
	v.Check(voucher.Type != "", "type", "must be provided")
	v.Check(voucher.Code != "", "code", "must be provided")
	v.Check(voucher.Value != 0, "value", "must be not be zero")
}

type VoucherModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m VoucherModel) GetAll(p Pagination) ([]*Voucher, Metadata, error) {
	var voucher []*Voucher
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("Brand").Preload("Logistic").Order("created_at DESC").Find(&voucher).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("vouchers").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return voucher, metadata, nil
}

func (m VoucherModel) Get(id int64) (*Voucher, error) {
	var voucher *Voucher

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).Preload("Brand").Preload("Logistic").First(&voucher).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return voucher, nil
}

func (m VoucherModel) Insert(voucher *Voucher) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&voucher).Error
	if err != nil {
		return err
	}

	return err
}

func (m VoucherModel) Update(v *Voucher) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var voucher *Voucher

	err := m.DB.WithContext(ctx).First(&voucher, v.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	voucher.Type = v.Type
	voucher.Name = v.Name
	voucher.Description = v.Description
	voucher.TermsAndCondition = v.TermsAndCondition
	voucher.ImageURL = v.ImageURL
	voucher.BrandID = v.BrandID
	voucher.LogisticID = v.LogisticID
	voucher.Code = v.Code
	voucher.IsPercent = v.IsPercent
	voucher.Value = v.Value
	voucher.Stock = v.Stock
	voucher.IsActive = v.IsActive
	voucher.Slug = v.Slug
	voucher.EffectiveAt = v.EffectiveAt
	voucher.ExpiredAt = v.ExpiredAt
	voucher.CreatedBy = v.CreatedBy
	voucher.UpdatedBy = v.UpdatedBy

	err = m.DB.Save(&voucher).Error
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

func (m VoucherModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Where("id = ?", id).Delete(&Voucher{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m VoucherModel) GetByBrands(p Pagination, brandID []int) ([]*Voucher, Metadata, error) {
	var voucher []*Voucher
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Preload("Brand").Preload("Logistic").Where("brand_id = ?", brandID).Order("brand_id ASC").Find(&voucher).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("vouchers").Where("brand_id = ?", brandID).Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return voucher, metadata, nil
}

func (m VoucherModel) Consume(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var voucher *Voucher

	err := m.DB.WithContext(ctx).First(&voucher, id).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	if voucher.Stock < 1 {
		return ErrOutOfStock
	}

	voucher.Stock = voucher.Stock - 1

	err = m.DB.WithContext(ctx).Save(&voucher).Error
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

func (m VoucherModel) GetByID(id int64) (*Voucher, error) {
	var voucher *Voucher

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).Where("is_active = ?", true).First(&voucher).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return voucher, nil
}
