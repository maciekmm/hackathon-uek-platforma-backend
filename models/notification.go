package models

import "github.com/jinzhu/gorm"

type NotificationPriority int

const (
	LowNotificationPriority NotificationPriority = iota
	MediumNotificationPriority
	HighNotificationPriority
)

type Notification struct {
	gorm.Model
	EventID  uint                 `json:"event_id,omitempty"`
	Channel  ChannelType          `json:"channel,omitempty"`
	Metadata string               `json:"metadata,omitempty"`
	Priority NotificationPriority `json:"priority,omitempty"`
}
