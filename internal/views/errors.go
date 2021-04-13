package views

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/noelruault/auction-bid-tracker/internal/models"
	"github.com/noelruault/auction-bid-tracker/internal/web"
)

// Error is a view that converts errors into API HTTP responses.
type Error struct {
	codes map[string]int
}

// JSON returns a JSON document with an error response to a requester.
//
// In case err has a "Public() string" method, it returns by default an HTTP Bad Request code and the
// JSON "error" field receives the result of calling Public().
// In case err does not have a "Public() string" method, it returns an HTTP Internal Server
// Error code and the JSON "error" field receives a "server_error" value.
//
// In case err is a models.ValidationError, it returns by default an HTTP Bad Request doce an error code of "validation_error"
// is returned, and the specific errors for each field are included as the value of the JSON "fields" field.
func (e Error) JSON(ctx context.Context, w http.ResponseWriter, err error) {
	// set the defaults we are going to return
	status := http.StatusInternalServerError
	data := map[string]interface{}{"error": "server_error"}

	// if it is a public error, must check if there's a different HTTP code set in the map
	if pe, ok := err.(models.PublicError); ok {
		status = http.StatusBadRequest

		public := pe.Public()
		detail := pe.Detail()
		data["error"] = public
		data["message"] = detail

		if s := e.codes[public]; s != 0 {
			status = s
		}
	}

	// if it's a validation error, we also need to check for codes and also add the fields to the output
	if ve, ok := err.(models.ValidationError); ok {
		status = http.StatusBadRequest
		data["error"] = "validation_error"
		var fields []map[string]string

		for field, err := range ve {
			public := err.Public()
			detail := err.Detail()

			if s := e.codes[public]; s != 0 {
				status = s
			}

			jsonField := make(map[string]string)
			jsonField["field"] = field
			jsonField["code"] = public
			jsonField["message"] = detail

			fields = append(fields, jsonField)
		}

		data["fields"] = fields
	}

	// Server log
	log.Printf("err: %s data: %v", err.Error(), data)

	// FIXME: Don't show server_error messages on production environment
	if data["error"] == "server_error" {
		data["message"] = fmt.Sprintf("unhandled_error: %v.", err.Error())
	}

	web.Respond(ctx, w, data, status)
}
