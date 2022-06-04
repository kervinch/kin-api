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

func (app *application) listUserVouchersHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	userVouchers, metadata, err := app.gorm.UserVouchers.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), userVouchers, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showUserVoucherHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userVoucher, err := app.gorm.UserVouchers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), userVoucher, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createUserVoucherHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID    int64 `json:"user_id"`
		VoucherID int64 `json:"voucher_id"`
		Quantity  int   `json:"quantity"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userVoucher := &data.UserVoucher{
		UserID:    input.UserID,
		VoucherID: input.VoucherID,
		Quantity:  input.Quantity,
	}

	v := validator.New()

	if data.ValidateUserVoucher(v, userVoucher); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.UserVouchers.Insert(userVoucher)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateKeyValue):
			app.violateUniqueConstraint(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/user_vouchers/%d", userVoucher.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), userVoucher, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
