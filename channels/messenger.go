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
	m.messenger.MessageReceived = m.Received

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
			template.NewWebURLButton("Zobacz więcej", fmt.Sprintf("https://margherita.xememah.com/#/event/%d/messenger/", event.ID)),
		},
	})
	_, err := m.messenger.SendMessage(mq)
	return err
}

func (m *Messenger) Received(_ messenger.Event, opt messenger.MessageOpts, _ messenger.ReceivedMessage) {
	fmt.Println(opt.Sender.ID)
}
