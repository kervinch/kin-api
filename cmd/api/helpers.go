package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/julienschmidt/httprouter"
	"github.com/kervinch/internal/data"
	"github.com/kervinch/internal/validator"
	"gorm.io/gorm"
)

type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app *application) readSlugParam(r *http.Request) string {
	params := httprouter.ParamsFromContext(r.Context())

	slug := params.ByName("slug")

	return slug
}

func (app *application) writeJSON(w http.ResponseWriter, status int, message string, data interface{}, headers http.Header) error {
	result := make(map[string]interface{})

	result["code"] = status
	result["message"] = message
	if status >= 400 {
		result["error"] = data
	} else {
		result["data"] = data
	}

	js, err := json.Marshal(result)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) writeJSONWithMeta(w http.ResponseWriter, status int, message string, data interface{}, headers http.Header, metadata interface{}) error {
	result := make(map[string]interface{})

	result["code"] = status
	result["message"] = message
	if status >= 400 {
		result["error"] = data
	} else {
		result["data"] = data
	}
	if metadata != nil {
		result["metadata"] = metadata
	}

	js, err := json.Marshal(result)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readJSONAllowUnknownFields(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readStrings(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		fn()
	}()
}

func (app *application) Paginate(w http.ResponseWriter, r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var input struct {
			Page     int
			PageSize int
		}

		v := validator.New()
		qs := r.URL.Query()

		input.Page = app.readInt(qs, "page", 1, v)
		input.PageSize = app.readInt(qs, "page_size", 20, v)

		if data.ValidatePagination(v, input); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return nil
		}

		page := input.Page
		if page == 0 {
			page = 1
		}

		pageSize := input.PageSize
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func (app *application) slugify(text string) string {
	slugged := slug.Make(text)
	return slugged
}

func (app *application) split(text string) ([]int64, error) {
	str := strings.Split(text, ",")
	slc := make([]int64, len(str))
	var value int64

	for i := range slc {
		tmp, err := strconv.Atoi(str[i])
		if err != nil {
			return []int64{}, err
		}
		value = int64(tmp)
		slc[i] = value
	}

	return slc, nil
}

func (app *application) appendIfMissing(slice []int64, i int64) []int64 {
	for _, ok := range slice {
		if ok == i {
			return slice
		}
	}

	return append(slice, i)
}

func (app *application) generateInvoiceNumber(userID int64, orderID int64, brandID int64) string {
	// Format: userID/orderID/brandID/dateTime

	invoiceNumber := fmt.Sprintf("%s/%s/%s/%s", strconv.Itoa(int(userID)), strconv.Itoa(int(orderID)), strconv.Itoa(int(brandID)), strconv.Itoa(int(time.Now().Unix())))

	return invoiceNumber
}
