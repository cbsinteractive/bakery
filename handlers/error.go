package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

//ErrorResponse holds the errore response message
type ErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
	Err     error
}

//NewErrorResponse holds a formatted error response
func NewErrorResponse(message string, err error) ErrorResponse {
	errList := strings.Split(err.Error(), ": ")
	errMap := make(map[string][]string)
	errMap[errList[0]] = errList[1:]
	return ErrorResponse{
		Message: message,
		Errors:  errMap,
		Err:     err,
	}
}

// HandleError will both log and handle the http error for a given error response
func (e *ErrorResponse) HandleError(log *logrus.Entry, w http.ResponseWriter, code int) {
	logError(log, e.Message, e.Err)
	httpError(w, code, *e)
}

func logError(log *logrus.Entry, message string, err error) {
	log.WithError(err).Infof(message)
}

func httpError(w http.ResponseWriter, code int, e ErrorResponse) {
	eResp, err := json.Marshal(e)
	if err != nil {
		http.Error(w, e.Message+": "+e.Err.Error(), code)
		return
	}
	http.Error(w, string(eResp), code)
}
