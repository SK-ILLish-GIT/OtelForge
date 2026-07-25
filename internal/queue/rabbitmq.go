package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName        = "otel.deploy"
	queueName           = "deploy.jobs"
	dlqName             = "deploy.jobs.dlq"
	routingKey          = "job"
	maxDeliveryAttempts = 3
	retryCountHeader    = "x-retry-count"
)

type Queue struct {
	conn *amqp.Connection
}

func Connect(url string) (*Queue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	return &Queue{conn: conn}, nil
}

func (q *Queue) Close() {
	if q.conn != nil {
		_ = q.conn.Close()
	}
}

func (q *Queue) newChannel() (*amqp.Channel, error) {
	return q.conn.Channel()
}

func (q *Queue) Declare(ctx context.Context) error {
	ch, err := q.newChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(dlqName, true, false, false, false, nil); err != nil {
		return err
	}
	args := amqp.Table{
		"x-dead-letter-exchange":    exchangeName,
		"x-dead-letter-routing-key": "dlq",
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, args); err != nil {
		return err
	}
	if err := ch.QueueBind(queueName, routingKey, exchangeName, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(dlqName, "dlq", exchangeName, false, nil); err != nil {
		return err
	}
	return nil
}

func (q *Queue) Publish(ctx context.Context, msg JobMessage) error {
	ch, err := q.newChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}

func (q *Queue) Consume(ctx context.Context, handler func(JobMessage) error) error {
	ch, err := q.newChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			var msg JobMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("invalid message: %v", err)
				_ = d.Nack(false, false)
				continue
			}
			if err := handler(msg); err != nil {
				log.Printf("job handler error: %v", err)
				retry := deliveryRetryCount(d.Headers)
				if retry+1 >= maxDeliveryAttempts {
					log.Printf("job exceeded %d attempts, sending to DLQ", maxDeliveryAttempts)
					_ = d.Nack(false, false)
					continue
				}
				headers := amqp.Table{}
				for k, v := range d.Headers {
					headers[k] = v
				}
				headers[retryCountHeader] = int32(retry + 1)
				if err := ch.PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{
					ContentType:  "application/json",
					Body:         d.Body,
					Headers:      headers,
					DeliveryMode: amqp.Persistent,
				}); err != nil {
					log.Printf("republish for retry failed: %v", err)
					_ = d.Nack(false, false)
					continue
				}
				_ = d.Ack(false)
				continue
			}
			_ = d.Ack(false)
		}
	}
}
