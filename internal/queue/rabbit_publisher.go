package queue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher encapsulates an AMQP channel along with the exchange and routing key
// used for publishing messages to a RabbitMQ broker.
type Publisher struct {
	ch                   *amqp.Channel
	exchange, routingKey string
}

// NewPublisher creates and returns a new Publisher instance using the provided
// AMQP channel, exchange name, and routing key. The Publisher can be used to
// publish messages to the specified exchange with the given routing key.
//
// Parameters:
//
//	ch         - Pointer to an AMQP channel used for publishing messages.
//	exchange   - Name of the exchange to publish messages to.
//	routingKey - Routing key to use when publishing messages.
//
// Returns:
//
//	*Publisher - A pointer to the newly created Publisher.
func NewPublisher(ch *amqp.Channel, exchange, routingKey string) *Publisher {
	return &Publisher{ch, exchange, routingKey}
}

// Publish sends a TxMessage to the configured RabbitMQ exchange and routing key.
// The message is marshaled to JSON and published with persistent delivery mode.
// It uses the provided context for cancellation and timeout control.
// Returns an error if the message could not be published.
func (p *Publisher) Publish(ctx context.Context, msg TxMessage) error {
	b, _ := json.Marshal(msg)
	return p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         b,
		DeliveryMode: amqp.Persistent,
	})
}

// PublishTransfer publishes a TransferMessage to the configured RabbitMQ exchange and routing key.
// The message is marshaled to JSON and sent with persistent delivery mode.
// Returns an error if publishing fails.
//
// Parameters:
//   - ctx: The context for controlling cancellation and deadlines.
//   - msg: The TransferMessage to be published.
//
// Returns:
//   - error: Non-nil if the message could not be published.
func (p *Publisher) PublishTransfer(ctx context.Context, msg TransferMessage) error {
	b, _ := json.Marshal(msg)
	return p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         b,
		DeliveryMode: amqp.Persistent,
	})
}
