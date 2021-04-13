package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

// validate holds the settings and caches for validating request struct values.
var validate = validator.New()

// Decode reads the body of an HTTP request looking for a JSON document. The
// body is decoded into the provided value.
//
// If the provided value is a struct then it is checked for validation tags.
func Decode(r *http.Request, val interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // return an error when the destination is a struct and the input
	// contains object keys which do not match the destination.

	if err := decoder.Decode(val); err != nil {
		if strings.Contains(err.Error(), "json: unknown field") { // Used alongside DisallowUnknownFields to return an idiomatic error
			uf := strings.Trim(strings.ReplaceAll(err.Error(), "json: unknown field ", ""), "\"") // Gets the unknown field name from the error
			return fmt.Errorf("json error: invalid field %s", uf)
		}
		return fmt.Errorf("json error: invalid format")
	}

	if err := validate.Struct(val); err != nil {

		// Use a type assertion to get the real error value.
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		var fields []FieldError
		for _, verror := range verrors {
			field := FieldError{
				Field: verror.Field(),
				// Error: verror.Message(),
			}
			fields = append(fields, field)
		}

		return &Error{
			Err:    errors.New("field validation error"),
			Status: http.StatusBadRequest,
			Fields: fields,
		}
	}

	return nil
}
