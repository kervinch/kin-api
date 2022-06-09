package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/validator"
	"github.com/xendit/xendit-go"
)

// ====================================================================================
// Backoffice Handlers
// ====================================================================================

func (app *application) listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	var pagination data.Pagination

	v := validator.New()
	qs := r.URL.Query()

	pagination.Page = app.readInt(qs, "page", 1, v)
	pagination.PageSize = app.readInt(qs, "page_size", 20, v)

	orders, metadata, err := app.gorm.Orders.GetAll(pagination)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), orders, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showOrderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	order, err := app.gorm.Orders.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), order, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// ====================================================================================
// Business Handlers
// ====================================================================================

func (app *application) getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		Status string
		data.Pagination
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Status = app.readStrings(qs, "status", "")
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)

	if data.ValidatePagination(v, input.Pagination); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	orders, metadata, err := app.gorm.Orders.GetAPI(input.Pagination, user.ID, input.Status)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSONWithMeta(w, http.StatusOK, http.StatusText(http.StatusOK), orders, nil, metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createOrdersHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(data.DefaultMaxMemory)

	user := app.contextGetUser(r)

	var x struct {
		Customer        xendit.InvoiceCustomer
		CustomerAddress xendit.CustomerAddress
		InvoiceItem     []xendit.InvoiceItem
		InvoiceFee      []xendit.InvoiceFee
	}

	// =============
	// Product Logic
	// =============

	order := &data.Order{
		UserID:      user.ID,
		Receiver:    r.FormValue("receiver"),
		PhoneNumber: r.FormValue("phone_number"),
		City:        r.FormValue("city"),
		PostalCode:  r.FormValue("postal_code"),
		Address:     r.FormValue("address"),
		Status:      "awaiting_payment",
	}

	v := validator.New()

	if data.ValidateOrder(v, order); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	db := app.gorm.Transaction.DB
	tx := db.Begin()

	orderID, err := app.gorm.Orders.InsertWithTx(order, tx)
	if err != nil {
		tx.Rollback()
		app.serverErrorResponse(w, r, err)
		return
	}

	x.Customer = xendit.InvoiceCustomer{
		GivenNames:   user.Name,
		Email:        user.Email,
		MobileNumber: user.PhoneNumber,
		Address:      r.FormValue("address"),
	}

	x.CustomerAddress = xendit.CustomerAddress{
		Country:     "Indonesia",
		StreetLine1: r.FormValue("address"),
		City:        r.FormValue("city"),
		PostalCode:  r.FormValue("postal_code"),
	}

	// =============
	// Input Logic
	// =============

	var input struct {
		ProductDetail []*data.ProductDetail
		Quantity      []int
	}

	var brandID []int64

	count, err := strconv.Atoi(r.FormValue("count"))
	if err != nil {
		tx.Rollback()
		app.badRequestResponse(w, r, err)
		return
	}

	for i := 0; i < count; i++ {
		pdid, err := strconv.Atoi(r.FormValue("product_detail_id_" + strconv.Itoa(i)))
		if err != nil {
			tx.Rollback()
			app.badRequestResponse(w, r, err)
			return
		}

		quantity, err := strconv.Atoi(r.FormValue("quantity_" + strconv.Itoa(i)))
		if err != nil {
			tx.Rollback()
			app.badRequestResponse(w, r, err)
			return
		}

		pd, err := app.gorm.ProductDetails.Get(int64(pdid))
		if err != nil {
			tx.Rollback()
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		input.ProductDetail = append(input.ProductDetail, pd)
		input.Quantity = append(input.Quantity, quantity)
		brandID = app.appendIfMissing(brandID, pd.Product.BrandID)
	}

	if len(input.ProductDetail) != count {
		tx.Rollback()
		app.badRequestResponse(w, r, err)
		return
	}

	// ===================================
	// Order Detail & Invoice Detail Logic
	// ===================================

	var odids []int64

	for _, bid := range brandID {
		orderDetail := &data.OrderDetail{
			OrderID:       orderID,
			BrandID:       bid,
			InvoiceNumber: app.generateInvoiceNumber(user.ID, orderID, bid),
			Status:        "awaiting_payment",
		}

		v = validator.New()

		if data.ValidateOrderDetail(v, orderDetail); !v.Valid() {
			tx.Rollback()
			app.failedValidationResponse(w, r, v.Errors)
			return
		}

		orderDetailID, err := app.gorm.OrderDetails.InsertWithTx(orderDetail, tx)
		if err != nil {
			tx.Rollback()
			app.serverErrorResponse(w, r, err)
			return
		}

		odids = append(odids, orderDetailID)

		for i := 0; i < count; i++ {
			if input.ProductDetail[i].Product.BrandID == bid {
				invoiceDetail := &data.InvoiceDetail{
					OrderDetailID:   orderDetailID,
					ProductDetailID: input.ProductDetail[i].ID,
					ProductName:     input.ProductDetail[i].Product.Name,
					Quantity:        input.Quantity[i],
					Price:           input.ProductDetail[i].Price,
					Total:           int64(input.Quantity[i]) * input.ProductDetail[i].Price,
				}

				v = validator.New()

				if data.ValidateInvoiceDetail(v, invoiceDetail); !v.Valid() {
					tx.Rollback()
					app.failedValidationResponse(w, r, v.Errors)
					return
				}

				err := app.gorm.InvoiceDetails.InsertWithTx(invoiceDetail, tx)
				if err != nil {
					tx.Rollback()
					app.serverErrorResponse(w, r, err)
					return
				}

				// Xendit InvoiceItem logic
				invoiceItem := xendit.InvoiceItem{
					Name:     input.ProductDetail[i].Product.Name,
					Price:    float64(input.ProductDetail[i].Price),
					Quantity: input.Quantity[i],
				}

				x.InvoiceItem = append(x.InvoiceItem, invoiceItem)
			}
		}
	}

	// =============
	// Voucher Logic
	// =============

	var voucher *data.Voucher

	voucherID, _ := strconv.Atoi(r.FormValue("voucher_id"))

	if voucherID > 0 {
		voucher, err = app.gorm.Vouchers.GetByID(int64(voucherID))
		if err != nil {
			tx.Rollback()
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
	} else {
		voucher = &data.Voucher{}
	}

	tx.Commit()

	// ===========
	// Total Logic
	// ===========

	tx = db.Begin()

	for _, odid := range odids {
		od, err := app.gorm.OrderDetails.GetWithTx(odid, tx)
		if err != nil {
			tx.Rollback()
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		subtotal := 0
		total := 0

		for _, id := range od.InvoiceDetail {
			subtotal += int(id.Total)
		}

		if voucher.Type == "brand" && voucher.BrandID.Int64 == od.BrandID {
			if voucher.IsPercent {
				total = subtotal - (subtotal * voucher.Value / 100)
			} else {
				total = subtotal - voucher.Value
			}

			// Xendit InvoiceFee logic
			invoiceFee := xendit.InvoiceFee{
				Type:  "discount",
				Value: float64((subtotal - total) * -1),
			}

			x.InvoiceFee = append(x.InvoiceFee, invoiceFee)

			err = app.gorm.OrderDetails.SetTotalWithVoucherAndTx(od.ID, int64(subtotal), voucher.ID, int64(total), tx)
			if err != nil {
				tx.Rollback()
				app.serverErrorResponse(w, r, err)
				return
			}

			err = app.gorm.Vouchers.Consume(voucher.ID)
			if err != nil {
				tx.Rollback()
				app.serverErrorResponse(w, r, err)
				return
			}
		} else {
			total = subtotal
			err = app.gorm.OrderDetails.SetTotalWithTx(od.ID, int64(subtotal), int64(total), tx)
			if err != nil {
				tx.Rollback()
				app.serverErrorResponse(w, r, err)
				return
			}
		}
	}

	o, err := app.gorm.Orders.GetWithTx(orderID, tx)
	if err != nil {
		tx.Rollback()
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	subtotal := 0
	total := 0

	for _, od := range o.OrderDetail {
		subtotal += int(od.Total)
	}

	if voucher.Type == "total" {
		if voucher.IsPercent {
			total = subtotal - (subtotal * voucher.Value / 100)
		} else {
			total = subtotal - voucher.Value
		}

		err = app.gorm.Orders.SetTotalWithVoucherAndTx(o.ID, int64(subtotal), voucher.ID, int64(total), tx)
		if err != nil {
			tx.Rollback()
			app.serverErrorResponse(w, r, err)
			return
		}

		err = app.gorm.Vouchers.Consume(voucher.ID)
		if err != nil {
			tx.Rollback()
			app.serverErrorResponse(w, r, err)
			return
		}
	} else {
		total = subtotal
		err = app.gorm.Orders.SetTotalWithTx(o.ID, int64(subtotal), int64(total), tx)
		if err != nil {
			tx.Rollback()
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	tx.Commit()

	// Generate Xendit invoice
	notificationType := []string{"email", "sms"}
	invoice, err := app.xendit.GenerateInvoice(orderID, x.Customer, x.CustomerAddress, x.InvoiceItem, x.InvoiceFee, notificationType, total)
	if err != nil {
		app.failedInvoiceResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), envelope{"order": order, "invoice": invoice}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateOrdersHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		Status string `json:"status"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	order, err := app.gorm.Orders.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	order.Status = input.Status

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

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), order, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
