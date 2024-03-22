package common

import (
	"bufio"

	"os"
	"strings"
)

const MAX_READ_SIZE = 1024

type CSVFile struct {
	FilePath string
	Index    int
}

// NewCSVFile Initializes a new ProcessedFile
func NewCSVFile(file_path string) *CSVFile {
	file := &CSVFile{
		FilePath: file_path,
		Index:    0,
	}
	return file
}

// Returns the next line in the CSVFile or any error that occurs
func (f *CSVFile) GetNextLine() (map[string]string, error) {
	token_map := make(map[string]string)
	file, err := os.Open(f.FilePath)
	if err != nil {
		return token_map, err
	}
	defer file.Close()
	_, err = file.Seek(int64(f.Index), 0)
	if err != nil {
		return token_map, err
	}

	_, err = file.Seek(int64(f.Index), 0)
	if err != nil {
		return token_map, err
	}

	reader := bufio.NewReader(file)

	// Read a line ending in '\n'
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
