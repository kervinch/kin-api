package data

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateSlug  = errors.New("duplicate slug")
	ErrInvalidEnum    = errors.New("invalid enum value")
	ErrBadRequest     = errors.New("bad request")
	ErrImageFormat    = errors.New("unknown image format")
)

type TransactionModel struct {
	DB *gorm.DB
}

type Models struct {
	Movies      MovieModel
	Permissions PermissionModel
	Tokens      TokenModel
	Users       UserModel
	Banners     BannerModel
}

type Gorm struct {
	Transaction                    TransactionModel
	Banners                        GormBannerModel
	Brands                         BrandModel
	Blogs                          BlogModel
	BlogCategories                 BlogCategoryModel
	Carts                          CartModel
	Favorites                      FavoriteModel
	Products                       ProductModel
	ProductCategories              ProductCategoryModel
	ProductDetails                 ProductDetailModel
	ProductImages                  ProductImageModel
	ProductVideos                  ProductVideoModel
	ProductStorefrontSubscriptions ProductStorefrontSubscriptionModel
	Storefronts                    StorefrontModel
	UserAddresses                  UserAddressModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
		Banners:     BannerModel{DB: db},
	}
}

func GormModels(db *gorm.DB) Gorm {
	return Gorm{
		Transaction:                    TransactionModel{DB: db},
		Banners:                        GormBannerModel{DB: db},
		Brands:                         BrandModel{DB: db},
		Blogs:                          BlogModel{DB: db},
		BlogCategories:                 BlogCategoryModel{DB: db},
		Carts:                          CartModel{DB: db},
		Favorites:                      FavoriteModel{DB: db},
		Products:                       ProductModel{DB: db},
		ProductCategories:              ProductCategoryModel{DB: db},
		ProductDetails:                 ProductDetailModel{DB: db},
		ProductImages:                  ProductImageModel{DB: db},
		ProductVideos:                  ProductVideoModel{DB: db},
		ProductStorefrontSubscriptions: ProductStorefrontSubscriptionModel{DB: db},
		Storefronts:                    StorefrontModel{DB: db},
		UserAddresses:                  UserAddressModel{DB: db},
	}
}
