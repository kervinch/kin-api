package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/kervinch/internal/data"
)

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) invoiceCallbackHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("x-callback-token", os.Getenv("XENDIT_CALLBACK_VERIFICATION_TOKEN"))

	var callbackPayload struct {
		ID                     string  `json:"id"`
		ExternalID             string  `json:"external_id"`
		UserID                 string  `json:"user_id"`
		IsHigh                 bool    `json:"is_high"`
		PaymentMethod          string  `json:"payment_method"`
		Status                 string  `json:"status"`
		MerchantName           string  `json:"merchant_name"`
		Amount                 float64 `json:"amount"`
		PaidAmount             float64 `json:"paid_amount"`
		BankCode               string  `json:"bank_code"`
		PaidAt                 string  `json:"paid_at"`
		PayerEmail             string  `json:"payer_email"`
		Description            string  `json:"description"`
		AdjustedReceivedAmount float64 `json:"adjusted_received_amount"`
		FeesPaidAmount         float64 `json:"fees_paid_amount"`
		Updated                string  `json:"updated"`
		Created                string  `json:"created"`
		Currency               string  `json:"currency"`
		PaymentChannel         string  `json:"payment_channel"`
		PaymentDestination     string  `json:"payment_destination"`
	}

	err := app.readJSON(w, r, &callbackPayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	orderID, err := strconv.ParseInt(callbackPayload.ExternalID, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	order, err := app.gorm.Orders.Get(orderID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	order.Status = strings.ToLower(callbackPayload.Status)

	err = app.gorm.Orders.Update(order)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), order, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) testInvoiceCallbackHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("x-callback-token", os.Getenv("XENDIT_CALLBACK_VERIFICATION_TOKEN"))

	var callbackPayload struct {
		ID                     string  `json:"id"`
		ExternalID             string  `json:"external_id"`
		UserID                 string  `json:"user_id"`
		IsHigh                 bool    `json:"is_high"`
		PaymentMethod          string  `json:"payment_method"`
		Status                 string  `json:"status"`
		MerchantName           string  `json:"merchant_name"`
		Amount                 float64 `json:"amount"`
		PaidAmount             float64 `json:"paid_amount"`
		BankCode               string  `json:"bank_code"`
		PaidAt                 string  `json:"paid_at"`
		PayerEmail             string  `json:"payer_email"`
		Description            string  `json:"description"`
		AdjustedReceivedAmount float64 `json:"adjusted_received_amount"`
		FeesPaidAmount         float64 `json:"fees_paid_amount"`
		Updated                string  `json:"updated"`
		Created                string  `json:"created"`
		Currency               string  `json:"currency"`
		PaymentChannel         string  `json:"payment_channel"`
		PaymentDestination     string  `json:"payment_destination"`
	}

	err := app.readJSON(w, r, &callbackPayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), nil, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
