package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type Logistic struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func ValidateLogistic(v *validator.Validator, voucher *Voucher) {
	v.Check(voucher.Type != "", "name", "must be provided")
	v.Check(voucher.Code != "", "cotypede", "must be provided")
}

type LogisticModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m LogisticModel) GetAll(p Pagination) ([]*Logistic, Metadata, error) {
	var logsitic []*Logistic
	var count int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Scopes(Paginate(p)).Order("created_at DESC").Find(&logsitic).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	err = m.DB.Table("logistics").Count(&count).Error
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(count), p.Page, p.PageSize)

	return logsitic, metadata, nil
}

func (m LogisticModel) Get(id int64) (*Logistic, error) {
	var logistic *Logistic

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ?", id).First(&logistic).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return logistic, nil
}

func (m LogisticModel) Insert(logistic *Logistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&logistic).Error
	if err != nil {
		return err
	}

	return err
}

func (m LogisticModel) Update(l *Logistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var logistic *Logistic

	err := m.DB.WithContext(ctx).First(&logistic, l.ID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	logistic.Name = l.Name
	logistic.Type = l.Type
	logistic.IsActive = l.IsActive

	err = m.DB.Save(&logistic).Error
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

func (m LogisticModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Where("id = ?", id).Delete(&Logistic{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

// ====================================================================================
// Business Functions
// ====================================================================================
