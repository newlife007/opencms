package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQQueue implements QueueService using RabbitMQ
type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQQueue creates a new RabbitMQ queue service
func NewRabbitMQQueue(url string) (*RabbitMQQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	
	return &RabbitMQQueue{
		conn:    conn,
		channel: channel,
		url:     url,
	}, nil
}

// DeclareQueue declares a queue with dead letter exchange for retry/DLQ
func (q *RabbitMQQueue) DeclareQueue(queueName string) error {
	// Declare dead letter exchange
	if err := q.channel.ExchangeDeclare(
		"dlx",      // name
		"direct",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	); err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	// Declare dead letter queue
	dlqName := queueName + "_dlq"
	if _, err := q.channel.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	); err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Bind dead letter queue to exchange
	if err := q.channel.QueueBind(
		dlqName, // queue name
		dlqName, // routing key
		"dlx",   // exchange
		false,   // no-wait
		nil,     // arguments
	); err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// Declare main queue with DLX configuration
	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx",
		"x-dead-letter-routing-key": dlqName,
		"x-message-ttl":             86400000, // 24 hours
	}

	_, err := q.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		args,  // arguments with DLX
	)
	
	return err
}

// Publish publishes a message to a queue
func (q *RabbitMQQueue) Publish(ctx context.Context, queueName string, message *Message) error {
	// Declare queue with DLX
	if err := q.DeclareQueue(queueName); err != nil {
		return err
	}
	
	// Publish message
	err := q.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message.Body),
			MessageId:    message.ID,
			Timestamp:    time.Now(),
		},
	)
	
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	log.Printf("Published message to queue %s: %s", queueName, message.ID)
	return nil
}

// Subscribe subscribes to a queue and processes messages with retry logic
func (q *RabbitMQQueue) Subscribe(ctx context.Context, queueName string, handler func(*Message) error) error {
	// Declare queue with DLX
	if err := q.DeclareQueue(queueName); err != nil {
		return err
	}
	
	// Set prefetch count for fair dispatch
	if err := q.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	
	// Start consuming
	msgs, err := q.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}
	
	log.Printf("Started consuming from queue: %s", queueName)
	
	// Process messages
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer")
			return ctx.Err()
			
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Channel closed")
				return fmt.Errorf("message channel closed")
			}
			
			message := &Message{
				ID:   msg.MessageId,
				Body: string(msg.Body),
			}
			
			// Get retry count from headers
			retryCount := 0
			if msg.Headers != nil {
				if val, ok := msg.Headers["x-retry-count"].(int32); ok {
					retryCount = int(val)
				}
			}
			
			maxRetries := 3
			
			// Process message
			if err := handler(message); err != nil {
				log.Printf("Failed to process message %s (attempt %d/%d): %v", message.ID, retryCount+1, maxRetries, err)
				
				// Check if we should retry
				if retryCount < maxRetries {
					// Nack and requeue with delay
					if err := msg.Nack(false, false); err != nil { // Don't requeue immediately
						log.Printf("Failed to nack message: %v", err)
					}
					
					// Republish with incremented retry count
					go q.republishWithDelay(queueName, message, retryCount+1)
				} else {
					// Max retries exceeded, send to DLQ
					log.Printf("Max retries exceeded for message %s, sending to DLQ", message.ID)
					if err := msg.Reject(false); err != nil { // Reject without requeue (goes to DLQ)
						log.Printf("Failed to reject message: %v", err)
					}
				}
			} else {
				// Acknowledge successful processing
				if err := msg.Ack(false); err != nil {
					log.Printf("Failed to ack message: %v", err)
				}
				log.Printf("Successfully processed message: %s", message.ID)
			}
		}
	}
}

// republishWithDelay republishes a message with exponential backoff
func (q *RabbitMQQueue) republishWithDelay(queueName string, message *Message, retryCount int) {
	// Calculate delay: 1s, 5s, 15s
	delays := []time.Duration{1 * time.Second, 5 * time.Second, 15 * time.Second}
	delay := delays[0]
	if retryCount-1 < len(delays) {
		delay = delays[retryCount-1]
	}
	
	log.Printf("Retrying message %s after %v (retry %d)", message.ID, delay, retryCount)
	time.Sleep(delay)
	
	// Republish with retry count in headers
	ctx := context.Background()
	headers := amqp.Table{
		"x-retry-count": int32(retryCount),
	}
	
	err := q.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message.Body),
			MessageId:    message.ID,
			Headers:      headers,
			Timestamp:    time.Now(),
		},
	)
	
	if err != nil {
		log.Printf("Failed to republish message: %v", err)
	}
}

// HealthCheck checks if RabbitMQ is healthy
func (q *RabbitMQQueue) HealthCheck(ctx context.Context) error {
	if q.conn == nil || q.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}

	if q.channel == nil {
		return fmt.Errorf("rabbitmq channel is nil")
	}

	// Try to declare a test queue
	testQueue := "_healthcheck"
	if _, err := q.channel.QueueDeclare(
		testQueue,
		false, // durable
		true,  // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("rabbitmq health check failed: %w", err)
	}

	// Delete test queue
	if _, err := q.channel.QueueDelete(testQueue, false, false, false); err != nil {
		log.Printf("Warning: failed to delete health check queue: %v", err)
	}

	return nil
}

// GetQueueStats returns statistics for a queue
func (q *RabbitMQQueue) GetQueueStats(queueName string) (int, int, error) {
	queue, err := q.channel.QueueInspect(queueName)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return queue.Messages, queue.Consumers, nil
}

// Close closes the RabbitMQ connection
func (q *RabbitMQQueue) Close() error {
	if q.channel != nil {
		q.channel.Close()
	}
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}
