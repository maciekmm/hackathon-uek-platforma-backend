package channels

import (
	"fmt"
	"os"

	"github.com/gorilla/mux"
	messenger "github.com/maciekmm/messenger-platform-go-sdk"
	"github.com/maciekmm/messenger-platform-go-sdk/template"
	"github.com/maciekmm/uek-bruschetta/models"
)

const (
	ChannelTypeMessenger models.ChannelType = "messenger"
)

type Messenger struct {
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
			template.NewWebURLButton("Zobacz więcej", "https://google.com"),
			template.NewWebURLButton("JP2", "https://en.wikipedia.org/wiki/Pope_John_Paul_II"),
		},
	})
	_, err := m.messenger.SendMessage(mq)
	return err
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
