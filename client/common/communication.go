package common

import (
	"bufio"
	"fmt"
	"net"
)

// Constants for the communication protocol
const SIZE_FIELD_LENGTH = 2      // Size of the length field in bytes
const NUMBER_LENGTH_IN_BYTES = 2 // Size of the number field in bytes
const AGENCY_LENGTH_IN_BYTES = 1 // Size of the agency field in bytes
const DNI_LENGTH_IN_BYTES = 4    // Size of the DNI field in bytes
const YEAR_LENGTH_IN_BYTES = 2   // Size of the year field in bytes
const MONTH_LENGTH_IN_BYTES = 1  // Size of the month field in bytes
const DAY_LENGTH_IN_BYTES = 1    // Size of the day field in bytes

const DEFAULT_MAX_PACKET_SIZE = 8000 // Maximum packet size in bytes

const CONFIRMATION_CODE = 21 // What the server sends to confirm a packet

// Sends bets to the server and partitions them if they exceed the max packet size
// Returns the number of bets sent and an error if any.
func SendBets(bets []*Bet, conn net.Conn, agency_id int) error {

	total_bets_sent := 0
	for total_bets_sent < len(bets) {
		buffer := make([]byte, 0)
		bets_that_fit := _SerializeBets(bets, &buffer)
		total_bets_sent += bets_that_fit
		err := _SendAux(buffer, conn, agency_id)
		if err != nil {
			return err
		}
	}
	return nil
}

// Receives a packet from the server and returns an error if cannot read the
// packet or the packet is not a confirmation.
func RecieveConfirmation(conn net.Conn) error {
	// Read confirmation from the server
	msg, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		return err
	}

	// At least 1 byte was read
	message_code := int(msg[0])

	if message_code != CONFIRMATION_CODE {
		return fmt.Errorf("invalid confirmation message")
	}

	return nil
}

// Sends a buffer to the server guarding against short writes and returns an error if any.
func _SendAux(buffer []byte, conn net.Conn, agency_id int) error {
	total_bytes_written := 0

	header := make([]byte, SIZE_FIELD_LENGTH+AGENCY_LENGTH_IN_BYTES)
	// Add the length of the packet as the packet header
	header[0] = byte((len(buffer) + len(header)) >> 8)
	header[1] = byte(len(buffer) + len(header))
	header[2] = byte(agency_id)
	buffer = append(header, buffer...)

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

// Writes a serialization for a list of Bets in a buffer. Returns the number of Bets that could fit
// in a packet and were serialized.
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
	packet_size := _BeetSerializaitionLength(bet)
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
func _BeetSerializaitionLength(bet *Bet) int {
	// Returns the size of the packet in bytes or -1 if the packet is too big

	// Max packet size in bytes
	bettor_info_str := fmt.Sprintf("%s|%s|", bet.bettor.name, bet.bettor.lastname)

	// Calculate packet size
	return NUMBER_LENGTH_IN_BYTES + DNI_LENGTH_IN_BYTES + DAY_LENGTH_IN_BYTES + MONTH_LENGTH_IN_BYTES + YEAR_LENGTH_IN_BYTES + len(bettor_info_str)
}
