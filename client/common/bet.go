package common

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
)

type BettorInfo struct {
	name      string
	lastname  string
	dni       string
	birthdate string
}

type Bet struct {
	number int
	agency int
	bettor BettorInfo
}

func NewBettorInfo(name, lastname, dni, birthdate string) *BettorInfo {
	bettorInfo := &BettorInfo{
		name:      name,
		lastname:  lastname,
		dni:       dni,
		birthdate: birthdate,
	}
	return bettorInfo
}

func NewBet(number, agency int, bettor BettorInfo) *Bet {
	if float64(number) > math.Pow(2, 16)-1 || number < 0 {
		log.Errorf("action: new_bet | result: fail | error: invalid_bet_number | number: %d", number)
		panic("Invalid Bet Number")
	}

	bet := &Bet{
		number: number,
		agency: agency,
		bettor: bettor,
	}
	return bet
}

func SerializeBet(bet *Bet) []byte {
	PACKET_SIZE_LENGTH_IN_BYTES := 1
	CHOSEN_NUMBER_SIZE_IN_BYTES := 2
	AGENCY_ID_SIZE_IN_BYTES := 1
	MAX_PACKET_SIZE_IN_BYTES := int(math.Pow(2, float64((PACKET_SIZE_LENGTH_IN_BYTES*8))) - 1) // 255 BYTES

	bettor_info_str := fmt.Sprintf("%s|%s|%s|%s", bet.bettor.name, bet.bettor.lastname, bet.bettor.dni, bet.bettor.birthdate)

	packet_size := PACKET_SIZE_LENGTH_IN_BYTES + AGENCY_ID_SIZE_IN_BYTES + CHOSEN_NUMBER_SIZE_IN_BYTES + len(bettor_info_str)

	if packet_size > MAX_PACKET_SIZE_IN_BYTES {
		log.Errorf("action: serialize_bet | result: fail | error: content too long")
		panic("Cannot Serialize Bet: Content too long")
	}

	serialized_bet := make([]byte, packet_size)

	serialized_bet[0] = byte(packet_size)
	serialized_bet[1] = byte(bet.agency)
	serialized_bet[2] = byte(bet.number >> 8)
	serialized_bet[3] = byte(bet.number)
	copy(serialized_bet[4:], []byte(bettor_info_str))
	return serialized_bet
}
