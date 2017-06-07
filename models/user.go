package models

import (
	"github.com/jinzhu/gorm"
)

type UserRole int

const (
	RoleUser UserRole = iota
	RoleAdmin
)

type User struct {
	gorm.Model
	Name     string   `json:"name,omitempty"`
	Email    string   `json:"email" gorm:"index"`
	Role     UserRole `json:"role" gorm:"default:0"`
	Password string   `json:"password,omitempty"`
	Group    *uint     `json:"group,omitempty"`
}
