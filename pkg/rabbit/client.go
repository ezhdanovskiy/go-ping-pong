package rabbit

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

func NewClient(url string) (*AmqpClient, error) {
	c := &AmqpClient{
		url: url,
	}
	if err := c.init(); err != nil {
		return nil, errors.Wrap(err, "failed to init client")
	}
	return c, nil
}

type AmqpClient struct {
	url  string
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (c *AmqpClient) Ping(id int) error {
	return c.publish(id, DefaultPingsQueueName)
}

func (c *AmqpClient) Pong(id int) error {
	return c.publish(id, DefaultPongsQueueName)
}

func (c *AmqpClient) publish(id int, queueName string) error {
	event := Event{ID: id}
	body, err := json.Marshal(&event)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal event %+v", event)
	}

	err = c.ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to publish event %+v", event)
	}

	return nil
}

func (c *AmqpClient) Pings() (<-chan int, error) {
	chIDs := make(chan int)

	log.Printf("start consuming %s", DefaultPingsQueueName)
	msgs, err := c.ch.Consume(
		DefaultPingsQueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start to consume events")
	}

	go func() {
		for msg := range msgs {
			log.Printf("received: %s", msg.Body)
			var ev Event
			err := json.Unmarshal(msg.Body, &ev)
			if err != nil {
				log.Printf("Failed to unmarshal event: %s", msg.Body)
			}
			chIDs <- ev.ID
		}
		log.Printf("stop consuming %s", DefaultPingsQueueName)
	}()

	return chIDs, nil
}

func (c *AmqpClient) Pongs() (<-chan int, error) {
	chIDs := make(chan int)

	log.Printf("start consuming %s", DefaultPongsQueueName)
	msgs, err := c.ch.Consume(
		DefaultPongsQueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start to consume events")
	}

	go func() {
		for msg := range msgs {
			log.Printf("received: %s", msg.Body)
			var ev Event
			err := json.Unmarshal(msg.Body, &ev)
			if err != nil {
				log.Printf("Failed to unmarshal event: %s", msg.Body)
			}
			chIDs <- ev.ID
		}
		log.Printf("stop consuming %s", DefaultPongsQueueName)
	}()

	return chIDs, nil
}

func (c *AmqpClient) init() error {
	var err error
	log.Printf("dial: %s", c.url)
	c.conn, err = amqp.Dial(c.url)
	if err != nil {
		return errors.Wrap(err, "failed to connect to RabbitMQ")
	}

	log.Printf("get channel")
	c.ch, err = c.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to open a channel")
	}

	log.Printf("declare queue %s", DefaultPingsQueueName)
	_, err = c.ch.QueueDeclare(
		DefaultPingsQueueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to declare queue ping")
	}

	log.Printf("declare queue %s", DefaultPongsQueueName)
	_, err = c.ch.QueueDeclare(
		DefaultPongsQueueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to declare queue ping")
	}

	return nil
}

func (c *AmqpClient) Close() error {
	log.Printf("close channel")
	if c.ch != nil {
		err := c.ch.Close()
		if err != nil {
			return errors.Wrap(err, "failed to close channel")
		}
	}

	if c.conn != nil {
		log.Printf("close connection")
		err := c.conn.Close()
		if err != nil {
			return errors.Wrap(err, "failed to close connection")
		}
	}

	return nil
}
