package helpers

import tgbotapi "github.com/joeyave/telegram-bot-api/v5"

var songOptionsKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: Voices}, {Text: Audios}},
	{{Text: Transpose}, {Text: Style}},
	{{Text: Menu}},
}

func GetSongOptionsKeyboard() [][]tgbotapi.KeyboardButton {
	return append(songOptionsKeyboard[:0:0], songOptionsKeyboard...)
}

var MainMenuKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: CreateDoc}},
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
