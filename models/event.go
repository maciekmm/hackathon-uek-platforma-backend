package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/utils"
)

type EventPriority int

const (
	EventPriorityLow EventPriority = iota
	EventPriorityMedium
	EventPriorityHigh
)

var (
	ErrEventsUnknown                   = errors.New("unknown error")
	ErrEventIDInvalid                  = errors.New("invalid id")
	ErrEventDescriptionInvalid         = errors.New("invalid description")
	ErrEventNameInvalid                = errors.New("invalid name")
	ErrEventNotificationMessageInvalid = errors.New("invalid notification message")
)

type EventPipe interface {
	Send(*Event) error
}

type Event struct {
	gorm.Model
	UserID              uint          `json:"user_id,omitempty"`
	Image               string        `json:"image,omitempty"`
	Name                string        `json:"name,omitempty"`
	Description         string        `json:"description,omitempty"`
	NotificationMessage string        `json:"message,omitempty"`
	Priority            EventPriority `json:"priority,omitempty"`
	Group               *uint         `json:"group,omitempty"`
}

func (event *Event) Add(db *gorm.DB, coord EventPipe) error {
	errs := []error{}
	if len(event.Description) == 0 {
		errs = append(errs, ErrEventDescriptionInvalid)
	}
	if len(event.Name) == 0 {
		errs = append(errs, ErrEventNameInvalid)
	}
	if len(event.NotificationMessage) == 0 {
		errs = append(errs, ErrEventNotificationMessageInvalid)
	}

	if len(errs) > 0 {
		return utils.NewErrorResponse(errs...)
	}

	dbEvent := Event{}
	res := db.Create(event)

	// update if record already exists, this should be done using PATCH or PUT methods, but it's easier to do it this way
	if !res.RecordNotFound() {
		if res := db.Model(&dbEvent).Updates(&event); res.Error != nil {
			return (&utils.ErrorResponse{
				Errors:      []string{ErrEventsUnknown.Error()},
				DebugErrors: []string{res.Error.Error()},
			})
		}
	} else {
		if res := db.Create(&event); res.Error != nil {
			return (&utils.ErrorResponse{
				Errors:      []string{ErrEventsUnknown.Error()},
				DebugErrors: []string{res.Error.Error()},
			})
		}
	}

	// verify completeness of provided model
	if err := coord.Send(event); err != nil {
		return (&utils.ErrorResponse{
			Errors:      []string{errors.New("event was added, but sending notifications failed").Error()},
			DebugErrors: []string{err.Error()},
		})
	}
	return nil
}
