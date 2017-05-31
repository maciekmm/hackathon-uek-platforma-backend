package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Event struct {
	gorm.Model
	User            User
	UserID          uint      `json:"user_id,omitempty"`
	Name            string    `json:"name,omitempty"`
	Description     string    `json:"description,omitempty"`
	DateTakingPlace time.Time `json:"date_taking_place,omitempty"`
}
