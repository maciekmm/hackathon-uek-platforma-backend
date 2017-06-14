package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/utils"
)

type ChannelType string

var (
	ErrSubscriptionsUnknown         = errors.New("unknown error")
	ErrSubscriptionChannelInvalid   = errors.New("invalid channel")
	ErrSubscriptionChannelIDInvalid = errors.New("invalid channel id")
	ErrSubscriptionIDInvalid        = errors.New("invalid subscription id")
)

type Subscription struct {
	gorm.Model
	UserID          uint          `json:"user_id,omitempty"`
	MinimumPriority EventPriority `json:"priority,omitempty" gorm:"default:0"`
	Channel         ChannelType   `json:"channel,omitempty"`
	ChannelID       string        `json:"channel_id,omitempty"`
}

func (s *Subscription) Add(db *gorm.DB) error {
	errors := []error{}
	if len(s.Channel) == 0 {
		errors = append(errors, ErrSubscriptionChannelInvalid)
	}

	if len(s.ChannelID) == 0 {
		errors = append(errors, ErrSubscriptionChannelIDInvalid)
	}

	if len(errors) > 0 {
		return utils.NewErrorResponse(errors...)
	}

	if res := db.Create(&s); res.Error != nil {
		return &utils.ErrorResponse{
			Errors:      []string{ErrSubscriptionsUnknown.Error()},
			DebugErrors: []string{res.Error.Error()},
		}
	}
	return nil
}
