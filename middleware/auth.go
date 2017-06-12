package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"

	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/maciekmm/uek-bruschetta/models"
	"github.com/maciekmm/uek-bruschetta/utils"
)

var (
	ErrAuthInvalidToken = errors.New("invalid token")
	ErrAuthNoPermission = errors.New("inferior user-role")
	ErrAuthUnknown      = errors.New("unknown error")
)

type ContextKey string

const (
	ContextUserKey ContextKey = "user"
)

type AuthClaims struct {
	jwt.StandardClaims
	User *models.User
}

func ParseToken(req *http.Request) (*jwt.Token, *AuthClaims, error) {
	tok, err := jwt.ParseWithClaims(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer "), &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return tok, nil, err
	} else if claims, ok := tok.Claims.(*AuthClaims); ok {
		return tok, claims, err
	}
	return tok, nil, err
}

func RequiresAuth(role models.UserRole, h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		tok, claims, err := ParseToken(req)

		if err != nil || !tok.Valid {
			utils.NewErrorResponse(ErrAuthInvalidToken, err).Write(http.StatusUnauthorized, rw)
			return
		}

		if claims != nil && claims.User.Role >= role {
			ctx := context.WithValue(req.Context(), ContextUserKey, claims.User)
			h.ServeHTTP(rw, req.WithContext(ctx))
		} else if claims != nil && claims.User.Role < role {
			utils.NewErrorResponse(ErrAuthNoPermission).Write(http.StatusUnauthorized, rw)
		} else {
			utils.NewErrorResponse(ErrAuthUnknown).Write(http.StatusInternalServerError, rw)
		}
	})
}
