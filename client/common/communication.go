package common

import (
	"bufio"
	"fmt"
	"math"
	"net"

	log "github.com/sirupsen/logrus"
)

const PACKET_SIZE_LENGTH_IN_BYTES = 1
const NUMBER_LENGTH_IN_BYTES = 2
const AGENCY_LENGTH_IN_BYTES = 1
const DNI_LENGTH_IN_BYTES = 4
const YEAR_LENGTH_IN_BYTES = 2
const MONTH_LENGTH_IN_BYTES = 1
const DAY_LENGTH_IN_BYTES = 1

func SendBet(bet *Bet, conn *net.Conn) error {
	// Serializes bet and sends it to the server.

	serializedBet, err := SerializeBet(bet)
	log.Debugf("Serialized Bet: %x", serializedBet)
	log.Debugf("Len: %d", len(serializedBet))

	total_bytes_written := 0
	bytes_written := 0

	// No error in the beginning

	for total_bytes_written < len(serializedBet) && err == nil {
		bytes_written, err = (*conn).Write(serializedBet[total_bytes_written:])
		total_bytes_written += bytes_written
	}

	return err
}

func RecieveConfirmation(conn *net.Conn) (int, int, error) {
	// Read confirmation from the server
	msg, err := bufio.NewReader(*conn).ReadString('\n')

	if err != nil {
		return 0, 0, err
	}

	// At least 1 byte was read
	message_code := int(msg[0])

	if message_code != 21 {
		return 0, 0, fmt.Errorf("invalid confirmation message")
	}

	// Remove the new-line character
	msg = msg[:len(msg)-1]

	for len(msg) < 6 {
		reading, err := bufio.NewReader(*conn).ReadString('\n')
		if err != nil {
			return 0, 0, err
		}
		msg += reading[:len(msg)-1]
	}

	dni := int(msg[1])<<24 | int(msg[2])<<16 | int(msg[3])<<8 | int(msg[4])
	bet_number := int(msg[5])<<8 | int(msg[6])

	return dni, bet_number, error(nil)
}

// func ParseConfirmation(response []byte) (int, int, error) {
// 	// Parse the confirmation message
// 	// Returns the number and the dni of the bettor

// 	if int(msg[0]) != 21 || len(msg) != 8{
// 		return 0, 0, fmt.Errorf("Invalid confirmation message")
// 	}

// 	//Int from bytes
// 	next_two_bytes =
// }

func SerializeBet(bet *Bet) ([]byte, error) {
	// Size in bytes of each fixed field

	packet_size := _PacketLength(bet)

	if packet_size == -1 {
		return nil, fmt.Errorf("name and lastname are too long")
	}

	serialized_bet := make([]byte, packet_size)

	// Example
	// SIZE AGENCY NUMBER NUMBER DNI DNI DAY MONTH YEAR YEAR NAME | LASTNAME

	// Bet info
	serialized_bet[0] = byte(packet_size)
	serialized_bet[1] = byte(bet.agency)
	serialized_bet[2] = byte(bet.number >> 8)
	serialized_bet[3] = byte(bet.number)

	// DNI
	serialized_bet[4] = byte(bet.bettor.dni >> 24)
	serialized_bet[5] = byte(bet.bettor.dni >> 16)
	serialized_bet[6] = byte(bet.bettor.dni >> 8)
	serialized_bet[7] = byte(bet.bettor.dni)

	// Birthdate
	serialized_bet[8] = byte(bet.bettor.birthdate.Day())
	serialized_bet[9] = byte(int(bet.bettor.birthdate.Month()))
	serialized_bet[10] = byte(bet.bettor.birthdate.Year() >> 8)
	serialized_bet[11] = byte(bet.bettor.birthdate.Year())

	// Name and Lastname
	bettor_info_str := fmt.Sprintf("%s|%s", bet.bettor.name, bet.bettor.lastname)
	copy(serialized_bet[12:], []byte(bettor_info_str))

	return serialized_bet, error(nil)
}

func _PacketLength(bet *Bet) int {
	// Returns the size of the packet in bytes or -1 if the packet is too big

	// Max packet size in bytes
	MAX_PACKET_SIZE_IN_BYTES := int(math.Pow(2, float64((PACKET_SIZE_LENGTH_IN_BYTES*8))) - 1) // 255 BYTES
	bettor_info_str := fmt.Sprintf("%s|%s", bet.bettor.name, bet.bettor.lastname)

	// Calculate packet size
	packet_size := PACKET_SIZE_LENGTH_IN_BYTES + AGENCY_LENGTH_IN_BYTES + NUMBER_LENGTH_IN_BYTES + DNI_LENGTH_IN_BYTES + DAY_LENGTH_IN_BYTES + MONTH_LENGTH_IN_BYTES + YEAR_LENGTH_IN_BYTES + len(bettor_info_str)

	// Check if packet size is valid
	if packet_size <= MAX_PACKET_SIZE_IN_BYTES {
		return packet_size
	} else {
		return -1
	}
}
