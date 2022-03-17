package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/validator"
)

// ====================================================================================
// Backoffice Handlers
// ====================================================================================

func (app *application) listBannersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readStrings(qs, "title", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readStrings(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "-id", "-title"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	banners, metadata, err := app.models.Banners.GetAll(input.Title, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), envelope{"result": banners, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBannerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	banner, err := app.models.Banners.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), banner, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createBannerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ImageURL    string `json:"image_url"`
		Title       string `json:"title"`
		Deeplink    string `json:"deeplink"`
		OutboundURL string `json:"outbound_url"`
		IsActive    bool   `json:"is_active"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	banner := &data.Banner{
		ImageURL:    input.ImageURL,
		Title:       input.Title,
		Deeplink:    input.Deeplink,
		OutboundURL: input.OutboundURL,
		IsActive:    input.IsActive,
	}

	v := validator.New()

	if data.ValidateBanner(v, banner); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Banners.Insert(banner)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/banners/%d", banner.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), banner, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) fullUpdateBannerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	banner, err := app.models.Banners.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		ImageURL    string `json:"image_url"`
		Title       string `json:"title"`
		Deeplink    string `json:"deeplink"`
		OutboundURL string `json:"outbound_url"`
		IsActive    bool   `json:"is_active"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	banner.ImageURL = input.ImageURL
	banner.Title = input.Title
	banner.Deeplink = input.Deeplink
	banner.OutboundURL = input.OutboundURL
	banner.IsActive = input.IsActive

	v := validator.New()

	if data.ValidateBanner(v, banner); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Banners.Update(banner)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), banner, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBannerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Banners.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "banner successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) gormListBannerHandler(w http.ResponseWriter, r *http.Request) {
	banners, err := app.gorm.Banners.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), banners, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getBannersHandler(w http.ResponseWriter, r *http.Request) {
	banners, err := app.models.Banners.GetAPI()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), banners, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
