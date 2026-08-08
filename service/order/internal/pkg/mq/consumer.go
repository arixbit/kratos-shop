package mq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	MaxRetry       = 5
	DLXExchange    = "mall.dlx"
	RetryHeaderKey = "x-retry-count"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	handler func(context.Context, []byte) error
}

func NewConsumer(addr, exchange, queue string, routingKeys []string, handler func(context.Context, []byte) error) (*Consumer, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(DLXExchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	for _, key := range routingKeys {
		if err := ch.QueueBind(q.Name, key, exchange, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, err
		}
	}
	return &Consumer{conn: conn, channel: ch, queue: q.Name, handler: handler}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := c.handler(ctx, d.Body); err != nil {
				c.retryOrDLQ(d, err)
				continue
			}
			_ = d.Ack(false)
		}
	}
}

func (c *Consumer) retryOrDLQ(d amqp.Delivery, err error) {
	attempts := deliveryRetryCount(d.Headers)
	if attempts+1 >= MaxRetry {
		_ = c.channel.Publish(
			DLXExchange,
			"failed."+c.queue,
			false,
			false,
			amqp.Publishing{
				ContentType:  d.ContentType,
				DeliveryMode: amqp.Persistent,
				Headers:      d.Headers,
				Body:         d.Body,
			},
		)
		_ = d.Ack(false)
		return
	}

	headers := make(amqp.Table, len(d.Headers)+1)
	for k, v := range d.Headers {
		headers[k] = v
	}
	headers[RetryHeaderKey] = attempts + 1
	time.Sleep(time.Duration(attempts+1) * time.Second)
	_ = c.channel.Publish(
		"",
		c.queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  d.ContentType,
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
			Body:         d.Body,
		},
	)
	_ = d.Ack(false)
}

func deliveryRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if v, ok := headers[RetryHeaderKey]; ok {
		switch n := v.(type) {
		case int32:
			return int(n)
		case int64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}

func (c *Consumer) Publish(exchange, routingKey string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.channel.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (c *Consumer) Close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
