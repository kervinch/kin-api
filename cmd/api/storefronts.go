package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/s3"
	"github.com/kervinch/internal/validator"
)

// ====================================================================================
// Backoffice Handlers
// ====================================================================================

func (app *application) listStorefrontsHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	storefronts, metadata, err := app.gorm.Storefronts.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), storefronts, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showStorefrontHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	storefront, err := app.gorm.Storefronts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), storefront, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createStorefrontHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	file, handler, err := r.FormFile("storefront_image")
	if err != nil {
		app.fileNotFoundResponse(w, r, "storefront_image")
		return
	}
	defer file.Close()

	var imageURL string
	ch := make(chan string)

	app.background(func() {
		url, err := app.s3.Upload(file, s3.STOREFRONT, handler.Filename, handler.Header.Get("Content-Type"))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		ch <- url
		close(ch)
	})
	imageURL = <-ch

	storefront := &data.Storefront{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		ImageURL:    imageURL,
		Slug:        app.slugify(r.FormValue("name")),
		IsActive:    r.FormValue("is_active") == "true",
	}

	v := validator.New()

	if data.ValidateStorefront(v, storefront); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Storefronts.Insert(storefront)
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

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/storefronts/%d", storefront.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), storefront, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateStorefrontHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	storefront, err := app.gorm.Storefronts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var imageURL string
	ch := make(chan string)
	r.ParseMultipartForm(data.DefaultMaxMemory)

	file, handler, err := r.FormFile("storefront_image")
	if err == nil && handler.Size > 0 {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.STOREFRONT, handler.Filename, handler.Header.Get("Content-Type"))
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
		imageURL = storefront.ImageURL
	}

	storefront.ImageURL = imageURL
	storefront.Name = r.FormValue("name")
	storefront.Description = r.FormValue("description")
	storefront.Slug = app.slugify(r.FormValue("name"))
	storefront.IsActive = r.FormValue("is_active") == "true"

	v := validator.New()

	if data.ValidateStorefront(v, storefront); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Storefronts.Update(storefront)
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

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), storefront, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteStorefrontHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.Storefronts.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "storefront successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================
