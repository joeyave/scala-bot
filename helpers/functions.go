package helpers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"github.com/rs/zerolog/log"
)

func AddCallbackData(message string, url string) string {
	message = fmt.Sprintf("%s\n<a href=\"%s\">&#8203;</a>", message, url)
	return message
}

func ParseCallbackData(data string) (int, int, string) {
	
	log.Debug().Msg(data)
	parsedData := strings.Split(data, ":")
	stateStr := parsedData[0]
	indexStr := parsedData[1]
	payload := strings.Join(parsedData[2:], ":")

	state, err := strconv.Atoi(stateStr)
	if err != nil {
		state = 0 // TODO
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		index = 0 // TODO
	}

	return state, index, payload
}

func AggregateCallbackData(state int, index int, payload string) string {
	return fmt.Sprintf("%d:%d:%s", state, index, payload)
}

func JsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}

func CleanUpQuery(query string) string {
	numbersRegex := regexp.MustCompile(`\(.*?\)|[1-9.()_]*`)
	return numbersRegex.ReplaceAllString(query, "")
}

func SplitQueryByNewlines(query string) []string {
	newLinesRegex := regexp.MustCompile(`\s*[\t\r\n]+`)
	songNames := strings.Split(newLinesRegex.ReplaceAllString(query, "\n"), "\n")
	for _, songName := range songNames {
		songName = strings.TrimSpace(songName)
	}

	return songNames
}
