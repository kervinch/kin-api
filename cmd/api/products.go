package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/s3"
	"github.com/kervinch/internal/validator"
)

// ====================================================================================
// Backoffice Handlers
// ====================================================================================

func (app *application) listProductsHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	products, metadata, err := app.gorm.Products.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	product, err := app.gorm.Products.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), product, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createProductHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	// =============
	// Product Logic
	// =============

	productCategoryID, err := strconv.Atoi(r.FormValue("product_category_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	brandID, err := strconv.Atoi(r.FormValue("brand_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	weight, err := strconv.Atoi(r.FormValue("weight"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	minimumOrder, err := strconv.Atoi(r.FormValue("minimum_order"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	storefrontID, err := app.split(r.FormValue("storefront_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	preorderDays, err := strconv.Atoi(r.FormValue("preorder_days"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	product := &data.Product{
		ProductCategoryID: int64(productCategoryID),
		BrandID:           int64(brandID),
		Name:              r.FormValue("name"),
		Description:       r.FormValue("description"),
		Weight:            weight,
		MinimumOrder:      minimumOrder,
		PreorderDays:      preorderDays,
		Condition:         r.FormValue("condition"),
		Slug:              app.slugify(r.FormValue("name")),
		InsuranceRequired: r.FormValue("insurance_required") == "true",
		IsActive:          r.FormValue("is_active") == "true",
	}

	v := validator.New()

	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	db := app.gorm.Transaction.DB
	tx := db.Begin()

	productID, err := app.gorm.Products.InsertWithTx(product, tx)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateSlug):
			v.AddError("slug", "an entry with this slug already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.gorm.ProductStorefrontSubscriptions.InsertWithTx(productID, storefrontID, tx)
	if err != nil {
		tx.Rollback()
		app.serverErrorResponse(w, r, err)
		return
	}

	// ====================
	// Product Detail Logic
	// ====================

	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil {
		tx.Rollback()
		app.badRequestResponse(w, r, err)
		return
	}

	stock, err := strconv.Atoi(r.FormValue("stock"))
	if err != nil {
		tx.Rollback()
		app.badRequestResponse(w, r, err)
		return
	}

	productDetail := &data.ProductDetail{
		ProductID: productID,
		Color:     r.FormValue("color"),
		Size:      r.FormValue("size"),
		Price:     int64(price),
		SKU:       r.FormValue("sku"),
		Stock:     stock,
		IsActive:  r.FormValue("is_active") == "true",
	}

	v = validator.New()

	if data.ValidateProductDetail(v, productDetail); !v.Valid() {
		tx.Rollback()
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	productDetailID, err := app.gorm.ProductDetails.InsertWithTx(productDetail, tx)
	if err != nil {
		tx.Rollback()
		app.serverErrorResponse(w, r, err)
		return
	}

	tx.Commit()

	// ====================
	// Product Images Logic
	// ====================

	app.background(func() {
		tx := app.gorm.Transaction.DB

		for _, handler := range r.MultipartForm.File["product_images"] {
			file, err := handler.Open()
			if err != nil {
				tx.Rollback()
				app.fileNotFoundResponse(w, r, "product_images")
				return
			}

			defer file.Close()

			url, err := app.s3.Upload(file, s3.PRODUCT, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				tx.Rollback()
				switch {
				case errors.Is(err, data.ErrImageFormat):
					app.badRequestResponse(w, r, err)
				default:
					app.serverErrorResponse(w, r, err)
				}
				return
			}

			productImages := &data.ProductImage{
				ProductDetailID: productDetailID,
				ImageURL:        url,
				IsMain:          r.FormValue("is_main") == "true",
			}

			v = validator.New()

			if data.ValidateProductImage(v, productImages); !v.Valid() {
				tx.Rollback()
				app.failedValidationResponse(w, r, v.Errors)
				return
			}

			err = app.gorm.ProductImages.InsertWithTx(productImages, tx)
			if err != nil {
				tx.Rollback()
				app.serverErrorResponse(w, r, err)
				return
			}
		}

		tx.Commit()
	})

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/products/%d", product.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), product, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	product, err := app.gorm.Products.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	productDetail, err := app.gorm.ProductDetails.Get(product.ProductDetail[0].ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// =============
	// Product Logic
	// =============

	productCategoryID, err := strconv.Atoi(r.FormValue("product_category_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	brandID, err := strconv.Atoi(r.FormValue("brand_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	weight, err := strconv.Atoi(r.FormValue("weight"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	minimumOrder, err := strconv.Atoi(r.FormValue("minimum_order"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	storefrontID, err := app.split(r.FormValue("storefront_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	preorderDays, err := strconv.Atoi(r.FormValue("preorder_days"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	product.ProductCategoryID = int64(productCategoryID)
	product.BrandID = int64(brandID)
	product.Name = r.FormValue("name")
	product.Description = r.FormValue("description")
	product.Weight = weight
	product.MinimumOrder = minimumOrder
	product.PreorderDays = preorderDays
	product.Condition = r.FormValue("condition")
	product.Slug = app.slugify(r.FormValue("name"))
	product.InsuranceRequired = r.FormValue("insurance_required") == "true"
	product.IsActive = r.FormValue("is_active") == "true"

	v := validator.New()

	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	db := app.gorm.Transaction.DB
	tx := db.Begin()

	err = app.gorm.Products.UpdateWithTx(product, tx)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, data.ErrDuplicateSlug):
			v.AddError("slug", "an entry with this slug already exists")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.gorm.ProductStorefrontSubscriptions.UpdateWithTx(product.ID, storefrontID, tx)
	if err != nil {
		tx.Rollback()
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
	}

	// ====================
	// Product Detail Logic
	// ====================

	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil {
		tx.Rollback()
		app.badRequestResponse(w, r, err)
		return
	}

	stock, err := strconv.Atoi(r.FormValue("stock"))
	if err != nil {
		tx.Rollback()
		app.badRequestResponse(w, r, err)
		return
	}

	productDetail.Color = r.FormValue("color")
	productDetail.Size = r.FormValue("size")
	productDetail.Price = int64(price)
	productDetail.SKU = r.FormValue("sku")
	productDetail.Stock = stock
	productDetail.IsActive = r.FormValue("is_active") == "true"

	v = validator.New()

	if data.ValidateProductDetail(v, productDetail); !v.Valid() {
		tx.Rollback()
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.ProductDetails.UpdateWithTx(productDetail, tx)
	if err != nil {
		tx.Rollback()
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
	}

	tx.Commit()

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "product successfully updated", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	product, err := app.gorm.Products.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	db := app.gorm.Transaction.DB
	tx := db.Begin()

	err = app.gorm.Products.DeleteWithTx(product.ID, tx)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.gorm.ProductStorefrontSubscriptions.DeleteWithTx(product.ID, tx)
	if err != nil {
		tx.Rollback()
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	for _, pd := range product.ProductDetail {
		for _, pi := range pd.ProductImage {
			err = app.gorm.ProductImages.DeleteWithTx(pi.ID, tx)
			if err != nil {
				tx.Rollback()
				switch {
				case errors.Is(err, data.ErrRecordNotFound):
					app.notFoundResponse(w, r)
				default:
					app.serverErrorResponse(w, r, err)
				}
				return
			}
		}

		err = app.gorm.ProductDetails.DeleteWithTx(pd.ID, tx)
		if err != nil {
			tx.Rollback()
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
	}

	tx.Commit()

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "product successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getProductsHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	products, metadata, err := app.gorm.Products.GetAPI(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getProductsLatestHandler(w http.ResponseWriter, r *http.Request) {
	products, err := app.gorm.Products.GetLatest()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getProductsRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	products, err := app.gorm.Products.GetRecommendation()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getProductBySlugHandler(w http.ResponseWriter, r *http.Request) {
	slug := app.readSlugParam(r)

	products, err := app.gorm.Products.GetBySlug(slug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getProductsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	slug := app.readSlugParam(r)

	products, err := app.gorm.ProductCategories.GetBySlugWithProducts(slug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getProductsByBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	products, err := app.gorm.Brands.GetByIdWithProducts(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getProductsByStorefrontHandler(w http.ResponseWriter, r *http.Request) {
	slug := app.readSlugParam(r)

	products, err := app.gorm.Storefronts.GetBySlugWithProducts(slug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), products, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}