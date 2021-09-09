package helpers

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/telebot/v3"
	"google.golang.org/api/drive/v3"
)

func GetSongActionsKeyboard(user entities.User, song entities.Song, driveFile drive.File) [][]telebot.InlineButton {
	if song.BandID == user.BandID {
		return [][]telebot.InlineButton{
			{{Text: LinkToTheDoc, URL: driveFile.WebViewLink}},
			{{Text: Voices, Data: AggregateCallbackData(GetVoicesState, 0, "")}},
			{
				{Text: Transpose, Data: AggregateCallbackData(TransposeSongState, 0, "")},
				{Text: Style, Data: AggregateCallbackData(StyleSongState, 0, "")},
			},
		}
	} else {
		return [][]telebot.InlineButton{
			{{Text: driveFile.Name, URL: driveFile.WebViewLink}},
			{{Text: CopyToMyBand}},
			{{Text: Voices}},
		}
	}

}

func GetEventActionsKeyboard(user entities.User, event entities.Event) [][]telebot.InlineButton {
	member := false
	for _, membership := range event.Memberships {
		if user.ID == membership.UserID {
			member = true
		}
	}

	if user.Role == Admin || member {
		return [][]telebot.InlineButton{
			{
				{Text: FindChords, Data: AggregateCallbackData(EventActionsState, 1, "")},
			},
			{
				{Text: Edit, Data: AggregateCallbackData(EditInlineKeyboardState, 0, "")},
			},
		}
	}

	return [][]telebot.InlineButton{
		{
			{Text: FindChords, Data: AggregateCallbackData(EventActionsState, 1, "")},
		},
	}
}

func GetEditEventKeyboard(user entities.User) [][]telebot.InlineButton {
	if user.Role == Admin {
		return [][]telebot.InlineButton{
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
			{
				{Text: ChangeEventDate, Data: AggregateCallbackData(ChangeEventDateState, 0, "")},
			},
			{
				{Text: Delete, Data: AggregateCallbackData(DeleteEventState, 0, "")},
			},
			{
				{Text: Back, Data: AggregateCallbackData(EventActionsState, 0, "")},
			},
		}
	}

	return [][]telebot.InlineButton{
		{
			{Text: DeleteSong, Data: AggregateCallbackData(DeleteEventSongState, 0, "")},
			{Text: AddSong, Data: AggregateCallbackData(AddEventSongState, 0, "")},
		},
		{
			{Text: ChangeSongsOrder, Data: AggregateCallbackData(ChangeSongOrderState, 0, "")},
		},
		{
			{Text: Back, Data: AggregateCallbackData(EventActionsState, 0, "")},
		},
	}
}

var MainMenuKeyboard = [][]telebot.ReplyButton{
	{{Text: Schedule}},
	{{Text: Songs}, {Text: Members}},
	{{Text: Settings}},
}

var SettingsKeyboard = [][]telebot.ReplyButton{
	{{Text: BandSettings}, {Text: ProfileSettings}},
	{{Text: Back}},
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

var SearchEverywhereKeyboard = [][]telebot.ReplyButton{
	{{Text: Cancel}, {Text: SearchEverywhere}},
}

var ConfirmDeletingEventKeyboard = [][]telebot.InlineButton{
	{{Text: Cancel, Data: AggregateCallbackData(EventActionsState, 0, "EditEventKeyboard")}, {Text: Yes, Data: AggregateCallbackData(DeleteEventState, 1, "")}},
}
