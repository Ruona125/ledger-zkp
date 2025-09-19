package events

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Bus struct{ nc *nats.Conn }

func NewBus(url string) (*Bus, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Bus{nc: nc}, nil
}

func (b *Bus) Publish(subject string, v any) error {
	data, _ := json.Marshal(v)
	return b.nc.Publish(subject, data)
}

func (b *Bus) Subscribe(subject string, fn func([]byte)) (*nats.Subscription, error) {
	return b.nc.Subscribe(subject, func(m *nats.Msg) { fn(m.Data) })
}
