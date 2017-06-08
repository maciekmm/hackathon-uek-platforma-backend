package channels

import (
	"errors"
	"log"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/models"
)

const (
	bufferChannelCapacity = 1 << 10
)

var (
	ErrQueueFull = errors.New("queue is full")
)

type Channel interface {
	Type() models.ChannelType
	Register(*mux.Router)
	Send(*models.Subscription, *models.Event) error
}

type Coordinator struct {
	channels      map[models.ChannelType]Channel
	database      *gorm.DB
	bufferChannel chan *models.Event
	logger        *log.Logger
}

func NewCoordinator(logger *log.Logger, database *gorm.DB, channels ...Channel) *Coordinator {
	cc := &Coordinator{channels: make(map[models.ChannelType]Channel), logger: logger, database: database, bufferChannel: make(chan *models.Event, bufferChannelCapacity)}
	for _, channel := range channels {
		cc.channels[channel.Type()] = channel
	}
	return cc
}

func (c *Coordinator) Start() {
	for {
		msg, ok := <-c.bufferChannel
		if !ok {
			c.logger.Println("stopping channel coordinator")
			break
		}
		subs, err := c.subscriptions(msg)
		if err != nil {
			c.logger.Printf("could not fetch subscriptions for %+v, error: %s\n", msg, err.Error())
			continue
		}

		for _, sub := range subs {
			if ch, ok := c.channels[sub.Channel]; ok {
				go ch.Send(sub, msg)
			}
		}
	}
}

func (c *Coordinator) subscriptions(event *models.Event) ([]*models.Subscription, error) {
	subscriptions := []*models.Subscription{}
	res := c.database.Table("subscriptions").Select("subscriptions.*").Joins("right join users ON subscriptions.user_id=users.id").Where("minimum_priority >= ?", event.Priority)
	if event.Group != nil {
		res = res.Where("\"users\".\"group\" = ?", *event.Group)
	}
	res = res.Find(&subscriptions)
	if res.Error != nil {
		return subscriptions, res.Error
	}
	return subscriptions, nil
}

func (c *Coordinator) Send(event *models.Event) error {
	select {
	case c.bufferChannel <- event:
		return nil
	default:
		return ErrQueueFull
	}
}

func (c *Coordinator) Stop() {
	close(c.bufferChannel)
}
