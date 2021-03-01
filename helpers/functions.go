package helpers

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"os"
	"regexp"
	"strconv"
	"strings"
)

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

func LogError(update *tgbotapi.Update, bot *tgbotapi.BotAPI, err interface{}) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка. Поправим.")
	_, _ = bot.Send(msg)

	channelId, convErr := strconv.ParseInt(os.Getenv("LOG_CHANNEL"), 10, 0)
	if convErr == nil {
		msg = tgbotapi.NewMessage(channelId, fmt.Sprintf("<code>%v</code>", err))
		msg.ParseMode = tgbotapi.ModeHTML
		_, _ = bot.Send(msg)
	}
}
