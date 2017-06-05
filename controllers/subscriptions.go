package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/middleware"
	"github.com/maciekmm/uek-bruschetta/models"
)

type Subscriptions struct {
	Database *gorm.DB
}

func (s *Subscriptions) Register(router *mux.Router) {
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleAdd))).Methods("POST")
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleEdit))).Methods("PUT")
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleDelete))).Methods("DELETE")
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleGetAll))).Methods("GET")

}

func (s *Subscriptions) HandleAdd(rw http.ResponseWriter, r *http.Request) {

}

func (s *Subscriptions) HandleEdit(rw http.ResponseWriter, r *http.Request) {

}

func (s *Subscriptions) HandleDelete(rw http.ResponseWriter, r *http.Request) {

}

func (s *Subscriptions) HandleGetAll(rw http.ResponseWriter, r *http.Request) {

}
