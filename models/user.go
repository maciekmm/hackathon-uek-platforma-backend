package models

import (
	"github.com/jinzhu/gorm"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type UserError string

var (
	UserEmailInvalid    UserError = "email invalid"
	UserPasswordInvalid UserError = "password invalid"
	UserNameInvalid     UserError = "name invalid"
)

type User struct {
	gorm.Model
	Name     string   `json:"name,omitempty"`
	Email    string   `json:"email" gorm:"index"`
	Role     UserRole `json:"role" gorm:"default:'user'"`
	Password string   `json:"password,omitempty"`
}

func (u *User) VerifyConstraints() (errors []string) {
	if len(u.Email) == 0 {
		errors = append(errors, string(UserEmailInvalid))
	}
	if len(u.Password) == 0 {
		errors = append(errors, string(UserPasswordInvalid))
	}
	if len(u.Name) == 0 {
		errors = append(errors, string(UserNameInvalid))
	}
	return
}
