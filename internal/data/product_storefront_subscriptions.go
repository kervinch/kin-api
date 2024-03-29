package data

import (
	"context"
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := range storefrontID {
		pss := ProductStorefrontSubscription{
			ProductID:    productID,
			StorefrontID: storefrontID[i],
		}

		err = m.DB.WithContext(ctx).Create(&pss).Error
		if err != nil {
			// add duplicate entry error for unique column
			break
		}
	}

	return err
}

func (m ProductStorefrontSubscriptionModel) InsertWithTx(productID int64, storefrontID []int64, tx *gorm.DB) error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := range storefrontID {
		pss := ProductStorefrontSubscription{
			ProductID:    productID,
			StorefrontID: storefrontID[i],
		}

		err = tx.WithContext(ctx).Create(&pss).Error
		if err != nil {
			switch {
			case err.Error() == `pq: duplicate key value violates unique constraint "idx_product_storefront"`:
				return ErrDuplicateKeyValue
			default:
				return err
			}
		}
	}

	return err
}

func (m ProductStorefrontSubscriptionModel) Update(productID int64, storefrontID []int64) error {
	var err error
	var pss []*ProductStorefrontSubscription

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = m.DB.WithContext(ctx).Where("product_id = ?", productID).Find(&pss).Error
	if err != nil {
		return err
	}

	for _, v := range pss {
		if !slices.Contains(storefrontID, v.StorefrontID) {
			err = m.DB.WithContext(ctx).Where("product_id = ? AND storefront_id = ?", productID, v.StorefrontID).Delete(&ProductStorefrontSubscription{}).Error
			if err != nil {
				return err
			}
		}
	}

	for i := range storefrontID {
		m.DB.WithContext(ctx).Where("product_id = ? AND storefront_id = ?", productID, storefrontID[i]).First(&pss)

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

func (m ProductStorefrontSubscriptionModel) UpdateWithTx(productID int64, storefrontID []int64, tx *gorm.DB) error {
	var err error
	var pss []*ProductStorefrontSubscription

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = m.DB.WithContext(ctx).Where("product_id = ?", productID).Find(&pss).Error
	if err != nil {
		return gorm.ErrRecordNotFound
	}

	for _, v := range pss {
		if !slices.Contains(storefrontID, v.StorefrontID) {
			ra := tx.WithContext(ctx).Where("product_id = ? AND storefront_id = ?", productID, v.StorefrontID).Delete(&ProductStorefrontSubscription{}).RowsAffected
			if ra < 1 {
				return gorm.ErrRecordNotFound
			}
		}
	}

	for i := range storefrontID {
		m.DB.WithContext(ctx).Where("product_id = ? AND storefront_id = ?", productID, storefrontID[i]).First(&pss)

		if pss == nil {
			subscription := ProductStorefrontSubscription{
				ProductID:    productID,
				StorefrontID: storefrontID[i],
			}

			err = tx.WithContext(ctx).Create(&subscription).Error
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := m.DB.WithContext(ctx).Where("product_id = ?", productID).Delete(&ProductStorefrontSubscriptionModel{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}

func (m ProductStorefrontSubscriptionModel) DeleteWithTx(productID int64, tx *gorm.DB) error {
	if productID < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ra := tx.WithContext(ctx).Where("product_id = ?", productID).Delete(&ProductStorefrontSubscription{}).RowsAffected
	if ra < 1 {
		return ErrRecordNotFound
	}

	return nil
}
