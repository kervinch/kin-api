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

func (app *application) listInboxHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	favorites, metadata, err := app.gorm.Inbox.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), favorites, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showInboxHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	inbox, err := app.gorm.Inbox.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), inbox, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createInboxHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	var imageURL string
	ch := make(chan string)

	file, handler, err := r.FormFile("inbox_image")
	if err == nil {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.INBOX, handler.Filename, handler.Header.Get("Content-Type"))
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

	inbox := &data.Inbox{
		Title:    r.FormValue("title"),
		Content:  r.FormValue("content"),
		Deeplink: r.FormValue("deeplink"),
		ImageURL: imageURL,
		Slug:     app.slugify(r.FormValue("title")),
	}

	v := validator.New()

	if data.ValidateInbox(v, inbox); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Inbox.Insert(inbox)
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
	headers.Set("Location", fmt.Sprintf("/inbox/%d", inbox.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), inbox, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateInboxHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	inbox, err := app.gorm.Inbox.Get(id)
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

	file, handler, err := r.FormFile("inbox_image")
	if err == nil && handler.Size > 0 {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.INBOX, handler.Filename, handler.Header.Get("Content-Type"))
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
		imageURL = inbox.ImageURL
	}

	inbox.Title = r.FormValue("title")
	inbox.Content = r.FormValue("content")
	inbox.ImageURL = imageURL
	inbox.Deeplink = r.FormValue("deeplink")
	inbox.Slug = app.slugify(r.FormValue("title"))

	v := validator.New()

	if data.ValidateInbox(v, inbox); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Inbox.Update(inbox)
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

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), inbox, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteInboxHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.Inbox.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.gorm.InboxUsers.Delete(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "inbox successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getInboxHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	inbox, metadata, err := app.gorm.Inbox.GetAPI(pagination, user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), inbox, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getInboxBySlugHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	slug := app.readSlugParam(r)

	inbox, err := app.gorm.Inbox.GetBySlug(slug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	inboxUser := &data.InboxUser{
		InboxID: inbox.ID,
		UserID:  user.ID,
		IsRead:  true,
	}

	app.gorm.InboxUsers.Insert(inboxUser)

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), inbox, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
