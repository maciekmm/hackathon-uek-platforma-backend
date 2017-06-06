package models

import (
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
	UserID              uint          `json:"user_id,omitempty"`
	Image               string        `json:"image,omitempty"`
	Name                string        `json:"name,omitempty"`
	Description         string        `json:"description,omitempty"`
	NotificationMessage string        `json:"message,omitempty"`
	Priority            EventPriority `json:"priority,omitempty"`
}
