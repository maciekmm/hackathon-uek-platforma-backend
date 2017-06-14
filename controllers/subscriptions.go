package controllers

import (
	"net/http"
	"strconv"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/middleware"
	"github.com/maciekmm/uek-bruschetta/models"
	"github.com/maciekmm/uek-bruschetta/utils"
)

type Subscriptions struct {
	Database *gorm.DB
}

func (s *Subscriptions) Register(router *mux.Router) {
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleAdd))).Methods(http.MethodPost)
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleGetAll))).Methods(http.MethodGet)
	router.Handle("/{id:[0-9]+}/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleDelete))).Methods(http.MethodDelete)
	router.Handle("/{id:[0-9]+}/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandlePatch))).Methods(http.MethodPatch)

}

func (s *Subscriptions) HandleAdd(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	subscription := models.Subscription{}
	if err := decoder.Decode(&subscription); err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}
	subscription.UserID = user.ID
	if err := subscription.Add(s.Database); err != nil {
		if res, ok := err.(*utils.ErrorResponse); ok {
			res.Write(http.StatusBadRequest, rw)
			return
		}
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Subscriptions) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionIDInvalid.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	if res := s.Database.Unscoped().Where("id = ? AND user_id = ?", uint(id), user.ID).Delete(&models.Subscription{}); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Subscriptions) HandlePatch(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionIDInvalid.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	sub := models.Subscription{}
	if err := decoder.Decode(&sub); err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	sub.UserID = user.ID
	model := models.Subscription{}
	model.ID = uint(id)
	model.UserID = user.ID

	if res := s.Database.Model(&model).Updates(&sub); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Subscriptions) HandleGetAll(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)

	subs := []models.Subscription{}
	if res := s.Database.Where("user_id = ?", user.ID).Find(&subs); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	byt, err := json.Marshal(&subs)
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(byt)
}
