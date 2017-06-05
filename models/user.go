package models

import (
	"github.com/jinzhu/gorm"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	gorm.Model
	Name     string   `json:"name,omitempty"`
	Email    string   `json:"email" gorm:"index"`
	Role     UserRole `json:"role" gorm:"default:'user'"`
	Password string   `json:"password,omitempty"`
}
