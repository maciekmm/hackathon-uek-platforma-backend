package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/channels"
	"github.com/maciekmm/uek-bruschetta/middleware"
	"github.com/maciekmm/uek-bruschetta/models"
	"github.com/maciekmm/uek-bruschetta/utils"
)

type Events struct {
	Database    *gorm.DB
	Coordinator *channels.Coordinator
}

func (e *Events) Register(router *mux.Router) {
	router.Handle("/", middleware.RequiresAuth(models.RoleAdmin, http.HandlerFunc(e.HandleAdd))).Methods(http.MethodPost)
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(e.HandleGetAll))).Methods(http.MethodGet)
	router.Handle("/{id:[0-9]+}/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(e.HandleGetSingle))).Methods(http.MethodGet)
	router.Handle("/{id:[0-9]+}/", middleware.RequiresAuth(models.RoleAdmin, http.HandlerFunc(e.HandlePatchSingle))).Methods(http.MethodPatch)
	router.Handle("/{id:[0-9]+}/", middleware.RequiresAuth(models.RoleAdmin, http.HandlerFunc(e.HandlePutSingle))).Methods(http.MethodPut)
	router.Handle("/{id:[0-9]+}/", middleware.RequiresAuth(models.RoleAdmin, http.HandlerFunc(e.HandleDelete))).Methods(http.MethodDelete)
	router.Handle("/{id:[0-9]+}/interactions/", middleware.RequiresAuth(models.RoleAdmin, http.HandlerFunc(e.HandleGetInteractions))).Methods(http.MethodGet)
}

func (s *Events) HandleAdd(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	event := models.Event{}
	if err := decoder.Decode(&event); err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}
	event.UserID = user.ID

	if err := event.Add(s.Database); err != nil {
		if res, ok := err.(*utils.ErrorResponse); ok {
			res.Write(http.StatusBadRequest, rw)
			return
		}
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	// verify completeness of provided model
	if err := s.Coordinator.Send(&event); err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{errors.New("event was added, but sending notifications failed").Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusMultiStatus, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Events) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventIDInvalid.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	if res := s.Database.Delete(&models.Event{Model: gorm.Model{ID: uint(id)}}); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Events) HandleGetAll(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)
	events := []models.Event{}
	res := s.Database
	if user.Role != models.RoleAdmin {
		res = res.Where("\"group\" = ? OR \"group\" IS NULL", *user.Group)
	}
	if res := res.Find(&events); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	byt, err := json.Marshal(&events)
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(byt)
}

func (s *Events) HandleGetSingle(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)
	channel := r.URL.Query().Get("channel")

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventIDInvalid.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	// handle user interactions
	interaction := &models.Interaction{
		Timestamp: time.Now(),
		UserID:    user.ID,
		EventID:   uint(id),
	}

	if len(channel) > 0 {
		ch := models.ChannelType(channel)
		interaction.Channel = &ch
	}

	// save user interaction
	go func(db *gorm.DB, interaction *models.Interaction) {
		// this is just for statistics purposes, we don't care if it fails
		db.Create(interaction)
	}(s.Database, interaction)

	event := models.Event{}
	res := s.Database
	if user.Role != models.RoleAdmin {
		res = res.Where("\"group\" = ? OR \"group\" IS NULL", *user.Group)
	}
	if res := res.First(&event, uint(id)); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	event.UserID = user.ID

	byt, err := json.Marshal(&event)
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(byt)
}

func (s *Events) HandlePatchSingle(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventIDInvalid.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	event := models.Event{}
	if err := decoder.Decode(&event); err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	event.UserID = user.ID
	model := models.Event{}
	model.ID = uint(id)

	if res := s.Database.Model(&model).Updates(&event); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Events) HandlePutSingle(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.ContextUserKey).(*models.User)
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventIDInvalid.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	event := models.Event{}
	if err := decoder.Decode(&event); err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	event.ID = uint(id)
	event.UserID = user.ID
	if res := s.Database.Save(&event); res.Error != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (e *Events) HandleGetInteractions(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventIDInvalid.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}
	interactions := []models.Interaction{}

	e.Database.Where("event_id = ?", id).Find(&interactions)

	byt, err := json.Marshal(&interactions)
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{models.ErrEventsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(byt)
}
