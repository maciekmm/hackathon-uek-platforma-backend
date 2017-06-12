package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type ErrorResponse struct {
	Errors      []string `json:"errors"`
	DebugErrors []string `json:"debug_errors,omitempty"`
}

func NewErrorResponse(errors ...error) *ErrorResponse {
	resp := &ErrorResponse{}
	for _, err := range errors {
		resp.Errors = append(resp.Errors, err.Error())
	}
	return resp
}

func (re *ErrorResponse) Error() string {
	return strings.Join(re.Errors, ", ")
}

func (re *ErrorResponse) String() string {
	return re.Error()
}

func (re *ErrorResponse) Write(code int, rw http.ResponseWriter) {
	rw.WriteHeader(code)
	if os.Getenv("DEBUG") != "TRUE" {
		re.DebugErrors = []string{}
	}
	body, err := json.Marshal(re)
	if err != nil {
		panic(err)
	}
	if _, err := rw.Write(body); err != nil {
		panic(err)
	}
}
