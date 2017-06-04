package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/models"
	"github.com/maciekmm/uek-bruschetta/utils"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Logger   *log.Logger
	Database *gorm.DB
}

func (a *Account) Register(router *mux.Router) {
	postRouter := router
	postRouter.HandleFunc("/register", a.HandleRegister)
	postRouter.HandleFunc("/login", a.HandleLogin)
}

func (a *Account) HandleRegister(rw http.ResponseWriter, r *http.Request) {
	// decode request
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	user := models.User{}
	if err := decoder.Decode(&user); err != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("could not decode request body: %s", err.Error())},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	if errors := user.VerifyConstraints(); len(errors) != 0 {
		(&ErrorResponse{
			Errors: []string(errors),
		}).Write(http.StatusBadRequest, rw)
		return
	}

	// check if user already exists
	var existingUser models.User

	res := a.Database.First(&existingUser, "email = ?", user.Email)

	if !res.RecordNotFound() {
		(&ErrorResponse{
			Errors: []string{"account with this email already exists"},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	// encrypt password
	pwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("could not encrypt user's password: %s", err)},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	user.Password = string(pwd)

	if err := a.Database.Create(&user).Error; err != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("could not register username, db error: %s", err)},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	// clear the password for struct reuse
	user.Password = ""

	// generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, utils.AuthClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix(),
		},
		User: user,
	})
	tok, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("error occured while generating JWT: %s", err)},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	// return JWT to client
	body, _ := json.Marshal(utils.JWT{
		Token: tok,
	})
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}

func (a *Account) HandleLogin(rw http.ResponseWriter, r *http.Request) {
	// decode request
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	user := models.User{}
	if err := decoder.Decode(&user); err != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("could not decode request body: %s", err.Error())},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	errors := user.VerifyConstraints()
	for _, er := range errors {
		if er == string(models.UserEmailInvalid) || er == string(models.UserPasswordInvalid) {
			(&ErrorResponse{
				Errors: []string{string(er)},
			}).Write(http.StatusBadRequest, rw)
			return
		}
	}

	var dbUser models.User
	res := a.Database.First(&dbUser, "email = ?", user.Email)

	if res.Error != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("error occured while querying the database: %s", res.Error.Error())},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	if res.RecordNotFound() {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("wrong user/password combination")},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("invalid password")},
		}).Write(http.StatusBadRequest, rw)
		return
	}

	dbUser.Password = ""

	// generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, utils.AuthClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix(),
		},
		User: dbUser,
	})
	tok, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		(&ErrorResponse{
			Errors: []string{fmt.Sprintf("error occured while generating JWT: %s", err)},
		}).Write(http.StatusInternalServerError, rw)
		return
	}

	// return JWT to client
	body, _ := json.Marshal(utils.JWT{
		Token: tok,
	})
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}
