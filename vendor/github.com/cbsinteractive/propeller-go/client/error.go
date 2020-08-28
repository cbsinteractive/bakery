package client

import (
	"fmt"
)

//StatusError object sets the error response
type StatusError struct {
	Code int
	Msg  string
	body string
}

//NotFound will set the status code to 404
func (e StatusError) NotFound() bool {
	return e.Code == 404
}

//Error will return a formatted status error
func (e StatusError) Error() string {
	return fmt.Sprintf("http status: %d: %q", e.Code, e.body)
}
