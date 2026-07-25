package queue

import amqp "github.com/rabbitmq/amqp091-go"

func deliveryRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	v, ok := headers[retryCountHeader]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int32:
		return int(n)
	case int64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
