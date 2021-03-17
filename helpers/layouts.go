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

var MainMenuKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(CreateDoc),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(Schedule)),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(ChangeBand),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(Help)),
)

var KeysKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("C"),
		tgbotapi.NewKeyboardButton("C#"),
		tgbotapi.NewKeyboardButton("Db"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("D"),
		tgbotapi.NewKeyboardButton("D#"),
		tgbotapi.NewKeyboardButton("Eb"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("E"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("F"),
		tgbotapi.NewKeyboardButton("F#"),
		tgbotapi.NewKeyboardButton("Gb"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("G"),
		tgbotapi.NewKeyboardButton("G#"),
		tgbotapi.NewKeyboardButton("Ab"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("A"),
		tgbotapi.NewKeyboardButton("A#"),
		tgbotapi.NewKeyboardButton("Bb"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("B"),
	),
)

var TimesKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("2/4"),
		tgbotapi.NewKeyboardButton("3/4"),
		tgbotapi.NewKeyboardButton("4/4"),
	),
)

var CancelOrSkipKeyboard = tgbotapi.NewReplyKeyboard(
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
		tgbotapi.NewKeyboardButton(Back),
		tgbotapi.NewKeyboardButton(Menu),
	),
)

var SearchEverywhereKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(Cancel),
		tgbotapi.NewKeyboardButton(SearchEverywhere),
	),
)
