package main

import (
	"errors"
	"net/http"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/validator"
)

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	favorites, metadata, err := app.gorm.Favorites.GetAll(pagination, user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), favorites, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		ProductDetailID int64 `json:"product_detail_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	favorite := &data.Favorite{
		UserID:          user.ID,
		ProductDetailID: input.ProductDetailID,
	}

	v := validator.New()

	if data.ValidateFavorite(v, favorite); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Favorites.Insert(favorite)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateKeyValue):
			app.violateUniqueConstraint(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), favorite, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.Favorites.Delete(id, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "favorite successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
