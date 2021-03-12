package handlers

import (
	"fmt"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"google.golang.org/api/drive/v3"
	"regexp"
	"strconv"
	"sync"
)

func searchSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	// Print list of found songs.
	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		{
			switch update.Message.Text {
			case "":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мне нужно название песни.")
				_, err := updateHandler.bot.Send(msg)

				if user.State.Prev != nil {
					user.State = user.State.Prev
					user.State.Index = 0
				} else {
					user.State = &entities.State{
						Index: 0,
						Name:  helpers.MainMenuState,
					}
				}
				return &user, err
			default:
				chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
				_, _ = updateHandler.bot.Send(chatAction)

				query := update.Message.Text

				if user.State.Context.Query != "" {
					query = user.State.Context.Query
				}

				query = helpers.CleanUpQuery(query)
				songNames := helpers.SplitQueryByNewlines(query)

				if len(songNames) > 1 {
					user.State = &entities.State{
						Index: 0,
						Name:  helpers.SetlistState,
						Prev: &entities.State{
							Index: 0,
							Name:  helpers.MainMenuState,
						},
						Context: user.State.Context,
					}
					user.State.Context.Setlist = songNames
					return updateHandler.enterStateHandler(update, user)

				} else if len(songNames) == 1 {
					query = songNames[0]
					user.State.Context.Query = query
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Из запроса удаляются все числа, дифизы и скобки вместе с тем, что в них.")
					_, _ = updateHandler.bot.Send(msg)

					user.State = &entities.State{
						Index: 0,
						Name:  helpers.MainMenuState,
					}
					return updateHandler.enterStateHandler(update, user)
				}

				var driveFiles []*drive.File
				var err error
				if update.Message.Text == helpers.SearchEverywhere {
					driveFiles, _, err = updateHandler.songService.QueryDrive(query, "")
				} else {
					driveFiles, _, err = updateHandler.songService.QueryDrive(query, "", user.GetFolderIDs()...)
				}

				if err != nil {
					return nil, err
				}

				if len(driveFiles) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено. Попробуй еще раз.")
					keyboard := tgbotapi.NewReplyKeyboard()
					keyboard.Keyboard = append(keyboard.Keyboard, helpers.SearchEverywhereKeyboard.Keyboard...)
					msg.ReplyMarkup = keyboard
					_, err = updateHandler.bot.Send(msg)

					user.State.Context.Query = ""
					return &user, err
				}

				keyboard := tgbotapi.NewReplyKeyboard()
				keyboard.OneTimeKeyboard = false
				keyboard.ResizeKeyboard = true

				// TODO: some sort of pagination.
				for _, song := range driveFiles {

					songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.Name))
					keyboard.Keyboard = append(keyboard.Keyboard, songButton)
				}

				keyboard.Keyboard = append(keyboard.Keyboard, helpers.SearchEverywhereKeyboard.Keyboard...)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери песню:")
				msg.ReplyMarkup = keyboard
				_, _ = updateHandler.bot.Send(msg)

				user.State.Context.DriveFiles = driveFiles
				user.State.Index++
				return &user, err
			}
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мне нужно название песни.")
			_, err := updateHandler.bot.Send(msg)
			user.State.Index--
			return &user, err
		case helpers.SearchEverywhere:
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		default:
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
			_, _ = updateHandler.bot.Send(chatAction)

			driveFiles := user.State.Context.DriveFiles
			foundIndex := len(driveFiles)
			for i := range driveFiles {
				if driveFiles[i].Name == update.Message.Text {
					foundIndex = i
					break
				}
			}

			if foundIndex != len(driveFiles) {
				user.State.Index = 0
				user.State = &entities.State{
					Index: 0,
					Name:  helpers.SongActionsState,
					Context: entities.Context{
						CurrentSongID: driveFiles[foundIndex].Id,
					},
					Prev: user.State,
				}

				return updateHandler.enterStateHandler(update, user)
			} else {
				user.State.Index--
				user.State.Context.Query = ""
				return updateHandler.enterStateHandler(update, user)
			}
		}
	})

	return helpers.SearchSongState, handleFuncs
}

func songActionsHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		songID := user.State.Context.CurrentSongID
		song, err := updateHandler.songService.FindOneByID(songID)
		if err != nil {
			return nil, err
		}

		var msg tgbotapi.DocumentConfig

		if song.HasOutdatedPDF() {
			fileReader, err := updateHandler.songService.DownloadPDFByID(song.ID)
			if err != nil {
				return nil, err
			}

			msg = tgbotapi.NewDocument(update.Message.Chat.ID, fileReader)
		} else {
			msg = tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FileID(song.PDF.TgFileID))
		}

		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.DriveFile.Name)))

		if song.BelongsToUser(user) {
			keyboard.Keyboard = append(keyboard.Keyboard, helpers.SongActionsKeyboard...)
		} else {
			keyboard.Keyboard = append(keyboard.Keyboard, helpers.RestrictedSongActionsKeyboard...)
		}
		msg.ReplyMarkup = keyboard

		sendToUserResponse, err := updateHandler.bot.Send(msg)
		if err != nil {
			fileReader, err := updateHandler.songService.DownloadPDFByID(song.ID)
			if err != nil {
				return nil, err
			}

			msg = tgbotapi.NewDocument(update.Message.Chat.ID, fileReader)

			sendToUserResponse, err = updateHandler.bot.Send(msg)
			if err != nil {
				return nil, err
			}
		}

		song, err = updateHandler.songService.FindOneByID(song.ID)
		if err != nil {
			return nil, err
		}

		if song.HasOutdatedPDF() || song.PDF.TgChannelMessageID == 0 {
			song = helpers.SendToChannel(sendToUserResponse.Document.FileID, updateHandler.bot, song)
		}

		song.PDF.TgFileID = sendToUserResponse.Document.FileID
		song.PDF.ModifiedTime = song.DriveFile.ModifiedTime

		song, err = updateHandler.songService.UpdateOne(*song)
		if err != nil {
			return nil, fmt.Errorf("failed to cache file %v", err)
		}

		user.State.Index++
		user.State.Context.CurrentSongID = song.ID

		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		song, err := updateHandler.songService.FindOneByID(user.State.Context.CurrentSongID)
		if err != nil {
			return nil, err
		}

		switch update.Message.Text {
		case helpers.Back:
			user.State = user.State.Prev
			return updateHandler.enterStateHandler(update, user)

		case helpers.Voices:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.GetVoicesState,
				Context: entities.Context{
					CurrentSongID: user.State.Context.CurrentSongID,
				},
				Prev: user.State,
			}
			return updateHandler.enterStateHandler(update, user)

		case helpers.Audios:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Функция еще не реалилованна. В будущем планируется хранинть тут аудиозаписи песни в нужной тональности.")
			_, err := updateHandler.bot.Send(msg)
			return &user, err

		case helpers.Transpose:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.TransposeSongState,
				Context: entities.Context{
					CurrentSongID: user.State.Context.CurrentSongID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0
			return updateHandler.enterStateHandler(update, user)

		case helpers.Style:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.StyleSongState,
				Context: entities.Context{
					CurrentSongID: user.State.Context.CurrentSongID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0
			return updateHandler.enterStateHandler(update, user)

		case helpers.CopyToMyBand:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.CopySongState,
				Context: entities.Context{
					CurrentSongID: user.State.Context.CurrentSongID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0
			return updateHandler.enterStateHandler(update, user)

		case song.DriveFile.Name:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, song.DriveFile.WebViewLink)
			_, err := updateHandler.bot.Send(msg)
			return &user, err

		default:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SearchSongState,
			}
			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.SongActionsState, handleFuncs
}

func transposeSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери новую тональность:")
		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.ResizeKeyboard = true
		keyboard.Keyboard = append(keyboard.Keyboard, helpers.KeysKeyboard...)
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			if user.State.Index > 0 {
				user.State.Index--
			}
			return updateHandler.enterStateHandler(update, user)
		default:
			_, err := transposer.ParseChord(update.Message.Text)
			if err != nil {
				user.State.Index--
				return updateHandler.enterStateHandler(update, user)
			}
			user.State.Context.Key = update.Message.Text

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Куда ты хочешь вставить новую тональность?")
			keyboard := tgbotapi.NewReplyKeyboard()
			keyboard.Keyboard = append(keyboard.Keyboard,
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.AppendSection)))

			sections, err := updateHandler.songService.GetSectionsByID(user.State.Context.CurrentSongID)
			if err != nil {
				return nil, err
			}

			for i := range sections {
				keyboard.Keyboard = append(keyboard.Keyboard,
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("Вместо %d-й секции", i+1))))
			}
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))

			msg.ReplyMarkup = keyboard
			_, err = updateHandler.bot.Send(msg)

			user.State.Index++
			return &user, err
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			if user.State.Index > 0 {
				user.State.Index--
			}
			return updateHandler.enterStateHandler(update, user)
		default:
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
			_, _ = updateHandler.bot.Send(chatAction)

			sections, err := updateHandler.songService.GetSectionsByID(user.State.Context.CurrentSongID)
			if err != nil {
				return &user, err
			}

			re := regexp.MustCompile("[1-9]+")
			sectionIndex, err := strconv.Atoi(re.FindString(update.Message.Text))
			if err != nil {
				sections, err = updateHandler.songService.AppendSection(user.State.Context.CurrentSongID)
				if err != nil {
					return &user, err
				}

				sectionIndex = len(sections) - 1
			} else {
				sectionIndex--
			}

			if sectionIndex >= len(sections) {
				user.State.Index--
				return updateHandler.enterStateHandler(update, user)
			}

			song, err := updateHandler.songService.FindOneByID(user.State.Context.CurrentSongID)
			if err != nil {
				return nil, err
			}

			song, err = updateHandler.songService.Transpose(*song, user.State.Context.Key, sectionIndex)
			if err != nil {
				return nil, err
			}

			updateHandler.songService.UpdateOne(*song)

			user.State = user.State.Prev
			user.State.Context.CurrentSongID = song.ID
			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.TransposeSongState, handleFuncs
}

func styleSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	// Print list of found songs.
	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		_, _ = updateHandler.bot.Send(chatAction)

		song, err := updateHandler.songService.FindOneByID(user.State.Context.CurrentSongID)
		if err != nil {
			return nil, err
		}

		song, err = updateHandler.songService.Style(*song)
		if err != nil {
			return nil, err
		}

		updateHandler.songService.UpdateOne(*song)

		user.State = user.State.Prev
		user.State.Context.CurrentSongID = song.ID
		return updateHandler.enterStateHandler(update, user)
	})
	return helpers.StyleSongState, handleFuncs
}

func copySongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери группу, в которую ты хочешь скопировать эту песню:")

		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.OneTimeKeyboard = false
		keyboard.ResizeKeyboard = true

		for i := range user.Bands {
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(user.Bands[i].Name)))
		}
		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Context.Bands = user.Bands
		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		default:
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			_, _ = updateHandler.bot.Send(chatAction)

			foundIndex := len(user.Bands)
			for i := range user.Bands {
				if user.Bands[i].Name == update.Message.Text {
					foundIndex = i
					break
				}
			}

			if foundIndex != len(user.Bands) {
				song, err := updateHandler.songService.FindOneByID(user.State.Context.CurrentSongID)
				if err != nil {
					return nil, err
				}

				copiedSong, err := updateHandler.songService.DeepCopyToFolder(*song, user.Bands[foundIndex].DriveFolderID)
				if err != nil {
					return nil, err
				}

				user.State = user.State.Prev
				user.State.Context.CurrentSongID = copiedSong.ID

				return updateHandler.enterStateHandler(update, user)

			} else {
				user.State.Index--
			}

			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.CopySongState, handleFuncs
}

func getVoicesHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		song, err := updateHandler.songService.FindOneByID(user.State.Context.CurrentSongID)
		if err != nil {
			return nil, err
		}

		voices := song.Voices

		if voices == nil || len(voices) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У этой песни нет партий. Чтобы добавить, отправь мне голосовое сообщение.")
			_, err = updateHandler.bot.Send(msg)

			user.State = user.State.Prev
			return &user, err
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери партию:")

			keyboard := tgbotapi.NewReplyKeyboard()
			keyboard.ResizeKeyboard = true

			for _, voice := range voices {
				keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(voice.Caption)))
			}
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Back)))

			msg.ReplyMarkup = keyboard

			_, err = updateHandler.bot.Send(msg)
			user.State.Index++
			return &user, err
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case helpers.Back:
			user.State = user.State.Prev
			user.State.Index = 0
			return updateHandler.enterStateHandler(update, user)
		default:
			song, err := updateHandler.songService.FindOneByID(user.State.Context.CurrentSongID)
			if err != nil {
				return nil, err
			}

			voices := song.Voices
			foundIndex := len(voices)
			for i := range voices {
				if voices[i].Caption == update.Message.Text {
					foundIndex = i
				}
			}

			if foundIndex != len(voices) {
				msg := tgbotapi.NewVoice(update.Message.Chat.ID, tgbotapi.FileID(voices[foundIndex].TgFileID))
				msg.Caption = voices[foundIndex].Caption
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: helpers.Delete}),
					tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: helpers.Back}),
				)
				keyboard.ResizeKeyboard = true
				msg.ReplyMarkup = keyboard

				_, err := updateHandler.bot.Send(msg)

				user.State.Index++
				return &user, err
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет партии c таким названием. Попробуй еще раз.")
				_, err := updateHandler.bot.Send(msg)
				return &user, err
			}
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case helpers.Back:
			user.State.Index = 0
			return updateHandler.enterStateHandler(update, user)
		case helpers.Delete:
			// TODO: handle delete
			return &user, nil
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я тебя не понимаю. Нажми на кнопку.")
			_, err := updateHandler.bot.Send(msg)
			return &user, err
		}
	})

	return helpers.GetVoicesState, handleFuncs
}

func uploadVoiceHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введи название песни, к которой ты хочешь прикрепить эту партию:")
		keyboard := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
		keyboard.ResizeKeyboard = true
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		{
			switch update.Message.Text {
			case "":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Название песни должно быть текстом! Попробуй еще раз.")
				_, err := updateHandler.bot.Send(msg)

				user.State.Index--
				return &user, err
			default:
				chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
				_, _ = updateHandler.bot.Send(chatAction)

				var driveFiles []*drive.File
				var err error

				driveFiles, _, err = updateHandler.songService.QueryDrive(update.Message.Text, "", user.GetFolderIDs()...)
				if err != nil {
					return nil, err
				}

				if len(driveFiles) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено. Попробуй другое название.")
					keyboard := tgbotapi.NewReplyKeyboard()
					keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
					msg.ReplyMarkup = keyboard
					_, err = updateHandler.bot.Send(msg)

					return &user, err
				}

				keyboard := tgbotapi.NewReplyKeyboard()
				keyboard.OneTimeKeyboard = false
				keyboard.ResizeKeyboard = true

				// TODO: some sort of pagination.
				for _, song := range driveFiles {
					songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.Name))
					keyboard.Keyboard = append(keyboard.Keyboard, songButton)
				}

				keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери песню:")
				msg.ReplyMarkup = keyboard
				_, _ = updateHandler.bot.Send(msg)

				user.State.Context.DriveFiles = driveFiles
				user.State.Index++
				return &user, err
			}
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мне нужно название песни.")
			_, err := updateHandler.bot.Send(msg)
			user.State.Index--
			return &user, err
		default:
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
			_, _ = updateHandler.bot.Send(chatAction)

			driveFiles := user.State.Context.DriveFiles
			foundIndex := len(driveFiles)
			for i := range driveFiles {
				if driveFiles[i].Name == update.Message.Text {
					foundIndex = i
					break
				}
			}

			if foundIndex != len(driveFiles) {
				song, err := updateHandler.songService.FindOneByID(driveFiles[foundIndex].Id)
				if err != nil {
					return nil, err
				}

				user.State.Context.CurrentSongID = song.ID

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправь мне название этой партии:")
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: helpers.Cancel}),
				)
				keyboard.ResizeKeyboard = true
				msg.ReplyMarkup = keyboard
				_, err = updateHandler.bot.Send(msg)

				user.State.Index++
				return &user, err
			} else {
				user.State.Index--
				return updateHandler.enterStateHandler(update, user)
			}
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Название для партии должно быть текстом! Попробуй еще раз.")
			_, err := updateHandler.bot.Send(msg)

			user.State.Index--
			return &user, err

		default:
			user.State.Context.CurrentVoice.Caption = update.Message.Text

			song, err := updateHandler.songService.FindOneByID(user.State.Context.CurrentSongID)
			if err != nil {
				return nil, err
			}

			song.Voices = append(song.Voices, user.State.Context.CurrentVoice)
			song, err = updateHandler.songService.UpdateOne(*song)
			if err != nil {
				return &user, err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добавление завершено.")
			_, err = updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SongActionsState,
				Context: entities.Context{
					CurrentSongID: user.State.Context.CurrentSongID,
				},
			}
			return updateHandler.enterStateHandler(update, user)
		}
	})
	return helpers.UploadVoiceState, handleFuncs
}

func setlistHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		if len(user.State.Context.Setlist) < 1 {
			user.State.Index = 2
			return updateHandler.enterStateHandler(update, user)
		}

		songNames := user.State.Context.Setlist

		currentSongName := songNames[0]
		user.State.Context.Setlist = songNames[1:]

		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		_, _ = updateHandler.bot.Send(chatAction)

		driveFiles, _, err := updateHandler.songService.QueryDrive(currentSongName, "", user.GetFolderIDs()...)
		if err != nil {
			return nil, err
		}

		if len(driveFiles) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("По запросу \"%s\" ничего не найдено. Напиши новое название или пропусти эту песню.", currentSongName))
			msg.ReplyMarkup = helpers.SkipSongInSetlistKeyboard

			res, err := updateHandler.bot.Send(msg)
			if err != nil {
				return &user, err
			}

			user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, res.MessageID)
			user.State.Index++
			return &user, err
		}

		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.OneTimeKeyboard = false
		keyboard.ResizeKeyboard = true

		// TODO: some sort of pagination.
		for _, song := range driveFiles {
			songButton := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(song.Name),
			)
			keyboard.Keyboard = append(keyboard.Keyboard, songButton)
		}

		keyboard.Keyboard = append(keyboard.Keyboard, helpers.SkipSongInSetlistKeyboard.Keyboard...)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Выбери песню по запросу \"%s\" или введи другое название:", currentSongName))
		msg.ReplyMarkup = keyboard
		res, err := updateHandler.bot.Send(msg)
		if err != nil {
			return &user, err
		}

		user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, res.MessageID)
		user.State.Context.DriveFiles = driveFiles
		user.State.Index++

		return &user, nil
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, update.Message.MessageID)

		switch update.Message.Text {
		case "":
		case helpers.Skip:
			user.State.Index = 0
			return updateHandler.enterStateHandler(update, user)
		}

		driveFiles := user.State.Context.DriveFiles
		foundIndex := len(driveFiles)
		for i := range driveFiles {
			if driveFiles[i].Name == update.Message.Text {
				foundIndex = i
				break
			}
		}

		if foundIndex != len(driveFiles) {
			user.State.Context.FoundSongIDs = append(user.State.Context.FoundSongIDs, driveFiles[foundIndex].Id)
		} else {
			user.State.Context.Setlist = append([]string{update.Message.Text}, user.State.Context.Setlist...)
		}

		user.State.Index = 0
		return updateHandler.enterStateHandler(update, user)
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		foundSongIDs := user.State.Context.FoundSongIDs

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(foundSongIDs))
		documents := make([]interface{}, len(foundSongIDs))
		for i := range user.State.Context.FoundSongIDs {
			go func(i int) {
				defer waitGroup.Done()

				song, err := updateHandler.songService.FindOneByID(foundSongIDs[i])
				if err != nil {
					return
				}

				if song.HasOutdatedPDF() {
					fileReader, _ := updateHandler.songService.DownloadPDFByID(song.ID)
					documents[i] = tgbotapi.NewInputMediaDocument(fileReader)
				} else {
					documents[i] = tgbotapi.NewInputMediaDocument(tgbotapi.FileID(song.PDF.TgFileID))
				}
			}(i)
		}
		waitGroup.Wait()

		const chunkSize = 10
		chunks := chunkBy(documents, chunkSize)

		for i, chunk := range chunks {
			responses, err := updateHandler.bot.SendMediaGroup(tgbotapi.NewMediaGroup(update.Message.Chat.ID, chunk))
			if err != nil {
				fromIndex := 0
				toIndex := 0 + len(chunk)

				if i-1 < len(chunks) {
					fromIndex = i * len(chunks[i-1])
					toIndex = fromIndex + len(chunks[i])
				}

				foundSongIDs := user.State.Context.FoundSongIDs[fromIndex:toIndex]

				var waitGroup sync.WaitGroup
				waitGroup.Add(len(foundSongIDs))
				documents := make([]interface{}, len(foundSongIDs))
				for i := range foundSongIDs {
					go func(i int) {
						defer waitGroup.Done()
						fileReader, _ := updateHandler.songService.DownloadPDFByID(foundSongIDs[i])
						documents[i] = tgbotapi.NewInputMediaDocument(fileReader)
					}(i)
				}
				waitGroup.Wait()

				responses, err = updateHandler.bot.SendMediaGroup(tgbotapi.NewMediaGroup(update.Message.Chat.ID, documents))
				if err != nil {
					continue
				}
			}

			for j := range responses {
				foundSongID := user.State.Context.FoundSongIDs[j+(i*len(chunk))]

				song, err := updateHandler.songService.FindOneByID(foundSongID)
				if err != nil {
					return nil, err
				}

				if song.HasOutdatedPDF() || song.PDF.TgChannelMessageID == 0 {
					song = helpers.SendToChannel(responses[j].Document.FileID, updateHandler.bot, song)
				}

				song.PDF.TgFileID = responses[j].Document.FileID
				song.PDF.ModifiedTime = song.DriveFile.ModifiedTime

				_, _ = updateHandler.songService.UpdateOne(*song)
			}
		}

		user.State = user.State.Prev
		user.State.Index = 0

		return updateHandler.enterStateHandler(update, user)
	})

	return helpers.SetlistState, handleFuncs
}

func chunkBy(items []interface{}, chunkSize int) (chunks [][]interface{}) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}
