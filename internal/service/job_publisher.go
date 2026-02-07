package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

// TranscodingJobMessage represents a transcoding job
type TranscodingJobMessage struct {
	FileID       uint64 `json:"file_id"`
	InputPath    string `json:"input_path"`
	OutputPath   string `json:"output_path"`
	StorageType  string `json:"storage_type"`
	AttemptCount int    `json:"attempt_count"`
}

// JobPublisher handles publishing jobs to message queue
type JobPublisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

// NewJobPublisher creates a new job publisher
func NewJobPublisher(rabbitMQURL, queueName string) (*JobPublisher, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &JobPublisher{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

// PublishTranscodingJob publishes a transcoding job to the queue
func (p *JobPublisher) PublishTranscodingJob(ctx context.Context, job *TranscodingJobMessage) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	err = p.channel.Publish(
		"",          // exchange
		p.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close closes the connection
func (p *JobPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
