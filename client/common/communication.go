package common

import (
	"bufio"
	"fmt"
	"net"
)

// Constants for the communication protocol
const SIZE_FIELD_LENGTH = 2      // Size of the length field in bytes
const MSG_CODE_LENGTH = 1        // Size of the type field in bytes
const NUMBER_LENGTH_IN_BYTES = 2 // Size of the number field in bytes
const AGENCY_LENGTH_IN_BYTES = 1 // Size of the agency field in bytes
const DNI_LENGTH_IN_BYTES = 4    // Size of the DNI field in bytes
const YEAR_LENGTH_IN_BYTES = 2   // Size of the year field in bytes
const MONTH_LENGTH_IN_BYTES = 1  // Size of the month field in bytes
const DAY_LENGTH_IN_BYTES = 1    // Size of the day field in bytes

// Client Codes
const CONNECT_CODE = 10  // The code the client uses to connect to the server
const BET_MSG_CODE = 14  // The code the client uses to send a bet
const FINISHED_CODE = 20 // The code the client uses to end betting
const CONSULT_CODE = 23  // The code the client uses to request the results

// Server Codes
const CONFIRMATION_CODE = 21 // The code the server uses to confirm a bet batch
const RESULTS_MSG_CODE = 22  // The code the server uses to send the results
const WAIT_MSG_CODE = 25     // The code the server uses to tell the client to wait

// Sends bets to the server and returns an error if any.
func SendBets(bets []*Bet, conn net.Conn, agency_id int) error {

	total_bets_sent := 0
	for total_bets_sent < len(bets) {
		buffer := make([]byte, 0)
		bets_that_fit := _SerializeBets(bets, &buffer)
		total_bets_sent += bets_that_fit
		err := _SendAux(buffer, conn, agency_id, BET_MSG_CODE)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sends a message to the server indicating that the client has finished sending bets.
func SendFinishedMessage(conn net.Conn, agency_id int) error {
	return _SendAux([]byte{}, conn, agency_id, FINISHED_CODE)
}

// Sends a message to the server indicating that the client has finished sending bets.
func SendConnectMessage(conn net.Conn, agency_id int) error {
	return _SendAux([]byte{}, conn, agency_id, CONNECT_CODE)
}

// Sends a message to the server requesting the results of the lottery.
func ConsultResults(conn net.Conn, agency_id int) error {
	return _SendAux([]byte{}, conn, agency_id, CONSULT_CODE)
}

// Receives the results from the server and returns the winners, whether the server told the client to wait,
// and an error if any.
func ReceiveResults(conn net.Conn) ([]int, bool, error) {

	message, code, err := _ReadMessage(conn)
	winners := make([]int, 0)

	if err != nil {
		return winners, false, err
	}

	if code == WAIT_MSG_CODE {
		return winners, true, nil
	}

	if code == RESULTS_MSG_CODE {
		// Read the winners documents from the 4 bytes encoding
		for i := SIZE_FIELD_LENGTH + MSG_CODE_LENGTH; i < len(message)-DNI_LENGTH_IN_BYTES; i += DNI_LENGTH_IN_BYTES {
			winner := int(message[i])<<24 + int(message[i+1])<<16 + int(message[i+2])<<8 + int(message[i+3])
			winners = append(winners, winner)
		}
		return winners, false, nil
	}

	return winners, false, fmt.Errorf("invalid message code")
}

// Receives a packet from the server and returns an error if cannot read the
// packet or the packet is not a confirmation.
func RecieveBatchConfirmation(conn net.Conn) error {
	// Read confirmation from the server
	_, code, err := _ReadMessage(conn)
	if err != nil {
		return err
	}
	if code != CONFIRMATION_CODE {
		return fmt.Errorf("invalid confirmation message")
	}
	return nil
}

// Sends a buffer to the server guarding against short writes and returns an error if any.
// Adds a header to the packet with the length of the packet and the agency id.
func _SendAux(buffer []byte, conn net.Conn, agency_id, message_code int) error {

	header := make([]byte, SIZE_FIELD_LENGTH+MSG_CODE_LENGTH+AGENCY_LENGTH_IN_BYTES)
	// Add the length of the packet as the packet header
	header[0] = byte((len(buffer) + len(header)) >> 8)
	header[1] = byte((len(buffer) + len(header)))
	header[2] = byte(message_code)
	header[3] = byte(agency_id)
	buffer = append(header, buffer...)

	total_bytes_written := 0
	// Send the packet avoiding short writes
	for total_bytes_written < len(buffer) {
		bytes_written, err := conn.Write(buffer[total_bytes_written:])
		if err != nil {
			return err
		}
		total_bytes_written += bytes_written
	}

	// log.Debugf("Sent %v bytes to the server: %x", total_bytes_written, buffer)

	return nil
}

// Writes a serialization for a list of Bets in a buffer. Returns the number of Bets that were serialized.
func _SerializeBets(bets []*Bet, buffer *[]byte) int {

	bets_serialized := 0
	for _, bet := range bets {
		serialized_bet := _SerializeBet(bet)
		*buffer = append(*buffer, serialized_bet...)
		bets_serialized++

	}

	return bets_serialized
}

// Writes a serialization for a Bet and returns it.
func _SerializeBet(bet *Bet) []byte {
	packet_size := _BetSerializaitionLength(bet)
	serialized_bet := make([]byte, packet_size)

	// Bet info
	serialized_bet[0] = byte(bet.number >> 8)
	serialized_bet[1] = byte(bet.number)

	// DNI
	serialized_bet[2] = byte(bet.bettor.dni >> 24)
	serialized_bet[3] = byte(bet.bettor.dni >> 16)
	serialized_bet[4] = byte(bet.bettor.dni >> 8)
	serialized_bet[5] = byte(bet.bettor.dni)

	// Birthdate
	serialized_bet[6] = byte(bet.bettor.birthdate.Day())
	serialized_bet[7] = byte(int(bet.bettor.birthdate.Month()))
	serialized_bet[8] = byte(bet.bettor.birthdate.Year() >> 8)
	serialized_bet[9] = byte(bet.bettor.birthdate.Year())

	// Name and Lastname
	bettor_info_str := fmt.Sprintf("%s|%s|", bet.bettor.name, bet.bettor.lastname)
	copy(serialized_bet[10:], []byte(bettor_info_str))

	return serialized_bet
}

// Returns the size of the packet in bytes or -1 if the packet is too big
func _BetSerializaitionLength(bet *Bet) int {
	// Returns the size of the packet in bytes or -1 if the packet is too big

	// Max packet size in bytes
	bettor_info_str := fmt.Sprintf("%s|%s|", bet.bettor.name, bet.bettor.lastname)

	// Calculate packet size
	return NUMBER_LENGTH_IN_BYTES + DNI_LENGTH_IN_BYTES + DAY_LENGTH_IN_BYTES + MONTH_LENGTH_IN_BYTES + YEAR_LENGTH_IN_BYTES + len(bettor_info_str)
}

// Returns the bytes received from the server and its code, or an error if any.
func _ReadMessage(conn net.Conn) ([]byte, int, error) {
	// Read message from the server
	msg := make([]byte, 0)

	for len(msg) < 2 || len(msg) < int(msg[0])<<8+int(msg[1]) {
		just_read, err := bufio.NewReader(conn).ReadBytes(byte('\n'))
		msg = append(msg, just_read...)
		if err != nil {
			return msg, 0, err
		}
	}

	message_code_bytes := msg[SIZE_FIELD_LENGTH : SIZE_FIELD_LENGTH+MSG_CODE_LENGTH]
	message_code := int(message_code_bytes[0])

	return msg, message_code, nil
}
