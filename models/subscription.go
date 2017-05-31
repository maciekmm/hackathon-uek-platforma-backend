package models

import "github.com/jinzhu/gorm"

type Subscription struct {
	gorm.Model
	User            User                 `json:"_"`
	UserID          uint                 `json:"user_id,omitempty"`
	MinimumPriority NotificationPriority `json:"priority,omitempty"`
	Categories      string               `json:"categories,omitempty"`
	Channel         ChannelType          `json:"channel,omitempty"`
	ChannelID       string               `json:"channel_id,omitempty"`
}
