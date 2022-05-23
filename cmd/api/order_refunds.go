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

func (app *application) listOrderRefundsHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	orderRefunds, metadata, err := app.gorm.OrderRefunds.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), orderRefunds, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showOrderRefundHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	orderRefund, err := app.gorm.OrderRefunds.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), orderRefund, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) createOrderRefundHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	user := app.contextGetUser(r)

	var imageURL1 string
	var imageURL2 string
	var imageURL3 string
	var videoURL string
	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)
	ch4 := make(chan string)

	file, handler, err := r.FormFile("refund_image_1")
	if err == nil {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.PRODUCT_REFUND, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch1 <- url
			close(ch1)
		})
		imageURL1 = <-ch1
		defer file.Close()
	} else {
		imageURL1 = ""
	}

	file, handler, err = r.FormFile("refund_image_2")
	if err == nil {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.PRODUCT_REFUND, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch2 <- url
			close(ch2)
		})
		imageURL2 = <-ch2
		defer file.Close()
	} else {
		imageURL2 = ""
	}

	file, handler, err = r.FormFile("refund_image_3")
	if err == nil {
		app.background(func() {
			url, err := app.s3.Upload(file, s3.PRODUCT_REFUND, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch3 <- url
			close(ch3)
		})
		imageURL3 = <-ch3
		defer file.Close()
	} else {
		imageURL3 = ""
	}

	file, handler, err = r.FormFile("refund_video")
	if err == nil {
		app.background(func() {
			url, err := app.s3.UploadVideo(file, s3.PRODUCT_REFUND, handler.Filename, handler.Header.Get("Content-Type"))
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			ch4 <- url
			close(ch4)
		})
		videoURL = <-ch4
		defer file.Close()
	} else {
		videoURL = ""
	}

	orderDetailID, err := strconv.Atoi(r.FormValue("order_detail_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	brandID, err := strconv.Atoi(r.FormValue("brand_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	explanation := r.FormValue("explanation")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	orderRefund := &data.OrderRefund{
		UserID:        user.ID,
		OrderDetailID: int64(orderDetailID),
		BrandID:       int64(brandID),
		Image1:        imageURL1,
		Image2:        imageURL2,
		Image3:        imageURL3,
		Video:         videoURL,
		Explanation:   explanation,
	}

	v := validator.New()

	if data.ValidateOrderRefund(v, orderRefund); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.OrderRefunds.Insert(orderRefund)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), orderRefund, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getOrderRefundHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	orderRefunds, metadata, err := app.gorm.OrderRefunds.GetAPI(pagination, user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), orderRefunds, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
