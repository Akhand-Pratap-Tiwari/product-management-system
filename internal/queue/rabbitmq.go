package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"product-management-system/internal/models"

	"github.com/streadway/amqp"
)

type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQQueue(host string, port int) *RabbitMQQueue {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%d", host, port))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		"image_processing_queue", // name
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	return &RabbitMQQueue{
		conn:    conn,
		channel: ch,
		queue:   q,
	}
}

func (r *RabbitMQQueue) EnqueueImageProcessing(task *models.ImageProcessingTask) error {
	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = r.channel.Publish(
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQQueue) Close() {
	r.channel.Close()
	r.conn.Close()
}
