package main

import (
	"database/sql"
	"fmt"

	"os"

	"net/http"

	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/maciekmm/uek-bruschetta/controllers"
)

func main() {
	app := &Application{}
	err := app.init()
	if err != nil {
		panic(err)
	}
	err = app.serve()
	if err != nil {
		panic(err)
	}

}

type Application struct {
	Database *sql.DB
	router   *mux.Router
}

func (a *Application) init() error {
	con, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@database/%s?sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB")))
	if err != nil {
		return err
	}
	err = con.Ping()
	for err != nil {
		err = con.Ping()
		fmt.Println(err)
		time.Sleep(1 * time.Second)
	}
	a.Database = con
	// setup routes
	a.router = mux.NewRouter()
	accountController := &controllers.Account{}
	accountController.Register(a.router.PathPrefix("/account/").Subrouter())

	return nil
}

func (a *Application) serve() error {
	return http.ListenAndServe(":3000", a.router)
}
