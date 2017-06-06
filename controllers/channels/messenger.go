package channels

import (
	"fmt"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	messenger "github.com/maciekmm/messenger-platform-go-sdk"
)

type Messenger struct {
	Database  *gorm.DB
	messenger *messenger.Messenger
}

func (m *Messenger) Register(router *mux.Router) {
	m.messenger = &messenger.Messenger{
		VerifyToken: os.Getenv("FB_VERIFY_TOKEN"),
		AppSecret:   os.Getenv("FB_APP_SECRET"),
		AccessToken: os.Getenv("FB_ACCESS_TOKEN"),
	}
	m.messenger.Authentication = m.AuthenticationHandler
	m.messenger.MessageReceived = m.ReceivedHandler

	if os.Getenv("DEBUG") == "TRUE" {
		m.messenger.Debug = messenger.DebugAll
	}
	router.HandleFunc("/", m.messenger.Handler)
}

func (m *Messenger) AuthenticationHandler(event messenger.Event, opts messenger.MessageOpts, msg *messenger.Optin) {
	profile, err := m.messenger.GetProfile(opts.Sender.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = m.messenger.SendSimpleMessage(opts.Sender.ID, fmt.Sprintf("Hello, %s %s", profile.FirstName, profile.LastName))
	if err != nil {
		fmt.Println(err)
	}
}

func (m *Messenger) ReceivedHandler(event messenger.Event, opts messenger.MessageOpts, msg messenger.ReceivedMessage) {
	profile, err := m.messenger.GetProfile(opts.Sender.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = m.messenger.SendSimpleMessage(opts.Sender.ID, fmt.Sprintf("Hello, %s %s", profile.FirstName, profile.LastName))
	if err != nil {
		fmt.Println(err)
	}
}
