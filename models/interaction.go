package models

import "time"

type Interaction struct {
	ID             uint64    `json:"id,omitempty" db:"id"`
	Timestamp      time.Time `json:"timestamp,omitempty" db:"timestamp"`
	UserID         uint64    `json:"user_id,omitempty" db:"user_id"`
	NotificationID uint64    `json:"notification_id,omitempty" db:"notification_id"`
}
