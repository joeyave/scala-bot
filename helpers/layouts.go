package helpers

import tgbotapi "github.com/joeyave/telegram-bot-api/v5"

var SongActionsKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: Voices}, {Text: Audios}},
	{{Text: Transpose}, {Text: Style}},
	{{Text: Back}},
}

var RestrictedSongActionsKeyboard = [][]tgbotapi.KeyboardButton{
	{{Text: CopyToMyBand}},
	{{Text: Voices}, {Text: Audios}},
	{{Text: Back}},
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
