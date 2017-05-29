package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Account struct {
}

func (a *Account) Register(router *mux.Router) {
	postRouter := router.Methods("POST").MatcherFunc(func(r *http.Request, match *mux.RouteMatch) bool {
		return true
	}).Subrouter()
	postRouter.HandleFunc("/register", a.HandleRegister)
	postRouter.HandleFunc("/login", a.HandleLogin)
	postRouter.HandleFunc("/forgotten", a.HandleForgotten)
}

func (a *Account) HandleRegister(rw http.ResponseWriter, r *http.Request) {

}

func (a *Account) HandleLogin(rw http.ResponseWriter, r *http.Request) {

}

func (a *Account) HandleForgotten(rw http.ResponseWriter, r *http.Request) {

}
