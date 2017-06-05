package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/middleware"
	"github.com/maciekmm/uek-bruschetta/models"
)

type Events struct {
	Database *gorm.DB
}

func (e *Events) Register(router *mux.Router) {
	router.Handle("/", middleware.RequiresAuth(models.RoleAdmin, http.HandlerFunc(e.HandlePost))).Methods("POST")
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(e.HandleGetAll))).Methods("GET")
	router.Handle("/{id}/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(e.HandleGetSingle))).Methods("GET")
}

func (e *Events) HandlePost(rw http.ResponseWriter, r *http.Request) {
}

func (e *Events) HandleGetAll(rw http.ResponseWriter, r *http.Request) {
}

func (e *Events) HandleGetSingle(rw http.ResponseWriter, r *http.Request) {
}
