package messagebus

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type MQConnection struct {
	m               sync.Mutex
	host            string
	port            string
	user            string
	password        string
	queueName       string
	vHost           string
	disableSSL      bool
	connAttempts    uint64
	connection      *amqp.Connection
	channel         *amqp.Channel
	UnexpectedClose chan struct{}
	wharfInstanceID string
}

const (
	reconnectDelay = 3 * time.Second
)

func NewConnection(host, port, user, pass, qName, vHost string, dssl bool, connAttempts uint64) (*MQConnection, error) {
	instance, ok := os.LookupEnv("WHARF_INSTANCE")
	if !ok {
		return nil, fmt.Errorf("WHARF_INSTANCE environment variable required but not set")
	}
	return &MQConnection{
		host:            host,
		port:            port,
		user:            user,
		password:        pass,
		queueName:       qName,
		vHost:           vHost,
		disableSSL:      dssl,
		connAttempts:    connAttempts,
		wharfInstanceID: instance,
		UnexpectedClose: make(chan struct{}),
	}, nil
}

func (conn *MQConnection) Connect() error {
	if conn == nil {
		return errors.New("Mq connection not initialized")
	}

	if err := conn.connectToRabbitMQ(); err != nil {
		return fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}

	if err := conn.connectToChannel(); err != nil {
		if err2 := conn.CloseConnection(); err2 != nil {
			log.Printf("Failed to close connection: %v\n", err2)
			return fmt.Errorf("Failed to open a channel: %v and close connection: %v", err, err2)
		}
		return fmt.Errorf("Failed to connect to channel: %v", err)
	}

	return nil
}

func (conn *MQConnection) createAmqpURL() string {
	uri := fmt.Sprintf("%s:%s@%s:%s/",
		conn.user,
		conn.password,
		conn.host,
		conn.port)

	if conn.vHost != "" {
		uri = fmt.Sprintf("%s%s", uri, conn.vHost)
	}

	var prefix string
	if conn.disableSSL {
		prefix = "amqp://"
	} else {
		prefix = "amqps://"
	}

	return fmt.Sprintf("%s%s", prefix, uri)
}

func (conn *MQConnection) connectToRabbitMQ() error {
	url := conn.createAmqpURL()
	var err error

	for i := uint64(0); i < conn.connAttempts; i++ {
		conn.m.Lock()
		log.Printf("%d. Trying to connect to RabbitMQ at %s\n", i, url)

		if conn.disableSSL {
			conn.connection, err = amqp.Dial(url)
		} else {
			tlsConfig := &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
			conn.connection, err = amqp.DialTLS(url, tlsConfig)
		}

		conn.m.Unlock()

		if err == nil {
			go conn.handleConnectionClosed()

			log.Printf("Connected to RabbitMQ %s\n", url)
			return nil
		}

		log.Println(err)
		time.Sleep(reconnectDelay)
	}

	conn.connection = nil
	return err
}

func (conn *MQConnection) connectToChannel() error {
	if conn.connection == nil {
		return fmt.Errorf("missing connection")
	}

	var err error
	if conn.channel, err = conn.connection.Channel(); err != nil {
		conn.channel = nil
		return fmt.Errorf("Failed to open a channel: %v", err)
	}

	go conn.handleChannelClosed()

	_, err = conn.channel.QueueDeclare(
		conn.queueName, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Printf("Failed to declare a queue: %v\n", err)
		if err2 := conn.CloseConnection(); err2 != nil {
			log.Printf("Failed to close connection: %v\n", err2)
			return fmt.Errorf("Failed to declare a queue: %v and close connection: %v", err, err2)
		}
		return fmt.Errorf("Failed to declare a queue: %v", err)
	}

	return nil
}

func (conn *MQConnection) handleConnectionClosed() {
	var closeError = make(chan *amqp.Error)
	reason, ok := <-conn.connection.NotifyClose(closeError)

	conn.m.Lock()
	log.Println("connection closed")

	conn.connection = nil

	conn.m.Unlock()

	// exit this goroutine if closed by developer
	if !ok {
		close(conn.UnexpectedClose)
		return
	}
	log.Printf("connection close reason: %v\n", reason)

	if err := conn.connectToRabbitMQ(); err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v\n", err)
		conn.UnexpectedClose <- struct{}{}
		return
	}

	conn.m.Lock()
	if err := conn.connectToChannel(); err != nil {
		log.Printf("Failed to open a channel: %v\n", err)
		conn.connection.Close()
		conn.UnexpectedClose <- struct{}{}
	}
	conn.m.Unlock()
}

func (conn *MQConnection) handleChannelClosed() {
	var closeChannelError = make(chan *amqp.Error)
	reason, ok := <-conn.channel.NotifyClose(closeChannelError)

	conn.m.Lock()
	defer conn.m.Unlock()

	log.Println("channel closed")

	conn.channel = nil

	// exit this goroutine if closed by developer
	if !ok {
		return
	}

	log.Printf("channel close reason: %v\n", reason)

	err := conn.connectToChannel()
	if err != nil {
		log.Printf("Unable to connect to channel: %v\n", err)
	}
}

func (conn *MQConnection) CloseConnection() error {
	if conn == nil {
		return errors.New("Mq connection not initialized")
	}

	if conn.channel != nil {
		log.Printf("Closing message queue channel to %s:%s, queue %s\n", conn.host, conn.port, conn.queueName)
		err := conn.channel.Close()
		if err != nil {
			return err
		}
	}

	if conn.connection != nil && !conn.connection.IsClosed() {
		log.Printf("Closing message queue connection to %s:%s, queue %s\n", conn.host, conn.port, conn.queueName)
		err := conn.connection.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (conn *MQConnection) PublishMessage(message interface{}) error {
	if conn == nil {
		return errors.New("Mq connection not initialized")
	}

	jsonBody, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return conn.publish(string(jsonBody))
}

func (conn *MQConnection) publish(message string) error {
	conn.m.Lock()
	defer conn.m.Unlock()

	if conn.connection == nil {
		errStr := "Failed to publish a message, connection is missing"
		log.Printf(errStr)
		return errors.New(errStr)
	}

	if conn.connection.IsClosed() {
		errStr := "Failed to publish a message, connection is closed"
		log.Printf(errStr)
		return errors.New(errStr)
	}

	headers := amqp.Table{
		"WharfInstanceId": conn.wharfInstanceID,
		"Timestamp":       time.Now().UTC(),
	}

	err := conn.channel.Publish(
		"",             // exchange
		conn.queueName, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        []byte(message),
		})

	if err != nil {
		log.Printf("Failed to publish a message: %s", err.Error())
		return err
	}

	return nil
}
