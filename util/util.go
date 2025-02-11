package util

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters"
	"regexp"
	"strconv"
	"strings"
)

func CleanUpText(text string) string {
	numbersRegex := regexp.MustCompile(`\(.*?\)|[1-9.()_]*`)
	return numbersRegex.ReplaceAllString(text, "")
}

var newLinesRegex = regexp.MustCompile(`\s*[\t\r\n]+`)

func SplitTextByNewlines(query string) []string {
	songNames := strings.Split(newLinesRegex.ReplaceAllString(query, "\n"), "\n")
	for i := range songNames {
		songNames[i] = strings.TrimSpace(songNames[i])
	}

	return songNames
}

func IetfToIsoLangCode(languageCode string) string {
	switch languageCode {
	case "ru":
		return "ru_RU"
	default:
		return "uk_UA"
	}
}

func CallbackData(state int, payload string) string {
	callbackData := strconv.Itoa(state) + ":" + payload
	if len(callbackData) > 64 {
		panic(fmt.Sprintf("size of callback_data is bigger than 64: size=%d, callback_data=%s", len(callbackData), callbackData))
	}
	return callbackData
}

func CallbackState(state int) filters.CallbackQuery {
	return func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, strconv.Itoa(state)+":")
	}
}

func ParseCallbackPayload(data string) string {
	parsedData := strings.Split(data, ":")
	return strings.Join(parsedData[1:], ":")
}

// todo
func ToUpperFirstLetter(input string) string {
	if input == "" {
		return ""
	}
	//return strings.ToUpper(input[:1]) + input[1:]
	return input
}

const CallbackCacheURL = "https://t.me/callbackCache"
