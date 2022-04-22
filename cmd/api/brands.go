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

func (app *application) listBrandsHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	brands, metadata, err := app.gorm.Brands.GetAll(pagination)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), brands, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	brand, err := app.gorm.Brands.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), brand, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createBrandHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)
	file, handler, err := r.FormFile("brand_image")
	if err != nil {
		app.fileNotFoundResponse(w, r, "brand_image")
		return
	}
	defer file.Close()

	url, err := app.s3.Upload(file, s3.BRAND, handler.Filename, handler.Header.Get("Content-Type"))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	orderNumber, err := strconv.Atoi(r.FormValue("order_number"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	brand := &data.Brand{
		ImageURL:    url,
		Name:        r.FormValue("name"),
		OrderNumber: orderNumber,
		IsActive:    r.FormValue("is_active") == "true",
	}

	v := validator.New()

	if data.ValidateBrand(v, brand); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Brands.Insert(brand)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/brands/%d", brand.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), brand, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	brand, err := app.gorm.Brands.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var url string
	r.ParseMultipartForm(data.DefaultMaxMemory)
	file, handler, _ := r.FormFile("brand_image")
	if file != nil {
		url, err = app.s3.Upload(file, s3.BRAND, handler.Filename, handler.Header.Get("Content-Type"))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	} else {
		url = brand.ImageURL
	}
	defer file.Close()

	orderNumber, err := strconv.Atoi(r.FormValue("order_number"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	brand.ImageURL = url
	brand.Name = r.FormValue("name")
	brand.OrderNumber = orderNumber
	brand.IsActive = r.FormValue("is_active") == "true"

	v := validator.New()

	if data.ValidateBrand(v, brand); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Brands.Update(brand)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), brand, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.Brands.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "brand successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getBrandsHandler(w http.ResponseWriter, r *http.Request) {
	brands, err := app.gorm.Brands.GetAPI()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), brands, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
