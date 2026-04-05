package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
}

func newClient(cfg Config) (Bus, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("rabbitmq is not enabled")
	}

	vhost := cfg.VHost
	if strings.TrimSpace(vhost) == "" {
		vhost = "/"
	}

	urlString := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		url.QueryEscape(cfg.Username),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		url.PathEscape(vhost),
	)

	conn, err := amqp.Dial(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	c := &client{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}

	if err := c.channel.ExchangeDeclare(
		c.config.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return c, nil
}

func (c *client) Publish(ctx context.Context, event string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		c.config.Exchange,
		event,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (c *client) Subscribe(event string, handler func(context.Context, []byte) error) error {
	queueName := fmt.Sprintf("%s.%s.queue", c.config.ServiceName, event)

	q, err := c.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = c.channel.QueueBind(
		q.Name,
		event,
		c.config.Exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := c.channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			err := handler(context.Background(), msg.Body)
			if err != nil {
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}()

	return nil
}

func (c *client) Close() error {
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}
