package mq

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeliveryRetryCount(t *testing.T) {
	if got := deliveryRetryCount(nil); got != 0 {
		t.Fatalf("nil headers should return 0, got %d", got)
	}
	if got := deliveryRetryCount(amqp.Table{}); got != 0 {
		t.Fatalf("empty headers should return 0, got %d", got)
	}
	if got := deliveryRetryCount(amqp.Table{RetryHeaderKey: int32(3)}); got != 3 {
		t.Fatalf("int32 header should return 3, got %d", got)
	}
	if got := deliveryRetryCount(amqp.Table{RetryHeaderKey: int64(4)}); got != 4 {
		t.Fatalf("int64 header should return 4, got %d", got)
	}
	if got := deliveryRetryCount(amqp.Table{RetryHeaderKey: 5}); got != 5 {
		t.Fatalf("int header should return 5, got %d", got)
	}
}
