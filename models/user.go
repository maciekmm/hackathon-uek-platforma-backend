package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name     string `json:"name,omitempty"`
	LastName string `json:"last_name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password []byte `json:"_"`
}
