package models

import "github.com/jinzhu/gorm"

type ChannelType string

type Subscription struct {
	gorm.Model
	UserID          uint          `json:"user_id,omitempty"`
	MinimumPriority EventPriority `json:"priority,omitempty" gorm:"default:0"`
	Channel         ChannelType   `json:"channel,omitempty"`
	Year            uint          `json:"year,omitempty"`
	Department      string        `json:"department,omitempty"`
	ChannelID       string        `json:"channel_id,omitempty"`
}
