package queue

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeliveryRetryCount(t *testing.T) {
	if got := deliveryRetryCount(nil); got != 0 {
		t.Fatalf("nil headers: got %d", got)
	}
	if got := deliveryRetryCount(amqp.Table{retryCountHeader: int32(2)}); got != 2 {
		t.Fatalf("int32: got %d", got)
	}
}
