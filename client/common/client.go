package common

import (
	"bufio"
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
	dni := os.Getenv("DNI")
	birthdate := os.Getenv("BIRTHDATE")
	number, _ := strconv.Atoi(os.Getenv("NUMBER"))
	agency, _ := strconv.Atoi(c.config.ID)

	// Create the bet
	bettorInfo := NewBettorInfo(name, lastname, dni, birthdate)
	bet := NewBet(number, int(agency), *bettorInfo)
	serializedBet := SerializeBet(bet)

	log.Debugf("Serialized Bet: %x", serializedBet)
	log.Debugf("Len: %d", len(serializedBet))

	// Create the connection the server
	c.createClientSocket()

	//Sleep for one minute
	total_bytes_written := 0

	for total_bytes_written < len(serializedBet) && !c.terminated {
		bytes_written, err := c.conn.Write(serializedBet[total_bytes_written:])

		if err != nil {
			if !c.terminated {
				log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
					c.config.ID, err)
				c.conn.Close()
			}
			return
		}
		total_bytes_written += bytes_written
	}

	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil {
		if c.terminated {
			log.Infof("action: terminate | result: success | client_id: %v",
				c.config.ID)
		} else {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
		}
		return
	}

	switch msg {
	case "1\n":
		log.Infof("action: store_bet | result: success | dni: %v | number: %v",
			bet.bettor.dni, bet.number)
	default:
		log.Errorf("action: store_bet | result: fail | dni: %v | number: %v",
			bet.bettor.dni, bet.number)
	}

}

func (c *Client) Terminate() {
	c.terminated = true
	c.conn.Close()
}
