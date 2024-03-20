package common

import (
	"net"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config     ClientConfig
	conn       net.Conn
	terminated bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	client.terminated = false
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	if c.terminated {
		conn.Close()
	} else {
		c.conn = conn
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// Get the environment variables
	name := os.Getenv("NAME")
	lastname := os.Getenv("LASTNAME")
	dni, _ := strconv.Atoi(os.Getenv("DNI"))
	birthdate := os.Getenv("BIRTHDATE")
	number, _ := strconv.Atoi(os.Getenv("NUMBER"))
	agency, _ := strconv.Atoi(c.config.ID)

	// Create the bet
	bettorInfo := NewBettorInfo(name, lastname, dni, birthdate)
	bet := NewBet(number, int(agency), *bettorInfo)
	// Create the connection the server
	c.createClientSocket()

	err := SendBet(bet, &c.conn)

	if err != nil {
		if !c.terminated {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			c.conn.Close()
		}
	}

	received_confirmation := false

	for !received_confirmation && !c.terminated {
		ack_dni, ack_bet_number, err := RecieveConfirmation(&c.conn)

		if err != nil {
			if !c.terminated && err.Error() != "invalid confirmation message" {
				log.Errorf("action: recieve_confirmation | result: fail | client_id: %v | error: %v",
					c.config.ID, err)
				c.conn.Close()
				return
			}
		}

		if ack_dni == dni && ack_bet_number == number {
			received_confirmation = true
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
				ack_dni, ack_bet_number)
		}
	}

}

func (c *Client) Terminate() {
	c.terminated = true
	c.conn.Close()
}
