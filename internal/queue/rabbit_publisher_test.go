package queue_test

import (
	"context"
	"testing"

	"github.com/Bharat0908/ledger/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestNewPublisher(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		ch         *amqp.Channel
		exchange   string
		routingKey string
		want       *queue.Publisher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := queue.NewPublisher(tt.ch, tt.exchange, tt.routingKey)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewPublisher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublisher_Publish(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		ch         *amqp.Channel
		exchange   string
		routingKey string
		// Named input parameters for target function.
		msg     queue.TxMessage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := queue.NewPublisher(tt.ch, tt.exchange, tt.routingKey)
			gotErr := p.Publish(context.Background(), tt.msg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Publish() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Publish() succeeded unexpectedly")
			}
		})
	}
}

func TestPublisher_PublishTransfer(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		ch         *amqp.Channel
		exchange   string
		routingKey string
		// Named input parameters for target function.
		msg     queue.TransferMessage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := queue.NewPublisher(tt.ch, tt.exchange, tt.routingKey)
			gotErr := p.PublishTransfer(context.Background(), tt.msg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("PublishTransfer() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("PublishTransfer() succeeded unexpectedly")
			}
		})
	}
}
