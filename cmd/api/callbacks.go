package main

import (
	"net/http"
	"os"

	"github.com/xendit/xendit-go"
)

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) invoiceCallbackHandler(w http.ResponseWriter, r *http.Request) {
	callbackToken := r.Header.Get("x-callback-token")

	if callbackToken != os.Getenv("XENDIT_CALLBACK_VERIFICATION_TOKEN") {
		app.notPermittedResponse(w, r)
		return
	}

	var callbackPayload struct {
		ID                           string                      `json:"id,omitempty"`
		ExternalID                   string                      `json:"external_id,omitempty"`
		UserID                       string                      `json:"user_id,omitempty"`
		IsHigh                       bool                        `json:"is_high,omitempty"`
		PaymentMethod                string                      `json:"payment_method,omitempty"`
		Status                       string                      `json:"status,omitempty"`
		MerchantName                 string                      `json:"merchant_name,omitempty"`
		Amount                       int                         `json:"amount,omitempty"`
		PaidAmount                   float64                     `json:"paid_amount,omitempty"`
		BankCode                     int                         `json:"bank_code,omitempty"`
		PaidAt                       string                      `json:"paid_at,omitempty"`
		PayerEmail                   string                      `json:"payer_email,omitempty"`
		Description                  string                      `json:"description,omitempty"`
		AdjustedReceivedAmount       float64                     `json:"adjusted_received_amount,omitempty"`
		FeesPaidAmount               float64                     `json:"fees_paid_amount,omitempty"`
		Updated                      string                      `json:"updated,omitempty"`
		Created                      string                      `json:"created,omitempty"`
		Currency                     string                      `json:"currency,omitempty"`
		PaymentChannel               string                      `json:"payment_channel,omitempty"`
		PaymentDestination           string                      `json:"payment_destination,omitempty"`
		MerchantProfileURL           string                      `json:"merchant_profile_url,omitempty"`
		PaymentDetail                xendit.InvoicePaymentDetail `json:"payment_detail,omitempty"`
		SuccessRedirectURL           string                      `json:"success_redirect_url,omitempty"`
		FailureRedirectURL           string                      `json:"failure_redirect_url,omitempty"`
		MidLabel                     string                      `json:"mid_label,omitempty"`
		CreditCardChargeID           string                      `json:"credit_card_charge_id,omitempty"`
		Item                         []xendit.InvoiceItem        `json:"items,omitempty"`
		Fee                          []xendit.InvoiceItem        `json:"fees,omitempty"`
		ShouldAuthenticateCreditCard bool                        `json:"should_authenticate_credit_card,omitempty"`
		RetailOutletName             string                      `json:"retail_outlet_name,omitempty"`
		EwalletType                  string                      `json:"ewallet_type,omitempty"`
		OnDemandLink                 string                      `json:"on_demand_link,omitempty"`
		RecurringPaymentID           string                      `json:"recurring_payment_id,omitempty"`
	}

	err := app.readJSON(w, r, &callbackPayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), envelope{"callback_status": "OK"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	return

	// orderID, err := strconv.ParseInt(callbackPayload.ExternalID, 10, 64)
	// if err != nil {
	// 	app.badRequestResponse(w, r, err)
	// 	return
	// }

	// err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), envelope{"external_id": callbackPayload.ExternalID}, nil)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// }

	// return

	// order, err := app.gorm.Orders.Get(orderID)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, data.ErrRecordNotFound):
	// 		app.notFoundResponse(w, r)
	// 	default:
	// 		app.serverErrorResponse(w, r, err)
	// 	}
	// 	return
	// }

	// order.Status = strings.ToLower(callbackPayload.Status)

	// err = app.gorm.Orders.Update(order)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, data.ErrRecordNotFound):
	// 		app.notFoundResponse(w, r)
	// 	default:
	// 		app.serverErrorResponse(w, r, err)
	// 	}
	// 	return
	// }

	// err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), envelope{"order_id": orderID}, nil)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// }
}
