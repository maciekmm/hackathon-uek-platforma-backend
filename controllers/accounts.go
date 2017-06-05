package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/middleware"
	"github.com/maciekmm/uek-bruschetta/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserEmailInvalid    = errors.New("email invalid")
	ErrUserPasswordInvalid = errors.New("password invalid")
	ErrUserNameInvalid     = errors.New("name invalid")
	ErrUserEmailRegistered = errors.New("mail already registered")
	ErrUserEmailNotFound   = errors.New("email not found")
	ErrAccountsUnknown     = errors.New("unknown error occured")
)

type jwtResponse struct {
	Token string `json:"token"`
}

type Accounts struct {
	Database *gorm.DB
}

func (a *Accounts) Register(router *mux.Router) {
	postRouter := router
	postRouter.HandleFunc("/register", a.HandleRegister)
	postRouter.HandleFunc("/login", a.HandleLogin)
}

func (a *Accounts) generateJWT(user *models.User) (string, error) {
	// generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.AuthClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
		},
		User: user,
	})
	tok, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return tok, fmt.Errorf("error occured while generating JWT: %s", err)
	}
	return tok, nil
}

func (a *Accounts) HandleRegister(rw http.ResponseWriter, r *http.Request) {
	// decode request
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	user := models.User{}
	if err := decoder.Decode(&user); err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrAccountsUnknown.Error(), fmt.Sprintf("could not decode request body")},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	// validate input
	errors := []string{}
	if len(user.Email) == 0 {
		errors = append(errors, ErrUserEmailInvalid.Error())
	}
	if len(user.Password) == 0 {
		errors = append(errors, ErrUserPasswordInvalid.Error())
	}
	if len(user.Name) == 0 {
		errors = append(errors, ErrUserNameInvalid.Error())
	}

	if len(errors) != 0 {
		(&middleware.ErrorResponse{
			Errors: errors,
		}).Write(http.StatusBadRequest, rw)
		return
	}

	// check if user already exists
	var existingUser models.User

	res := a.Database.First(&existingUser, "email = ?", user.Email)

	if !res.RecordNotFound() {
		middleware.NewErrorResponse(ErrUserEmailRegistered).Write(http.StatusBadRequest, rw)
		return
	}

	// encrypt password
	pwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrAccountsUnknown.Error()},
			DebugErrors: []string{fmt.Sprintf("could not encrypt user's password: %s", err)},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	user.Password = string(pwd)

	if err := a.Database.Create(&user).Error; err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrAccountsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	// clear the password for struct reuse
	user.Password = ""

	tok, err := a.generateJWT(&user)

	if err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrAccountsUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	// return JWT to client
	body, _ := json.Marshal(&jwtResponse{
		Token: tok,
	})
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}

func (a *Accounts) HandleLogin(rw http.ResponseWriter, r *http.Request) {
	// decode request
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	user := models.User{}
	if err := decoder.Decode(&user); err != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrAccountsUnknown.Error(), fmt.Sprintf("could not decode request body")},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	errors := []error{}

	if len(user.Email) == 0 {
		errors = append(errors, ErrUserEmailInvalid)
	}

	if len(user.Password) == 0 {
		errors = append(errors, ErrUserPasswordInvalid)
	}

	if len(errors) != 0 {
		middleware.NewErrorResponse(errors...).Write(http.StatusBadRequest, rw)
		return
	}

	var dbUser models.User
	res := a.Database.First(&dbUser, "email = ?", user.Email)

	if res.RecordNotFound() {
		middleware.NewErrorResponse(ErrUserEmailNotFound).Write(http.StatusBadRequest, rw)
		return
	}

	if res.Error != nil {
		(&middleware.ErrorResponse{
			Errors:      []string{ErrAccountsUnknown.Error()},
			DebugErrors: []string{fmt.Sprintf("error occured while querying the database: %s", res.Error.Error())},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		middleware.NewErrorResponse(ErrUserPasswordInvalid).Write(http.StatusBadRequest, rw)
		return
	}

	dbUser.Password = ""

	// generate JWT
	tok, err := a.generateJWT(&dbUser)
	if err != nil {
		(&middleware.ErrorResponse{
			Errors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	// return JWT to client
	body, _ := json.Marshal(&jwtResponse{
		Token: tok,
	})
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}
