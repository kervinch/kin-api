package data

import (
	"errors"
	"time"

	"github.com/kervinch/internal/validator"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type ProductStorefrontSubscription struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	StorefrontID int64     `json:"storefront_id"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func ValidateProductStorefrontSubscription(v *validator.Validator, pss *ProductStorefrontSubscription) {
	v.Check(pss.ProductID != 0, "product_id", "must be provided")
	v.Check(pss.ProductID > 0, "product_id", "must be a positive integer")
	v.Check(pss.StorefrontID != 0, "storefront_id", "must be provided")
	v.Check(pss.StorefrontID > 0, "storefront_id", "must be a positive integer")
}

type ProductStorefrontSubscriptionModel struct {
	DB *gorm.DB
}

// ====================================================================================
// Backoffice Functions
// ====================================================================================

func (m ProductStorefrontSubscriptionModel) Insert(productID int64, storefrontID []int64) error {
	var err error

	for i := range storefrontID {
		pss := ProductStorefrontSubscription{
			ProductID:    productID,
			StorefrontID: storefrontID[i],
		}

		err = m.DB.Create(&pss).Error
		if err != nil {
			break
		}
	}

	return err
}

func (m ProductStorefrontSubscriptionModel) Update(productID int64, storefrontID []int64) error {
	var err error
	var pss []*ProductStorefrontSubscription

	err = m.DB.Where("product_id = ?", productID).Find(&pss).Error
	if err != nil {
		return err
	}

	for _, v := range pss {
		if !slices.Contains(storefrontID, v.StorefrontID) {
			err = m.DB.Where("product_id = ? AND storefront_id = ?", productID, v.StorefrontID).Delete(&ProductStorefrontSubscription{}).Error
			if err != nil {
				return err
			}
		}
	}

	for i := range storefrontID {
		m.DB.Where("product_id = ? AND storefront_id = ?", productID, storefrontID[i]).First(&pss)

		if pss == nil {
			subscription := ProductStorefrontSubscription{
				ProductID:    productID,
				StorefrontID: storefrontID[i],
			}

			err = m.DB.Create(&subscription).Error
			if err != nil {
				break
			}
		}
	}

	return err
}

func (m ProductStorefrontSubscriptionModel) Delete(productID int64) error {
	if productID < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Where("product_id = ?", productID).Delete(&ProductStorefrontSubscriptionModel{}).Error
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
