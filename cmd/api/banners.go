package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/s3"
	"github.com/kervinch/internal/validator"
)

// ====================================================================================
// Backoffice Handlers (SQL)
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
	headers.Set("Location", fmt.Sprintf("/banners/%d", banner.ID))

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

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getBannersHandler(w http.ResponseWriter, r *http.Request) {
	entry, _ := app.cache.Get("GET_BANNERS_API")
	if entry != nil {
		var e any
		err := json.Unmarshal(entry, &e)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), e, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	banners, err := app.models.Banners.GetAPI()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	b, _ := json.Marshal(banners)
	app.cache.Set("GET_BANNERS_API", b)

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), banners, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// GORM Handlers
// ====================================================================================

func (app *application) gormListBannerHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	banners, metadata, err := app.gorm.Banners.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), banners, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) gormShowBannerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	banner, err := app.gorm.Banners.Get(id)
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

func (app *application) gormCreateBannerHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	file, handler, err := r.FormFile("banner_image")
	if err != nil {
		app.fileNotFoundResponse(w, r, "banner_image")
		return
	}
	defer file.Close()

	url, err := app.s3.Upload(file, s3.BANNER, handler.Filename, handler.Header.Get("Content-Type"))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	banner := &data.Banner{
		ImageURL:    url,
		Title:       r.FormValue("title"),
		Deeplink:    r.FormValue("deeplink"),
		OutboundURL: r.FormValue("outbound_url"),
		IsActive:    r.FormValue("is_active") == "true",
	}

	v := validator.New()

	if data.ValidateBanner(v, banner); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Banners.Insert(banner)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/banners/%d", banner.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), banner, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) gormFullUpdateBannerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	banner, err := app.gorm.Banners.Get(id)
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

	file, handler, err := r.FormFile("banner_image")
	if err == nil && handler.Size > 0 {
		url, err = app.s3.Upload(file, s3.BANNER, handler.Filename, handler.Header.Get("Content-Type"))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		defer file.Close()
	} else {
		url = banner.ImageURL
	}

	banner.ImageURL = url
	banner.Title = r.FormValue("title")
	banner.Deeplink = r.FormValue("deeplink")
	banner.OutboundURL = r.FormValue("outbound_url")
	banner.IsActive = r.FormValue("title") == "true"

	v := validator.New()

	if data.ValidateBanner(v, banner); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Banners.Update(banner)
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

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), banner, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) gormDeleteBannerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.Banners.Delete(id)
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
