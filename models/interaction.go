package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Interaction struct {
	gorm.Model
	Timestamp      time.Time `json:"timestamp,omitempty"`
	UserId         uint      `json:"user_id,omitempty"`
	User           User      `json:"-"`
	NotificationID uint      `json:"notification_id,omitempty"`
}
