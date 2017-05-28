package models

import (
	"github.com/maciekmm/uek-bruschetta/channel"
)

type Subscription struct {
	ID         uint64               `json:"id,omitempty" db:"id"`
	UserID     uint64               `json:"user_id,omitempty" db:"user_id"`
	MinimumPriority   NotificationPriority `json:"priority,omitempty" db:"priority"`
	Categories []string             `json:"categories,omitempty" db:"categories"`
	Channel    channel.Type         `json:"channel,omitempty" db:"channel"`
	ChannelID  string               `json:"channel_id,omitempty" db:"channel_id"`
}
