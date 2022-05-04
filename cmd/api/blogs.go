package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/s3"
	"github.com/kervinch/internal/validator"
)

// ====================================================================================
// Backoffice Handlers
// ====================================================================================

func (app *application) listBlogsHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	blogs, metadata, err := app.gorm.Blogs.GetAll(pagination)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), blogs, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBlogHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	blog, err := app.gorm.Blogs.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blog, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createBlogHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	file, handler, err := r.FormFile("thumbnail")
	if err != nil {
		app.fileNotFoundResponse(w, r, "thumbnail")
		return
	}
	defer file.Close()

	url, err := app.s3.Upload(file, s3.BLOG, handler.Filename, handler.Header.Get("Content-Type"))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	blogCategoryId, err := strconv.Atoi(r.FormValue("blog_category_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	blog := &data.Blog{
		BlogCategoryID: blogCategoryId,
		Thumbnail:      url,
		Title:          r.FormValue("title"),
		Description:    r.FormValue("description"),
		Content:        r.FormValue("content"),
		Slug:           app.slugify(r.FormValue("title")),
		Type:           "buyer",
		PublishedAt:    time.Now(),
		Feature:        r.FormValue("feature") == "true",
		Status:         r.FormValue("status"),
		Tags:           r.FormValue("tags"),
		CreatedBy:      1,
		CreatedByText:  "Admin",
	}

	v := validator.New()

	if data.ValidateBlog(v, blog); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Blogs.Insert(blog)
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
	headers.Set("Location", fmt.Sprintf("/blogs/%d", blog.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), blog, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBlogHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	blog, err := app.gorm.Blogs.Get(id)
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

	file, handler, err := r.FormFile("thumbnail")
	if err == nil && handler.Size > 0 {
		url, err = app.s3.Upload(file, s3.BLOG, handler.Filename, handler.Header.Get("Content-Type"))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		defer file.Close()
	} else {
		url = blog.Thumbnail
	}

	blogCategoryId, err := strconv.Atoi(r.FormValue("blog_category_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	blog.Thumbnail = url
	blog.BlogCategoryID = blogCategoryId
	blog.Title = r.FormValue("title")
	blog.Description = r.FormValue("description")
	blog.Content = r.FormValue("content")
	blog.Slug = app.slugify(r.FormValue("title"))
	blog.Feature = r.FormValue("feature") == "true"
	blog.Status = r.FormValue("status")
	blog.Tags = r.FormValue("tags")

	v := validator.New()

	if data.ValidateBlog(v, blog); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Blogs.Update(blog)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blog, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBlogHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.Blogs.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "blog successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getBlogsHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	blogs, metadata, err := app.gorm.Blogs.GetAPI(pagination)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), blogs, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getBlogBySlugHandler(w http.ResponseWriter, r *http.Request) {
	slug := app.readSlugParam(r)

	blog, err := app.gorm.Blogs.GetBySlug(slug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blog, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getBlogsRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	blogs, err := app.gorm.Blogs.GetRecommendations()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), blogs, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
