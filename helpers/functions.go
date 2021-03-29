package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/telebot/v3"
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
	if update != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка. Поправим.")
		_, _ = bot.Send(msg)
	}

	msg := tgbotapi.NewMessage(LogsChannelID, fmt.Sprintf("<code>%v</code>", err))
	msg.ParseMode = tgbotapi.ModeHTML
	_, _ = bot.Send(msg)
}

// TODO: переписать пачелавечески
func SendToChannel(bot *telebot.Bot, song *entities.Song) *entities.Song {
	sendNew := func() {
		msg, err := bot.Send(telebot.ChatID(FilesChannelID), &telebot.Document{
			File: telebot.File{FileID: song.PDF.TgFileID},
		}, telebot.Silent)
		if err == nil {
			song.PDF.TgChannelMessageID = msg.ID
		}
	}

	edit := func() {
		_, err := bot.EditMedia(&telebot.Message{
			ID:   song.PDF.TgChannelMessageID,
			Chat: &telebot.Chat{ID: FilesChannelID},
		}, &telebot.Document{
			File: telebot.File{FileID: song.PDF.TgFileID},
			MIME: "application/pdf",
		})

		if err != nil {
			sendNew()
		}
	}

	if song.PDF.TgChannelMessageID == 0 {
		sendNew()
	} else {
		edit()
	}

	return song
}
