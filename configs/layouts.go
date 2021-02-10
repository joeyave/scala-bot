package configs

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

var songOptionsKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: Voices}, {Text: Audios}},
	{{Text: Transpose}, {Text: Style}},
	{{Text: Menu}},
}

func GetSongOptionsKeyboard() [][]tgbotapi.KeyboardButton {
	return append(songOptionsKeyboard[:0:0], songOptionsKeyboard...)
}
