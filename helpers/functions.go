package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"regexp"
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

	msg = tgbotapi.NewMessage(LogsChannelID, fmt.Sprintf("<code>%v</code>", err))
	msg.ParseMode = tgbotapi.ModeHTML
	_, _ = bot.Send(msg)
}

func SendToChannel(fileID string, bot *tgbotapi.BotAPI, song *entities.Song) *entities.Song {
	msg := tgbotapi.NewDocument(FilesChannelID, tgbotapi.FileID(fileID))
	msg.DisableNotification = true
	sendToChannelResponse, err := bot.Send(msg)
	if err == nil {
		if song.PDF.TgChannelMessageID != 0 {
			delMsg := tgbotapi.NewDeleteMessage(FilesChannelID, song.PDF.TgChannelMessageID)
			bot.Send(delMsg)
		}

		song.PDF.TgChannelMessageID = sendToChannelResponse.MessageID
	}

	return song
}
