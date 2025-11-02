package natsbroker

import (
	"encoding/json"

	"github.com/nats-io/nats.go"	
	"github.com/balaji-balu/margo-hello-world/internal/gitobserver"
)

type Broker struct {
    conn *nats.Conn
}

func New(url string) (*Broker, error) {
    nc, err := nats.Connect(url)
    if err != nil {
        return nil, err
    }
    return &Broker{conn: nc}, nil
}

func (b *Broker) Publish(topic string, msg interface{}) error {
    data, _ := json.Marshal(msg)
    return b.conn.Publish(topic, data)
}

func (b *Broker) Subscribe(topic string, handler func(gitobserver.GitEvent)) error {
    _, err := b.conn.Subscribe(topic, func(m *nats.Msg) {
        var ev gitobserver.GitEvent
        _ = json.Unmarshal(m.Data, &ev)
        handler(ev)
    })
    return err
}

func (b *Broker) Close() {
    if b.conn != nil {
        b.conn.Close()
    }
}
