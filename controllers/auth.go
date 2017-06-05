package controllers

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/maciekmm/uek-bruschetta/models"
)

type AuthClaims struct {
	jwt.StandardClaims
	User *models.User
}
