package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"os"

	"net/http"

	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/maciekmm/uek-bruschetta/channels"
	"github.com/maciekmm/uek-bruschetta/controllers"
	"github.com/maciekmm/uek-bruschetta/models"
)

func main() {
	logger := log.New(os.Stdout, "Bruschette", log.Ldate|log.Lshortfile)
	app := &Application{Logger: logger}

	err := app.init()
	if err != nil {
		logger.Fatal(err)
	}

	err = app.serve()
	if err != nil {
		logger.Fatal(err)
	}
}

type Application struct {
	Database           *gorm.DB
	Logger             *log.Logger
	ChannelCoordinator *channels.Coordinator
	router             *mux.Router
}

func (a *Application) init() error {
	a.Logger.Println("starting Bruschette")

	a.Logger.Println("setting up database connection")
	con, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@database/%s?sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB")))
	if err != nil {
		return fmt.Errorf("could not open database connection: %s", err.Error())
	}

	a.Logger.Println("establishing database connection")
	deadline := time.After(10 * time.Second)
out:
	for {
		select {
		case <-deadline:
			return fmt.Errorf("could not establish database connection, last error: %s", err.Error())
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err = con.PingContext(ctx)
			if err == nil {
				break out
			} else {
				a.Logger.Printf("pinging database failed: %s\n", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}
	// auto-migrating models
	a.Database, err = gorm.Open("postgres", con)
	a.Database.SetLogger(a.Logger)
	if err != nil {
		return err
	}

	a.Database.AutoMigrate(&models.User{}, &models.Event{}, &models.Interaction{}, &models.Subscription{})

	// setup channel coordinator
	messenger := &channels.Messenger{}
	a.ChannelCoordinator = channels.NewCoordinator(a.Logger, a.Database, messenger)
	go a.ChannelCoordinator.Start()

	// setup routes
	a.Logger.Println("setting up routes")
	a.router = mux.NewRouter()

	a.router.Methods("OPTIONS").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	// accounts
	accountController := &controllers.Accounts{Database: a.Database}
	accountController.Register(a.router.PathPrefix("/accounts/").Subrouter())

	// events
	eventsController := &controllers.Events{Database: a.Database, Coordinator: a.ChannelCoordinator}
	eventsController.Register(a.router.PathPrefix("/events/").Subrouter())

	// subscriptions
	subscriptionsController := &controllers.Subscriptions{Database: a.Database}
	subscriptionsController.Register(a.router.PathPrefix("/subscriptions/").Subrouter())

	// channels
	messenger.Register(a.router.PathPrefix("/channels/messenger/").Subrouter())

	// timetables
	timetable := &controllers.Timetable{}
	if err := timetable.Register(a.router.PathPrefix("/timetable/").Subrouter()); err != nil {
		return fmt.Errorf("could not register timetable endpoint: %s", err.Error())
	}
	return nil
}

func (a *Application) serve() error {
	server := http.Server{
		Addr:           ":3000",
		Handler:        a.router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return server.ListenAndServe()
}
