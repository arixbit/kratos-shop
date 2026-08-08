package mq

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	addr string

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(addr string) *Publisher {
	return &Publisher{addr: addr}
}

func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.ensure(); err != nil {
		return err
	}
	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return p.ch.PublishWithContext(pubCtx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (p *Publisher) ensure() error {
	if p.conn != nil && !p.conn.IsClosed() && p.ch != nil {
		return nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	conn, err := amqp.Dial(p.addr)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	if err := ch.ExchangeDeclare("payment.exchange", "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return err
	}
	p.conn = conn
	p.ch = ch
	return nil
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}
