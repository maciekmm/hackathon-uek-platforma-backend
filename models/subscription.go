package models

import "github.com/jinzhu/gorm"

type Subscription struct {
	gorm.Model
	User            User          `json:"-"`
	UserId          uint          `json:"user_id,omitempty"`
	MinimumPriority EventPriority `json:"priority,omitempty" gorm:"default:0"`
	Channel         ChannelType   `json:"channel,omitempty"`
	ChannelId       string        `json:"channel_id,omitempty"`
}
