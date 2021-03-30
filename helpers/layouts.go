package helpers

import (
	"github.com/joeyave/telebot/v3"
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

var EventActionsKeyboard = [][]telebot.ReplyButton{
	{{Text: AddMember}},
}

var MainMenuKeyboard = [][]telebot.ReplyButton{
	{{Text: Schedule}},
	{{Text: CreateDoc}},
	{{Text: Settings}},
}

var SettingsKeyboard = [][]telebot.ReplyButton{
	{{Text: ChangeBand}, {Text: CreateRole}},
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

var TimesKeyboard = [][]telebot.ReplyButton{
	{{Text: "2/4"}, {Text: "3/4"}, {Text: "4/4"}},
}

var CancelOrSkipKeyboard = [][]telebot.ReplyButton{
	{{Text: Cancel}, {Text: Skip}},
}

var FindChordsKeyboard = [][]telebot.ReplyButton{
	{{Text: FindChords}},
	{{Text: Back}, {Text: Menu}},
}

var BackOrMenuKeyboard = [][]telebot.ReplyButton{
	{{Text: Back}, {Text: Menu}},
}

var SearchEverywhereKeyboard = [][]telebot.ReplyButton{
	{{Text: Cancel}, {Text: SearchEverywhere}},
}
