package keyboard

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"golang.org/x/exp/slices"
	"google.golang.org/api/drive/v3"
	"os"
)

func Menu(user *entity.User, bands []*entity.Band, lang string) [][]gotgbot.KeyboardButton {
	keyboard := [][]gotgbot.KeyboardButton{{}}

	if len(user.BandIDs) > 1 {
		for _, band := range bands {
			text := band.Name
			if user.BandID == band.ID {
				text = SelectedButton(band.Name).Text
			}
			keyboard[0] = append(keyboard[0], gotgbot.KeyboardButton{Text: text})
		}
	}

	keyboard = append(keyboard, [][]gotgbot.KeyboardButton{
		{{Text: txt.Get("button.schedule", lang)}},
		{{Text: txt.Get("button.songs", lang)}, {Text: txt.Get("button.stats", lang), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/web-app/statistics?bandId=%s&lang=%s", os.Getenv("BOT_DOMAIN"), user.BandID.Hex(), lang)}}},
		{{Text: txt.Get("button.settings", lang)}},
	}...)
	return keyboard
}

func Settings(user *entity.User, lang string) [][]gotgbot.InlineKeyboardButton {
	keyboard := [][]gotgbot.InlineKeyboardButton{
		{{Text: txt.Get("button.changeBand", lang), CallbackData: util.CallbackData(state.SettingsBands, "")}},
	}
	if user.IsAdmin() {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.addAdmin", lang), CallbackData: util.CallbackData(state.SettingsBandMembers, user.BandID.Hex())}})
	}

	if user.ID == 195295372 {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.cleanupDatabase", lang), CallbackData: util.CallbackData(state.SettingsCleanupDatabase, user.BandID.Hex())}})
	}

	return keyboard
}

func NavigationByToken(nextPageToken *entity.NextPageToken, lang string) [][]gotgbot.KeyboardButton {

	var keyboard [][]gotgbot.KeyboardButton

	// если есть пред стр
	if nextPageToken.GetPrevValue() != "" {
		// если нет след стр
		if nextPageToken.GetValue() != "" {
			keyboard = append(keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.prev", lang)}, {Text: txt.Get("button.menu", lang)}, {Text: txt.Get("button.next", lang)}})
		} else { // если есть след
			keyboard = append(keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.prev", lang)}, {Text: txt.Get("button.menu", lang)}})
		}
	} else { // если нет пред стр
		if nextPageToken.GetValue() != "" {
			keyboard = append(keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.menu", lang)}, {Text: txt.Get("button.next", lang)}})
		} else {
			keyboard = append(keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.menu", lang)}})
		}
	}

	return keyboard
}

func EventInit(event *entity.Event, user *entity.User, lang string) [][]gotgbot.InlineKeyboardButton {

	keyboard := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: txt.Get("button.chords", lang), CallbackData: util.CallbackData(state.EventSetlistDocs, event.ID.Hex())},
			//{Text: txt.Get("button.metronome", lang), CallbackData: util.CallbackData(state.EventSetlistMetronome, event.ID.Hex())},
		},
	}

	if user.IsAdmin() || user.IsEventMember(event) {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
			//{Text: txt.Get("button.edit", lang), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/web-app/events/%s/edit", os.Getenv("BOT_DOMAIN"), event.ID.Hex())}},
			{Text: txt.Get("button.edit", lang), CallbackData: util.CallbackData(state.EventCB, event.ID.Hex()+":edit")},
		})
	}

	return keyboard
}

func EventEdit(event *entity.Event, user *entity.User, chatID, messageID int64, lang string) [][]gotgbot.InlineKeyboardButton {

	keyboard := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: txt.Get("button.setlist", lang), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/web-app/events/%s/edit?messageId=%d&chatId=%d&userId=%d&lang=%s", os.Getenv("BOT_DOMAIN"), event.ID.Hex(), messageID, chatID, user.ID, lang)}},
			{Text: txt.Get("button.members", lang), CallbackData: util.CallbackData(state.EventMembers, event.ID.Hex())},
		},
		{
			{Text: txt.Get("button.delete", lang), CallbackData: util.CallbackData(state.EventDeleteConfirm, event.ID.Hex())},
		},
		{
			{Text: txt.Get("button.back", lang), CallbackData: util.CallbackData(state.EventCB, event.ID.Hex()+":init")},
		},
	}

	return keyboard
}

func SongInit(song *entity.Song, user *entity.User, chatID int64, messageID int64, lang string) [][]gotgbot.InlineKeyboardButton {

	var keyboard [][]gotgbot.InlineKeyboardButton

	if song.BandID == user.BandID {

		liked := false
		for _, like := range song.Likes {
			if user.ID == like.UserID {
				liked = true
				break
			}
		}

		if liked {
			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
				{Text: txt.Get("button.like", lang), CallbackData: util.CallbackData(state.SongLike, song.ID.Hex()+":dislike")},
				{Text: txt.Get("button.voices", lang, len(song.Voices)), CallbackData: util.CallbackData(state.SongVoices, song.ID.Hex())},
				{Text: txt.Get("button.more", lang), CallbackData: util.CallbackData(state.SongCB, song.ID.Hex()+":edit")},
			})
		} else {
			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
				{Text: txt.Get("button.unlike", lang), CallbackData: util.CallbackData(state.SongLike, song.ID.Hex()+":like")},
				{Text: txt.Get("button.voices", lang, len(song.Voices)), CallbackData: util.CallbackData(state.SongVoices, song.ID.Hex())},
				{Text: txt.Get("button.more", lang), CallbackData: util.CallbackData(state.SongCB, song.ID.Hex()+":edit")},
			})
		}

		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.edit", lang), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/webapp-react/#/songs/%s/edit?userId=%d&messageId=%d&chatId=%d&lang=%s", os.Getenv("BOT_DOMAIN"), song.ID.Hex(), user.ID, messageID, chatID, lang)}}})
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.edit", lang), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/web-app/songs/%s/edit?userId=%d&messageId=%d&chatId=%d&lang=%s", os.Getenv("BOT_DOMAIN"), song.ID.Hex(), user.ID, messageID, chatID, lang)}}})

	} else {
		keyboard = [][]gotgbot.InlineKeyboardButton{
			{{Text: txt.Get("button.copyToMyBand", lang), CallbackData: util.CallbackData(state.SongCopyToMyBand, song.DriveFileID)}},
			//{{Text: txt.Get("button.voices", lang, len(song.Voices)), CallbackData: util.CallbackData(state.SongVoices, song.ID.Hex())}}, // todo: enable
		}

		if user.ID == 195295372 { // todo: remove
			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.more", lang), CallbackData: util.CallbackData(state.SongCB, song.ID.Hex()+":edit")}})
		}
	}

	return keyboard
}

func SongInitIQ(song *entity.Song, user *entity.User, lang string) [][]gotgbot.InlineKeyboardButton {

	var keyboard [][]gotgbot.InlineKeyboardButton

	liked := false
	for _, like := range song.Likes {
		if user.ID == like.UserID {
			liked = true
			break
		}
	}

	if liked {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
			{Text: txt.Get("button.voices", lang, len(song.Voices)), CallbackData: util.CallbackData(state.SongVoices, song.ID.Hex())},
		})
	} else {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
			{Text: txt.Get("button.voices", lang, len(song.Voices)), CallbackData: util.CallbackData(state.SongVoices, song.ID.Hex())},
		})
	}

	return keyboard
}

func SongEdit(song *entity.Song, driveFile *drive.File, user *entity.User, lang string) [][]gotgbot.InlineKeyboardButton {

	keyboard := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: txt.Get("button.docLink", lang), Url: song.PDF.WebViewLink},
		},
		{
			{Text: txt.Get("button.style", lang), CallbackData: util.CallbackData(state.SongStyle, song.DriveFileID)},
			{Text: txt.Get("button.lyrics", lang), CallbackData: util.CallbackData(state.SongAddLyricsPage, song.DriveFileID)},
		},
	}

	if user.IsAdmin() {
		if slices.Contains(driveFile.Parents, song.Band.ArchiveFolderID) {
			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
				{Text: txt.Get("button.unarchiveText", lang), CallbackData: util.CallbackData(state.SongArchive, song.ID.Hex()+":unarchive")},
				{Text: txt.Get("button.delete", lang), CallbackData: util.CallbackData(state.SongDeleteConfirm, song.ID.Hex())},
			})
		} else {
			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
				{Text: txt.Get("button.archiveText", lang), CallbackData: util.CallbackData(state.SongArchive, song.ID.Hex()+":archive")},
				{Text: txt.Get("button.delete", lang), CallbackData: util.CallbackData(state.SongDeleteConfirm, song.ID.Hex())},
			})
		}
	}

	keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.stats", lang), CallbackData: util.CallbackData(state.SongStats, song.ID.Hex())}})
	keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", lang), CallbackData: util.CallbackData(state.SongCB, song.ID.Hex()+":init")}})

	return keyboard
}
