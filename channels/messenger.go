package channels

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	messenger "github.com/maciekmm/messenger-platform-go-sdk"
	"github.com/maciekmm/messenger-platform-go-sdk/template"
	"github.com/maciekmm/uek-bruschetta/models"
)

const (
	ChannelTypeMessenger models.ChannelType = "messenger"
)

type Messenger struct {
	Logger    *log.Logger
	Database  *gorm.DB
	messenger *messenger.Messenger
}

func (m *Messenger) Type() models.ChannelType {
	return ChannelTypeMessenger
}

func (m *Messenger) Register(router *mux.Router) {
	m.messenger = &messenger.Messenger{
		VerifyToken: os.Getenv("FB_VERIFY_TOKEN"),
		AppSecret:   os.Getenv("FB_APP_SECRET"),
		AccessToken: os.Getenv("FB_ACCESS_TOKEN"),
	}
	m.messenger.Authentication = m.AuthenticationHandler

	if os.Getenv("DEBUG") == "TRUE" {
		m.messenger.Debug = messenger.DebugAll
	}

	router.HandleFunc("/", m.messenger.Handler)
}

func (m *Messenger) Send(sub *models.Subscription, event *models.Event) error {
	mq := messenger.MessageQuery{}
	mq.RecipientID(sub.ChannelID)
	mq.Template(template.GenericTemplate{
		Title:    event.Name,
		ImageURL: event.Image,
		Subtitle: event.NotificationMessage,
		Buttons: []template.Button{
			template.NewWebURLButton("Zobacz więcej", fmt.Sprintf("https://uek.kochanow.ski/#/dashboard/events/%d/messenger/", event.ID)),
		},
	})
	_, err := m.messenger.SendMessage(mq)
	return err
}

func (m *Messenger) AuthenticationHandler(event messenger.Event, opts messenger.MessageOpts, optin *messenger.Optin) {
	if optin == nil {
		return
	}
	fragments := strings.Split(optin.Ref, ":")
	if len(fragments) != 2 {
		m.Logger.Printf("number of fragments in '%s' is not equal to 2\n", optin.Ref)
		return
	}

	id, err := strconv.Atoi(fragments[0])
	if err != nil {
		m.Logger.Printf("could not parse id %s\n", err.Error())
		return
	}
	priority, err := strconv.Atoi(fragments[1])
	if err != nil {
		m.Logger.Printf("could not parse priority %s\n", err.Error())
		return
	}
	sub := &models.Subscription{
		UserID:          uint(id),
		MinimumPriority: models.EventPriority(priority),
		Channel:         "messenger",
		ChannelID:       opts.Sender.ID,
	}
	if err := sub.Add(m.Database); err != nil {
		m.Logger.Printf("could not register subscription %+v, error: %s\n", sub, err.Error())
	}
}
