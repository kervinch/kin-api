package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/s3"
	"github.com/kervinch/internal/validator"
)

// ====================================================================================
// Backoffice Handlers
// ====================================================================================

func (app *application) createProductImageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	var imageURL string
	ch := make(chan string)

	file, handler, err := r.FormFile("product_image")
	if err == nil {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.PRODUCT, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch <- url
			close(ch)
		})
		imageURL = <-ch
		defer file.Close()
	} else {
		imageURL = ""
	}

	productDetailID, err := strconv.Atoi(r.FormValue("product_detail_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	productImage := &data.ProductImage{
		ProductDetailID: int64(productDetailID),
		ImageURL:        imageURL,
		IsMain:          r.FormValue("is_main") == "true",
	}

	v := validator.New()

	if data.ValidateProductImage(v, productImage); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.ProductImages.Insert(productImage)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), productImage, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProductImageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	productImage, err := app.gorm.ProductImages.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	r.ParseMultipartForm(data.DefaultMaxMemory)

	var imageURL string
	ch := make(chan string)

	file, handler, err := r.FormFile("product_image")
	if err == nil && handler.Size > 0 {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.PRODUCT, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch <- url
			close(ch)
		})
		imageURL = <-ch
		defer file.Close()
	} else {
		imageURL = productImage.ImageURL
	}

	productDetailID, err := strconv.Atoi(r.FormValue("product_detail_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	productImage.ProductDetailID = int64(productDetailID)
	productImage.ImageURL = imageURL
	productImage.IsMain = r.FormValue("is_main") == "true"

	v := validator.New()

	if data.ValidateProductImage(v, productImage); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.ProductImages.Update(productImage)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), productImage, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProductImageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.ProductImages.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "product image successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
