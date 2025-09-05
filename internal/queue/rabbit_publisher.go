package queue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch                   *amqp.Channel
	exchange, routingKey string
}

func NewPublisher(ch *amqp.Channel, exchange, routingKey string) *Publisher {
	return &Publisher{ch, exchange, routingKey}
}

func (p *Publisher) Publish(ctx context.Context, msg TxMessage) error {
	b, _ := json.Marshal(msg)
	return p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         b,
		DeliveryMode: amqp.Persistent,
	})
}

func (p *Publisher) PublishTransfer(ctx context.Context, msg TransferMessage) error {
	b, _ := json.Marshal(msg)
	return p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         b,
		DeliveryMode: amqp.Persistent,
	})
}
