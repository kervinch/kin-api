package main

import (
	"database/sql"
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

func (app *application) listVouchersHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	productCategories, metadata, err := app.gorm.Vouchers.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), productCategories, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showVoucherHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	voucher, err := app.gorm.Vouchers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), voucher, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createVoucherHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	user := app.contextGetUser(r)

	var imageURL string
	var brandIDNullInt64 sql.NullInt64
	var logisticIDNullInt64 sql.NullInt64
	ch := make(chan string)

	file, handler, err := r.FormFile("voucher_image")
	if err == nil {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.VOUCHER, handler.Filename, handler.Header.Get("Content-Type"))
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

	brandID, _ := strconv.Atoi(r.FormValue("brand_id"))
	if brandID == 0 {
		brandIDNullInt64 = sql.NullInt64{
			Int64: int64(brandID),
			Valid: false,
		}
	} else {
		brandIDNullInt64 = sql.NullInt64{
			Int64: int64(brandID),
			Valid: true,
		}
	}

	logisticID, _ := strconv.Atoi(r.FormValue("logistic_id"))
	if logisticID == 0 {
		logisticIDNullInt64 = sql.NullInt64{
			Int64: int64(logisticID),
			Valid: false,
		}
	} else {
		logisticIDNullInt64 = sql.NullInt64{
			Int64: int64(logisticID),
			Valid: true,
		}
	}

	value, err := strconv.Atoi(r.FormValue("value"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	stock, err := strconv.Atoi(r.FormValue("stock"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	effectiveAt, err := time.Parse("2006-01-02", r.FormValue("effective_at"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	expiredAt, err := time.Parse("2006-01-02", r.FormValue("expired_at"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	voucher := &data.Voucher{
		Name:              r.FormValue("name"),
		Description:       r.FormValue("description"),
		TermsAndCondition: r.FormValue("terms_and_condition"),
		ImageURL:          imageURL,
		Slug:              app.slugify(r.FormValue("name")),
		Type:              r.FormValue("type"),
		IsActive:          r.FormValue("is_active") == "true",
		BrandID:           brandIDNullInt64,
		LogisticID:        logisticIDNullInt64,
		Code:              r.FormValue("code"),
		IsPercent:         r.FormValue("is_percent") == "true",
		Value:             value,
		Stock:             stock,
		EffectiveAt:       effectiveAt,
		ExpiredAt:         expiredAt,
		CreatedBy:         user.ID,
		UpdatedBy:         user.ID,
	}

	v := validator.New()

	if data.ValidateVoucher(v, voucher); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Vouchers.Insert(voucher)
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
	headers.Set("Location", fmt.Sprintf("/vouchers/%d", voucher.ID))

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), voucher, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateVoucherHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	voucher, err := app.gorm.Vouchers.Get(id)
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
	var brandIDNullInt64 sql.NullInt64
	var logisticIDNullInt64 sql.NullInt64
	ch := make(chan string)

	file, handler, err := r.FormFile("voucher_image")
	if err == nil && handler.Size > 0 {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.VOUCHER, handler.Filename, handler.Header.Get("Content-Type"))
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
		imageURL = voucher.ImageURL
	}

	brandID, _ := strconv.Atoi(r.FormValue("brand_id"))
	if brandID == 0 {
		brandIDNullInt64 = sql.NullInt64{
			Int64: int64(brandID),
			Valid: false,
		}
	} else {
		brandIDNullInt64 = sql.NullInt64{
			Int64: int64(brandID),
			Valid: true,
		}
	}

	logisticID, _ := strconv.Atoi(r.FormValue("logistic_id"))
	if logisticID == 0 {
		logisticIDNullInt64 = sql.NullInt64{
			Int64: int64(logisticID),
			Valid: false,
		}
	} else {
		logisticIDNullInt64 = sql.NullInt64{
			Int64: int64(logisticID),
			Valid: true,
		}
	}

	value, err := strconv.Atoi(r.FormValue("value"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	stock, err := strconv.Atoi(r.FormValue("stock"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	effectiveAt, err := time.Parse("2006-01-02", r.FormValue("effective_at"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	expiredAt, err := time.Parse("2006-01-02", r.FormValue("expired_at"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	voucher.Type = r.FormValue("type")
	voucher.Name = r.FormValue("name")
	voucher.Description = r.FormValue("description")
	voucher.TermsAndCondition = r.FormValue("terms_and_condition")
	voucher.ImageURL = imageURL
	voucher.BrandID = brandIDNullInt64
	voucher.LogisticID = logisticIDNullInt64
	voucher.Code = r.FormValue("code")
	voucher.IsPercent = r.FormValue("is_percent") == "true"
	voucher.Value = value
	voucher.Stock = stock
	voucher.IsActive = r.FormValue("is_active") == "true"
	voucher.Slug = app.slugify(r.FormValue("name"))
	voucher.EffectiveAt = effectiveAt
	voucher.ExpiredAt = expiredAt
	voucher.UpdatedBy = user.ID

	v := validator.New()

	if data.ValidateVoucher(v, voucher); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.Vouchers.Update(voucher)
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

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), voucher, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteVoucherHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.Vouchers.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "voucher successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

// func (app *application) getProductCategoriesHandler(w http.ResponseWriter, r *http.Request) {
// 	productCategories, err := app.gorm.ProductCategories.GetAPI()
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), productCategories, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }

// func (app *application) getProductCategoriesBySlugHandler(w http.ResponseWriter, r *http.Request) {
// 	slug := app.readSlugParam(r)

// 	productCategories, err := app.gorm.ProductCategories.GetBySlug(slug)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrRecordNotFound):
// 			app.notFoundResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), productCategories, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }
