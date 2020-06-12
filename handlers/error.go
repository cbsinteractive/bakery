package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cbsinteractive/bakery/logging"
)

//ErrorResponse holds the errore response message
type ErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
	Err     error               `json:"-"`
}

//NewErrorResponse holds a formatted error response
//Errors returned from the parser, origin, filter packages
//Will return in a `key: err` format. Where the key signals
//the package scope source of the error
func NewErrorResponse(message string, err error) ErrorResponse {
	errList := strings.Split(err.Error(), ": ")
	errMap := map[string][]string{
		errList[0]: errList[1:],
	}

	return ErrorResponse{
		Message: message,
		Errors:  errMap,
		Err:     err,
	}
}

// HandleError will both log and handle the http error for a given error response
func (e *ErrorResponse) HandleError(ctx context.Context, w http.ResponseWriter, code int) {
	logging.UpdateCtx(ctx, logging.Params{"error": fmt.Sprintf("%s: %v", e.Message, e.Err)})
	httpError(w, code, *e)
}

func httpError(w http.ResponseWriter, code int, e ErrorResponse) {
	eResp, err := json.Marshal(e)
	if err != nil {
		http.Error(w, e.Message+": "+e.Err.Error(), code)
		return
	}
	http.Error(w, string(eResp), code)
}
