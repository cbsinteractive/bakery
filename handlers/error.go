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
}

func httpError(logger *logrus.Logger, w http.ResponseWriter, err error, message string, code int) {
	errList := strings.Split(err.Error(), ": ")
	errMap := make(map[string][]string)
	errMap[errList[0]] = errList[1:]

	eResp, e := json.Marshal(ErrorResponse{
		Message: message,
		Errors:  errMap,
	})

	if e != nil {
		http.Error(w, message+": "+err.Error(), code)
	}

	logger.WithError(err).Infof(message)
	http.Error(w, string(eResp), code)
}
