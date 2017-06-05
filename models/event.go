package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type EventPriority int

const (
	EventPriorityLow EventPriority = iota
	EventPriorityMedium
	EventPriorityHigh
)

type Event struct {
	gorm.Model
	User                User
	UserId              uint          `json:"user_id,omitempty"`
	Name                string        `json:"name,omitempty"`
	Description         string        `json:"description,omitempty"`
	DateTakingPlace     time.Time     `json:"date_taking_place,omitempty"`
	NotificationMessage string        `json:"message,omitempty"`
	Priority            EventPriority `json:"priority,omitempty"`
}
