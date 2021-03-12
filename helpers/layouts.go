package helpers

import tgbotapi "github.com/joeyave/telegram-bot-api/v5"

var SongActionsKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: Voices}, {Text: Audios}},
	{{Text: Transpose}, {Text: Style}},
	{{Text: Back}, {Text: Menu}},
}

var RestrictedSongActionsKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: CopyToMyBand}},
	{{Text: Voices}, {Text: Audios}},
	{{Text: Back}, {Text: Menu}},
}

var MainMenuKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: CreateDoc}},
	{{Text: Schedule}},
	{{Text: Help}},
}

var KeysKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: "C"}, {Text: "C#"}, {Text: "Db"}},
	{{Text: "D"}, {Text: "D#"}, {Text: "Eb"}},
	{{Text: "E"}},
	{{Text: "F"}, {Text: "F#"}, {Text: "Gb"}},
	{{Text: "G"}, {Text: "G#"}, {Text: "Ab"}},
	{{Text: "A"}, {Text: "A#"}, {Text: "Bb"}},
	{{Text: "B"}},
	{{Text: Cancel}},
}

var SkipSongInSetlistKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(Cancel),
		tgbotapi.NewKeyboardButton(Skip),
	),
)

var FindChordsKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(FindChords),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(Menu),
	),
)

var SearchEverywhereKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(Cancel),
		tgbotapi.NewKeyboardButton(SearchEverywhere),
	),
)
