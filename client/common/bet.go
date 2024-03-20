package common

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

type BettorInfo struct {
	name      string
	lastname  string
	dni       int
	birthdate time.Time
}

type Bet struct {
	number int
	agency int
	bettor BettorInfo
}

func NewBettorInfo(name, lastname string, dni int, birthdate string) *BettorInfo {
	parsedDate, _ := time.Parse("2006-01-02", birthdate)
	bettorInfo := &BettorInfo{
		name:      name,
		lastname:  lastname,
		dni:       dni,
		birthdate: parsedDate,
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
