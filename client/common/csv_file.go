package common

import (
	"bufio"
	"strconv"

	"os"
	"strings"
)

const MAX_READ_SIZE = 1024

type CSVFile struct {
	FilePath string
	File     *os.File
	Index    int
}

// NewCSVFile Initializes a new ProcessedFile
func NewCSVFile(file_path string) *CSVFile {
	file := &CSVFile{
		FilePath: file_path,
		File:     nil,
		Index:    0,
	}
	return file
}

// Closes the file
func (f *CSVFile) Close() {
	if f.File != nil {
		f.File.Close()
	}
}

// _ReadBetsFromCSVFile Reads "bets_to-read" bets from a CSV file
func (f *CSVFile) ReadBetsFromCSVFile(bets_to_read, agency_id int) ([]*Bet, error) {
	bets := make([]*Bet, 0)
	for i := 0; i < bets_to_read; i++ {
		tokens, err := f._NextLineTokens()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return make([]*Bet, 0), err
		}
		parsed_dni, _ := strconv.Atoi(tokens["dni"])
		parsed_number, _ := strconv.Atoi(tokens["number"])
		bettorInfo := NewBettorInfo(tokens["name"], tokens["lastname"], parsed_dni, tokens["birthdate"])
		bet := NewBet(parsed_number, agency_id, *bettorInfo)
		bets = append(bets, bet)
	}
	return bets, nil
}

// Returns the next line in the CSVFile or any error that occurs
func (f *CSVFile) _NextLineTokens() (map[string]string, error) {
	if f.File == nil {
		file, err := os.Open(f.FilePath)
		if err != nil {
			return nil, err
		}
		f.File = file
		f.Index = 0
	}
	token_map := make(map[string]string)
	reader := bufio.NewReader(f.File)

	// Read a line ending in '\n'
	f.File.Seek(int64(f.Index), 0)
	content, err := reader.ReadString('\n')
	if err != nil {
		return token_map, err
	}

	// Remove whitespace and split by comma
	stripped_content := strings.TrimSpace(content)
	tokens := strings.Split(stripped_content, ",")
	token_map["name"] = tokens[0]
	token_map["lastname"] = tokens[1]
	token_map["dni"] = tokens[2]
	token_map["birthdate"] = tokens[3]
	token_map["number"] = tokens[4]
	f.Index += len(content)
	return token_map, nil
}
