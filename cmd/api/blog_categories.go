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

func (app *application) listBlogCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	blogCategories, metadata, err := app.gorm.BlogCategories.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), blogCategories, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBlogCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	blogCategory, err := app.gorm.BlogCategories.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blogCategory, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createBlogCategoryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)
	file, handler, err := r.FormFile("image")
	if err != nil {
		app.fileNotFoundResponse(w, r, "image")
		return
	}
	defer file.Close()

	url, err := app.s3.Upload(file, s3.BLOG_CATEGORY, handler.Filename, handler.Header.Get("Content-Type"))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	orderNumber, err := strconv.Atoi(r.FormValue("order_number"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	blogCategory := &data.BlogCategory{
		Image:       url,
		Name:        r.FormValue("name"),
		Slug:        app.slugify(r.FormValue("name")),
		Type:        "all",
		Status:      r.FormValue("status"),
		OrderNumber: orderNumber,
	}

	v := validator.New()

	if data.ValidateBlogCategory(v, blogCategory); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.BlogCategories.Insert(blogCategory)
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
	headers.Set("Location", fmt.Sprintf("/blog_categories/%d", blogCategory.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), blogCategory, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBlogCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	blogCategory, err := app.gorm.BlogCategories.Get(id)
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
	file, handler, err := r.FormFile("image")
	if err == nil && handler.Size > 0 {
		url, err = app.s3.Upload(file, s3.BLOG_CATEGORY, handler.Filename, handler.Header.Get("Content-Type"))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		defer file.Close()
	} else {
		url = blogCategory.Image
	}

	orderNumber, err := strconv.Atoi(r.FormValue("order_number"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	blogCategory.Image = url
	blogCategory.Name = r.FormValue("name")
	blogCategory.Slug = app.slugify(r.FormValue("name"))
	blogCategory.Status = r.FormValue("status")
	blogCategory.OrderNumber = orderNumber

	v := validator.New()

	if data.ValidateBlogCategory(v, blogCategory); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.BlogCategories.Update(blogCategory)
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

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blogCategory, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBlogCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.BlogCategories.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "blog category successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getBlogCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	blogCategories, err := app.gorm.BlogCategories.GetAPI()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blogCategories, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getBlogCategoriesBySlugHandler(w http.ResponseWriter, r *http.Request) {
	slug := app.readSlugParam(r)

	blogCategory, err := app.gorm.BlogCategories.GetBySlug(slug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blogCategory, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
