package middleware

import (
	"errors"
	"net/http"
	"os"

	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/maciekmm/uek-bruschetta/models"
)

var (
	ErrAuthInvalidToken = errors.New("invalid token")
	ErrAuthNoPermission = errors.New("inferior user-role")
	ErrAuthUnknown      = errors.New("unknown error")
)

type AuthClaims struct {
	jwt.StandardClaims
	User *models.User
}

func ParseToken(req *http.Request) (*jwt.Token, error) {
	return jwt.ParseWithClaims(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer "), &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
}

func RequiresAuth(role models.UserRole, h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		tok, err := ParseToken(req)

		if err != nil || !tok.Valid {
			NewErrorResponse(ErrAuthInvalidToken, err).Write(http.StatusUnauthorized, rw)
			return
		}

		if claims, ok := tok.Claims.(*AuthClaims); ok {
			if claims.User.Role < role {
				NewErrorResponse(ErrAuthNoPermission).Write(http.StatusUnauthorized, rw)
				return
			}
		} else {
			NewErrorResponse(ErrAuthUnknown).Write(http.StatusInternalServerError, rw)
			return
		}

		h.ServeHTTP(rw, req)
	})
}
