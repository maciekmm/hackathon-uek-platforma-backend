package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

func (re *ErrorResponse) Error() string {
	return strings.Join(re.Errors, ", ")
}

func (re *ErrorResponse) String() string {
	return re.Error()
}

func (re *ErrorResponse) Write(code int, rw http.ResponseWriter) {
	rw.WriteHeader(code)
	body, err := json.Marshal(re)
	if err != nil {
		panic(err)
	}
	if _, err := rw.Write(body); err != nil {
		panic(err)
	}
}
