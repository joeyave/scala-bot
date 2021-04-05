package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/telebot/v3"
	"regexp"
	"strings"
)

func AddCallbackData(message string, url string) string {
	message = fmt.Sprintf("%s\n<a href=\"%s\">&#8203;</a>", message, url)
	return message
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
