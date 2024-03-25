package common

import (
	"net"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const DEFAULT_BETS_PER_BATCH = 250

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	BetsPerBatch  int
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
	if config.BetsPerBatch <= 0 {
		config.BetsPerBatch = DEFAULT_BETS_PER_BATCH
	}
	client := &Client{
		config: config,
	}
	client.terminated = false
	return client
}

// _ReadBetsFromCSVFile Reads the bets from a CSV file
func (c *Client) _ReadBetsFromCSVFile(file *CSVFile, number int) ([]*Bet, error) {
	bets := make([]*Bet, 0)
	for i := 0; i < number; i++ {
		tokens, err := file.GetNextLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return make([]*Bet, 0), err
		}
		parsed_dni, _ := strconv.Atoi(tokens["dni"])
		parsed_number, _ := strconv.Atoi(tokens["number"])
		parsed_agency, _ := strconv.Atoi(c.config.ID)
		bettorInfo := NewBettorInfo(tokens["name"], tokens["lastname"], parsed_dni, tokens["birthdate"])
		bet := NewBet(parsed_number, parsed_agency, *bettorInfo)
		bets = append(bets, bet)
	}
	return bets, nil
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
	bets_file_path := os.Getenv("BETS_FILE")
	csv_file := NewCSVFile(bets_file_path)
	iteration_number := 1

loop:
	for timeout := time.After(c.config.LoopLapse); ; iteration_number++ {
		select {
		case <-timeout:
			log.Infof("action: timeout_detected | result: success | client_id: %v",
				c.config.ID,
			)
			break loop
		default:
		}

		// Create the connection the server
		c.createClientSocket()

		// Read the bets from the CSV file
		log.Debugf("action: read_bets | result: in progress | client_id: %v", c.config.ID)

		bets_batch, err := c._ReadBetsFromCSVFile(csv_file, c.config.BetsPerBatch)

		if err != nil {
			if err.Error() != "EOF" {
				log.Errorf("action: read_bets | result: fail | client_id: %v | error: %v",
					c.config.ID, err)
				break loop
			}
		}

		if len(bets_batch) == 0 {
			break loop
		}
		log.Debugf("action: read_bets | result: success | client_id: %v | bets read: %v", c.config.ID, len(bets_batch))

		// Send the bets to the server
		agency_id, _ := strconv.Atoi(c.config.ID)

		log.Debugf("action: send_message | result: in progress | client_id: %v",
			c.config.ID)

		err = SendBets(bets_batch, c.conn, agency_id)

		if err != nil {
			if !c.terminated {
				log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
					c.config.ID, err)
				c.conn.Close()
			}
		}

		// Wait for the confirmation message

		log.Errorf("action: receive_confirmation | result: in progress | client_id: %v",
			c.config.ID)

		err = RecieveConfirmation(c.conn)
		if err == nil {
			log.Infof("action: apuestas_enviadas | result: success | cantidad: %v | batch: %v", len(bets_batch), iteration_number)
		} else {
			if !c.terminated {
				log.Errorf("action: receive_confirmation | result: fail | client_id: %v | error: %v",
					c.config.ID, err)
			}
		}
		c.conn.Close()
	}
	csv_file.Close()
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) Terminate() {
	c.terminated = true
	c.conn.Close()
}
