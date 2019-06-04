package app

import (
	"github.com/streadway/amqp"
	"log"
)

type AmqpClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func (m *AmqpClient) Connect() {
	connection, err := amqp.Dial("amqp://guest:guest@home.fritz.box:5672/")
	if err != nil {
		panic(err)
	}

	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}
	m.connection = connection
	m.channel = channel
}

func (m *AmqpClient) Disconnect() {
	m.channel.Close()
	m.connection.Close()
}

func (m *AmqpClient) QueueDeclare(name string) amqp.Queue {
	q, _ := m.channel.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	return q
}

func (m *AmqpClient) Publish(queue amqp.Queue, body []byte) {
	err := m.channel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		log.Println(err)
	}
}
