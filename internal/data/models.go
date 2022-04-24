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
)

type Models struct {
	Movies      MovieModel
	Permissions PermissionModel
	Tokens      TokenModel
	Users       UserModel
	Banners     BannerModel
}

type Gorm struct {
	Banners           GormBannerModel
	Brands            BrandModel
	Blogs             BlogModel
	BlogCategories    BlogCategoryModel
	ProductCategories ProductCategoryModel
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
		Banners:           GormBannerModel{DB: db},
		Brands:            BrandModel{DB: db},
		Blogs:             BlogModel{DB: db},
		BlogCategories:    BlogCategoryModel{DB: db},
		ProductCategories: ProductCategoryModel{DB: db},
	}
}

// func NewMockModels() Models {
// 	return Models{
// 		Movies: MockMovieModel{},
// 	}
// }
