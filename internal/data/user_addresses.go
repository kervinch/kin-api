package data

import (
	"context"
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type UserAddress struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Receiver    string    `json:"receiver"`
	PhoneNumber string    `json:"phone_number"`
	City        string    `json:"city"`
	PostalCode  string    `json:"postal_code"`
	Address     string    `json:"address"`
	IsMain      bool      `json:"is_main"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func ValidateUserAddress(v *validator.Validator, userAddress *UserAddress) {
	v.Check(userAddress.Name != "", "name", "must be provided")
	v.Check(userAddress.Receiver != "", "receiver name", "must be provided")
	v.Check(userAddress.PhoneNumber != "", "phone number", "must be provided")
	v.Check(userAddress.City != "", "city", "must be provided")
	v.Check(userAddress.PostalCode != "", "postal code", "must be provided")
	v.Check(userAddress.Address != "", "address", "must be provided")
}

type UserAddressModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Business Functions
// ====================================================================================

func (m UserAddressModel) GetAPI(user *User) ([]*UserAddress, error) {
	var userAddresses []*UserAddress

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("user_id = ?", user.ID).Order("id").Find(&userAddresses).Error
	if err != nil {
		return nil, err
	}

	return userAddresses, nil
}

func (m UserAddressModel) Get(id int64, user *User) (*UserAddress, error) {
	var userAddress *UserAddress

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, user.ID).First(&userAddress).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return userAddress, nil
}

func (m UserAddressModel) Insert(userAddress *UserAddress) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.WithContext(ctx).Create(&userAddress).Error

	return err
}

func (m UserAddressModel) Update(ua *UserAddress, user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userAddress *UserAddress

	err := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", ua.ID, user.ID).First(&userAddress).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	userAddress.Name = ua.Name
	userAddress.Receiver = ua.Receiver
	userAddress.PhoneNumber = ua.PhoneNumber
	userAddress.City = ua.City
	userAddress.PostalCode = ua.PostalCode
	userAddress.Address = ua.Address

	err = m.DB.WithContext(ctx).Save(&userAddress).Error
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

func (m UserAddressModel) UpdateMain(ua *UserAddress, user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userAddress *UserAddress

	err := m.DB.WithContext(ctx).Table("user_addresses").Where("user_id = ?", user.ID).Update("is_main", false).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	err = m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", ua.ID, user.ID).First(&userAddress).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	userAddress.IsMain = ua.IsMain

	err = m.DB.WithContext(ctx).Save(&userAddress).Error
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

func (m UserAddressModel) Delete(id int64, userID int64) error {
	if id < 1 || userID < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&UserAddress{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}
