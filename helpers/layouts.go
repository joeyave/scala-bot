package helpers

import (
	"github.com/joeyave/telebot/v3"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
)

var SongActionsKeyboard = [][]telebot.ReplyButton{
	{{Text: Voices}, {Text: Audios}},
	{{Text: Transpose}, {Text: Style}},
	{{Text: Back}, {Text: Menu}},
}

var RestrictedSongActionsKeyboard = [][]telebot.ReplyButton{
	{{Text: CopyToMyBand}},
	{{Text: Voices}, {Text: Audios}},
	{{Text: Back}, {Text: Menu}},
}

var MainMenuKeyboard = [][]telebot.ReplyButton{
	{{Text: CreateDoc}},
	{{Text: Schedule}, {Text: ChangeBand}},
	{{Text: Help}},
}

var KeysKeyboard = [][]telebot.ReplyButton{
	{{Text: "C"}, {Text: "C#"}, {Text: "Db"}},
	{{Text: "D"}, {Text: "D#"}, {Text: "Eb"}},
	{{Text: "E"}},
	{{Text: "F"}, {Text: "F#"}, {Text: "Gb"}},
	{{Text: "G"}, {Text: "G#"}, {Text: "Ab"}},
	{{Text: "A"}, {Text: "A#"}, {Text: "Bb"}},
	{{Text: "B"}},
}

var TimesKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("2/4"),
		tgbotapi.NewKeyboardButton("3/4"),
		tgbotapi.NewKeyboardButton("4/4"),
	),
)

var CancelOrSkipKeyboard = [][]telebot.ReplyButton{
	{{Text: Cancel}, {Text: Skip}},
}

var FindChordsKeyboard = [][]telebot.ReplyButton{
	{{Text: FindChords}},
	{{Text: Back}, {Text: Menu}},
}

var SearchEverywhereKeyboard = [][]telebot.ReplyButton{
	{{Text: Cancel}, {Text: SearchEverywhere}},
}
