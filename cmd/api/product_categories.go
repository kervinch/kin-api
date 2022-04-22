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

func (app *application) listProductCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	productCategories, metadata, err := app.gorm.ProductCategories.GetAll(pagination)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), productCategories, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProductCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	productCategory, err := app.gorm.ProductCategories.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), productCategory, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createProductCategoryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)
	file, handler, err := r.FormFile("image")
	ch := make(chan string)
	var url string
	if err == nil {
		app.background(func() {
			url, err = app.s3.Upload(file, s3.PRODUCT_CATEGORY, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch <- url
			close(ch)
		})
		url = <-ch
		defer file.Close()
	} else {
		url = ""
	}

	orderNumber, err := strconv.Atoi(r.FormValue("order_number"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	productCategory := &data.ProductCategory{
		Image:       url,
		Name:        r.FormValue("name"),
		Slug:        app.slugify(r.FormValue("name")),
		IsActive:    r.FormValue("is_active") == "true",
		OrderNumber: orderNumber,
	}

	v := validator.New()

	if data.ValidateProductCategory(v, productCategory); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.ProductCategories.Insert(productCategory)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/product_categories/%d", productCategory.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), productCategory, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProductCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	productCategory, err := app.gorm.ProductCategories.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var imgURL string
	ch := make(chan string)
	r.ParseMultipartForm(data.DefaultMaxMemory)
	file, handler, _ := r.FormFile("image")
	if file != nil {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.PRODUCT_CATEGORY, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch <- url
			close(ch)
		})
		imgURL = <-ch
		defer file.Close()
	} else {
		imgURL = productCategory.Image
	}

	orderNumber, err := strconv.Atoi(r.FormValue("order_number"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	productCategory.Image = imgURL
	productCategory.Name = r.FormValue("name")
	productCategory.Slug = app.slugify(r.FormValue("name"))
	productCategory.IsActive = r.FormValue("is_active") == "true"
	productCategory.OrderNumber = orderNumber

	v := validator.New()

	if data.ValidateProductCategory(v, productCategory); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.ProductCategories.Update(productCategory)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), productCategory, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProductCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.ProductCategories.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "product category successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getProductCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	productCategories, err := app.gorm.ProductCategories.GetAPI()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), productCategories, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getProductCategoriesBySlugHandler(w http.ResponseWriter, r *http.Request) {
	slug := app.readSlugParam(r)

	productCategories, err := app.gorm.ProductCategories.GetBySlug(slug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), productCategories, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
