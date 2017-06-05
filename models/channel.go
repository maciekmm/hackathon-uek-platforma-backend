package models

type ChannelType string

const (
	Messenger ChannelType = "messenger"
)

type Channel interface {
	Send(Event, Subscription) error
}
