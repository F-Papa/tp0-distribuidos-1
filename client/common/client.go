package common

import (
	"net"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const DEFAULT_BETS_PER_BATCH = 250

const SEND_BETS_PHASE = 0
const CONSULT_WINNERS_PHASE = 1
const ANNOUNCE_WINNERS_PHASE = 2

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
	phase      int
	winners    []int
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	if config.BetsPerBatch <= 0 {
		log.Warnf("Invalid bets per batch. Using default value: %v", DEFAULT_BETS_PER_BATCH)
		config.BetsPerBatch = DEFAULT_BETS_PER_BATCH
	}
	client := &Client{
		config:  config,
		phase:   SEND_BETS_PHASE,
		winners: make([]int, 0),
	}
	client.terminated = false
	return client
}

// _NextPhase Changes the phase of the client to the next one
func (c *Client) _NextPhase() {
	switch c.phase {
	case SEND_BETS_PHASE:
		c.phase = CONSULT_WINNERS_PHASE
	case CONSULT_WINNERS_PHASE:
		c.phase = ANNOUNCE_WINNERS_PHASE
	}
}

// SetWinners Sets the winners of the client
func (c *Client) SetWinners(winners []int) {
	c.winners = winners
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
	// Create the connection the server
	err := c.createClientSocket()
	if err != nil {
		if !c.terminated {
			log.Errorf("action: create_client_socket | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
		}
		return
	}
	agency_id_int, _ := strconv.Atoi(c.config.ID)
	err = SendConnectMessage(c.conn, agency_id_int)
	if err != nil {
		if !c.terminated {
			log.Errorf("action: send_connect | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
		}
		return
	}

	bets_file_path := os.Getenv("BETS_FILE")
	csv_file := NewCSVFile(bets_file_path)
	defer csv_file.Close()

loop:
	for timeout := time.After(c.config.LoopLapse); !c.terminated; {
		select {
		case <-timeout:
			log.Infof("action: timeout_detected | result: success | client_id: %v",
				c.config.ID,
			)
			break loop
		default:
		}

		switch c.phase {
		case SEND_BETS_PHASE:
			err = c.SendBetsPhase(csv_file)
		case CONSULT_WINNERS_PHASE:
			err = c.ConsultWinnersPhase()
			if c.phase == ANNOUNCE_WINNERS_PHASE {
				break loop
			}
		case ANNOUNCE_WINNERS_PHASE:
			break loop
		}
		if err != nil {
			return
		}
	}

	if c.terminated {
		log.Infof("action: terminate client | result: success | client_id: %v", c.config.ID)
		return
	}
	log.Infof("action: consulta_ganadores | result: success | client_id: %v | cant_ganadores: %v",
		agency_id_int, len(c.winners))
}

// Handles the sending of bets to the server and advances to the next phase
// if all bets have been sent
func (c *Client) SendBetsPhase(bets_file *CSVFile) error {
	agency_id_int, _ := strconv.Atoi(c.config.ID)
	bets_batch, err := bets_file.ReadBetsFromCSVFile(c.config.BetsPerBatch, agency_id_int)
	if err != nil && err.Error() != "EOF" {
		log.Errorf("action: read_bets | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	if len(bets_batch) == 0 {
		// All bets have been read and sent
		err := SendFinishedMessage(c.conn, agency_id_int)
		if err != nil && !c.terminated {
			log.Errorf("action: send finished | result: fail | client_id: %v | error: %v",
				agency_id_int, err)
			return err
		}
		c._NextPhase()
		return nil
	}

	log.Debugf("action: read_bets | result: success | client_id: %v | bets read: %v", c.config.ID, len(bets_batch))
	err = SendBets(bets_batch, c.conn, agency_id_int)
	if err != nil && !c.terminated {
		log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v",
			agency_id_int, err)
		return err
	}

	err = RecieveBatchConfirmation(c.conn)
	if err != nil && !c.terminated {
		log.Errorf("action: batch confirmation | result: fail | client_id: %v | error: %v",
			agency_id_int, err)
		return err
	}
	log.Infof("action: batch confirmation | result: success | client_id: %v", agency_id_int)
	return nil
}

// Handles the receiving of winners from the server during the second phase and advances to the next phase
// if the winners are received
func (c *Client) ConsultWinnersPhase() error {
	agency_id_int, _ := strconv.Atoi(c.config.ID)

	err := ConsultResults(c.conn, agency_id_int)
	if err != nil {
		if !c.terminated {
			log.Errorf("action: consult winners | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
		}
		return err
	}

	winners, wait, err := ReceiveResults(c.conn)
	if err != nil {
		if !c.terminated {
			log.Errorf("action: receive winners | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
		}
		return err
	}

	if wait {
		time.Sleep(time.Second * 2)
	} else {
		c.SetWinners(winners)
		c._NextPhase()
	}

	return nil
}

// Terminate Closes the connection and sets the client as terminated
func (c *Client) Terminate() {
	c.terminated = true
	c.conn.Close()
}
