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

func (app *application) getUserAddressesHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	userAddresses, err := app.gorm.UserAddresses.GetAPI(user)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), userAddresses, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createUserAddressHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Receiver    string `json:"receiver"`
		PhoneNumber string `json:"phone_number"`
		City        string `json:"city"`
		PostalCode  string `json:"postal_code"`
		Address     string `json:"address"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	userAddress := &data.UserAddress{
		UserID:      user.ID,
		Name:        input.Name,
		Receiver:    input.Receiver,
		PhoneNumber: input.PhoneNumber,
		City:        input.City,
		PostalCode:  input.PostalCode,
		Address:     input.Address,
		IsMain:      false,
	}

	v := validator.New()

	if data.ValidateUserAddress(v, userAddress); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.UserAddresses.Insert(userAddress)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, http.StatusText(http.StatusCreated), userAddress, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateUserAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userAddress, err := app.gorm.UserAddresses.Get(id, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name        string `json:"name"`
		Receiver    string `json:"receiver"`
		PhoneNumber string `json:"phone_number"`
		City        string `json:"city"`
		PostalCode  string `json:"postal_code"`
		Address     string `json:"address"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userAddress.Name = input.Name
	userAddress.Receiver = input.Receiver
	userAddress.PhoneNumber = input.PhoneNumber
	userAddress.City = input.City
	userAddress.PostalCode = input.PostalCode
	userAddress.Address = input.Address

	v := validator.New()

	if data.ValidateUserAddress(v, userAddress); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.UserAddresses.Update(userAddress, user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), userAddress, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMainUserAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userAddress, err := app.gorm.UserAddresses.Get(id, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		IsMain bool `json:"is_main"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userAddress.IsMain = input.IsMain

	v := validator.New()

	if data.ValidateUserAddress(v, userAddress); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.gorm.UserAddresses.UpdateMain(userAddress, user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), userAddress, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteUserAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.gorm.UserAddresses.Delete(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, http.StatusText(http.StatusOK), "user address successfully deleted", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
