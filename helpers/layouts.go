package helpers

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/telebot/v3"
)

func GetSongActionsKeyboard(user entities.User, song entities.Song) [][]telebot.ReplyButton {
	if song.BandID == user.BandID {
		return [][]telebot.ReplyButton{
			{{Text: Voices}, {Text: Audios}},
			{{Text: Transpose}, {Text: Style}},
			{{Text: Back}, {Text: Menu}},
		}
	} else {
		return [][]telebot.ReplyButton{
			{{Text: CopyToMyBand}},
			{{Text: Voices}, {Text: Audios}},
			{{Text: Back}, {Text: Menu}},
		}
	}

}

func GetEventActionsKeyboard(user entities.User, event entities.Event) [][]telebot.InlineButton {
	if user.Role == Admin {
		return [][]telebot.InlineButton{
			{
				{Text: FindChords, Data: AggregateCallbackData(EventActionsState, 1, "")},
			},
			{
				{Text: DeleteMember, Data: AggregateCallbackData(DeleteEventMemberState, 0, "")},
				{Text: AddMember, Data: AggregateCallbackData(AddEventMemberState, 0, "")},
			},
			{
				{Text: DeleteSong, Data: AggregateCallbackData(DeleteEventSongState, 0, "")},
				{Text: AddSong, Data: AggregateCallbackData(AddEventSongState, 0, "")},
			},
			{
				{Text: ChangeSongsOrder, Data: AggregateCallbackData(ChangeSongOrderState, 0, "")},
			},
		}
	}

	for _, membership := range event.Memberships {
		if user.ID == membership.UserID {
			return [][]telebot.InlineButton{
				{
					{Text: FindChords, Data: AggregateCallbackData(EventActionsState, 1, "")},
				},
				{
					{Text: DeleteSong, Data: AggregateCallbackData(DeleteEventSongState, 0, "")},
					{Text: AddSong, Data: AggregateCallbackData(AddEventSongState, 0, "")},
				},
				{
					{Text: ChangeSongsOrder, Data: AggregateCallbackData(ChangeSongOrderState, 0, "")},
				},
			}
		}
	}

	return [][]telebot.InlineButton{
		{
			{Text: FindChords, Data: AggregateCallbackData(EventActionsState, 1, "")},
		},
	}
}

var MainMenuKeyboard = [][]telebot.ReplyButton{
	//{{Text: Schedule}},
	{{Text: Schedule}},
	{{Text: Songs}, {Text: Members}},
	//{{Text: CreateDoc}},
	{{Text: Settings}},
}

var SettingsKeyboard = [][]telebot.ReplyButton{
	{{Text: BandSettings}, {Text: ProfileSettings}},
	//{{Text: ChangeBand}, {Text: CreateRole}},
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
