package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Interaction struct {
	gorm.Model
	Timestamp      time.Time `json:"timestamp,omitempty"`
	UserID         uint      `json:"user_id,omitempty"`
	NotificationID uint      `json:"notification_id,omitempty"`
}
