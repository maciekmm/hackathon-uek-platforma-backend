package models

import (
	"time"
)

type Interaction struct {
	ID        uint         `json:"id,omitempty"`
	EventID   uint         `json:"event_id,omitempty" gorm:"index"`
	Timestamp time.Time    `json:"timestamp,omitempty"`
	UserID    uint         `json:"user_id,omitempty"`
	Channel   *ChannelType `json:"channel,omitempty"`
}
