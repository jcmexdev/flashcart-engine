package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	URL          string
	ExchangeName string
	ExchangeType string
}

type rabbitPublisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
}

// NewPublisher initializes connection to RabbitMQ, declares topic exchange, and returns an EventPublisher.
func NewPublisher(cfg Config) (ports.EventPublisher, func() error, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, nil, apperror.NewInternal(fmt.Sprintf("failed to connect to rabbitmq at %s", cfg.URL), err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, apperror.NewInternal("failed to open rabbitmq channel", err)
	}

	exchangeName := cfg.ExchangeName
	if exchangeName == "" {
		exchangeName = "order.events"
	}
	exchangeType := cfg.ExchangeType
	if exchangeType == "" {
		exchangeType = "topic"
	}

	err = ch.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, apperror.NewInternal("failed to declare rabbitmq exchange", err)
	}

	cleanup := func() error {
		var firstErr error
		if err := ch.Close(); err != nil {
			firstErr = err
		}
		if err := conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	pub := &rabbitPublisher{
		conn:         conn,
		channel:      ch,
		exchangeName: exchangeName,
	}

	return pub, cleanup, nil
}

// Publish sends the outbox event to RabbitMQ using aggregate routing key pattern: <aggregate_type>.<event_type>
func (p *rabbitPublisher) Publish(ctx context.Context, event *domain.OutboxEvent) error {
	routingKey := fmt.Sprintf("%s.%s", event.AggregateType, event.EventType)

	log.Printf("[RabbitMQ Publisher] Publishing event %s with routing key: %s", event.ID, routingKey)

	err := p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.ID.String(),
			Timestamp:    event.CreatedAt,
			Type:         event.EventType,
			Body:         event.Payload,
		},
	)
	if err != nil {
		return apperror.NewInternal("failed to publish message to rabbitmq", err)
	}

	return nil
}
