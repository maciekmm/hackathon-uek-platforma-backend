package controllers

import (
	"errors"
	"net/http"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/middleware"
	"github.com/maciekmm/uek-bruschetta/models"
)

var (
	ErrSubscriptionsUnknown         = errors.New("unknown error")
	ErrSubscriptionChannelInvalid   = errors.New("invalid channel")
	ErrSubscriptionChannelIdInvalid = errors.New("invalid channel id")
)

type Subscriptions struct {
	Database *gorm.DB
}

func (s *Subscriptions) Register(router *mux.Router) {
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleAdd))).Methods("POST")
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleDelete))).Methods("DELETE")
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(s.HandleGetAll))).Methods("GET")
}

func (s *Subscriptions) HandleAdd(rw http.ResponseWriter, r *http.Request) {
	_, claims, _ := middleware.ParseToken(r)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	subscription := models.Subscription{}
	if err := decoder.Decode(&subscription); err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}
	subscription.UserId = claims.User.ID

	errors := []error{}
	if len(subscription.Channel) == 0 {
		errors = append(errors, ErrSubscriptionChannelInvalid)
	}

	if len(subscription.ChannelId) == 0 {
		errors = append(errors, ErrSubscriptionChannelIdInvalid)
	}

	if len(errors) > 0 {
		middleware.NewErrorResponse(errors...).Write(http.StatusBadRequest, rw)
		return
	}

	dbSubscription := models.Subscription{}
	res := s.Database.Where("channel = ? AND user_id = ?", subscription.Channel, claims.User.ID).First(&dbSubscription)
	if !res.RecordNotFound() {
		if res := s.Database.Model(&dbSubscription).Updates(&subscription); res.Error != nil {
			(&middleware.ErrorResponse{
				Errors:      []string{ErrSubscriptionsUnknown.Error()},
				DebugErrors: []string{res.Error.Error()},
			}).Write(http.StatusInternalServerError, rw)
			return
		}
	} else {
		if res := s.Database.Create(&subscription); res.Error != nil {
			(&middleware.ErrorResponse{
				Errors:      []string{ErrSubscriptionsUnknown.Error()},
				DebugErrors: []string{res.Error.Error()},
			}).Write(http.StatusInternalServerError, rw)
			return
		}
	}
	rw.WriteHeader(200)
}

func (s *Subscriptions) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	_, claims, _ := middleware.ParseToken(r)
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	subscription := models.Subscription{}
	if err := decoder.Decode(&subscription); err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	subscription.UserId = claims.User.ID
	if res := s.Database.Set("gorm:delete_option", "OPTION (OPTIMIZE FOR UNKNOWN)").Where("(id = ?) OR (user_id = ? AND channel = ?)", subscription.ID, claims.User.ID, subscription.Channel).Delete(&models.Subscription{}); res.Error != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(200)
}

func (s *Subscriptions) HandleGetAll(rw http.ResponseWriter, r *http.Request) {
	_, claims, _ := middleware.ParseToken(r)
	subs := []models.Subscription{}
	if res := s.Database.Where("user_id = ?", claims.User.ID).Find(&subs); res.Error != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(200)
	byt, err := json.Marshal(&subs)
	if err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.Write(byt)
}
