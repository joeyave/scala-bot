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
	"time"
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

				if query == helpers.SearchEverywhere || query == helpers.Back {
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
					driveFiles, _, err = updateHandler.driveFileService.FindSomeByNameAndFolderID(query, "", "")
				} else {
					driveFiles, _, err = updateHandler.driveFileService.FindSomeByNameAndFolderID(query, user.Band.DriveFolderID, "")
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
						CurrentDriveFileID: driveFiles[foundIndex].Id,
					},
					Prev: user.State,
				}

				return updateHandler.enterStateHandler(update, user)
			} else {
				user.State.Index--
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

		songID := user.State.Context.CurrentDriveFileID
		song, err := updateHandler.songService.FindOneByDriveFileID(songID)
		if err != nil {
			err = nil
			driveFile, err := updateHandler.driveFileService.FindOneByID(songID)
			if err != nil {
				return nil, err
			}

			song = &entities.Song{
				DriveFileID: driveFile.Id,
			}

			for _, parentFolderID := range driveFile.Parents {
				band, err := updateHandler.bandService.FindOneByDriveFolderID(parentFolderID)
				if err == nil {
					song.BandID = band.ID
					break
				}
			}

			song, err = updateHandler.songService.UpdateOne(*song)
			if err != nil {
				return nil, err
			}
		}

		var msg tgbotapi.DocumentConfig

		if song.HasOutdatedPDF() {
			reader, err := updateHandler.driveFileService.DownloadOneByID(songID)
			if err != nil {
				return nil, err
			}

			msg = tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FileReader{
				Name:   song.DriveFile.Name + ".pdf",
				Reader: *reader,
			})
		} else {
			msg = tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FileID(song.PDF.TgFileID))
		}

		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.DriveFile.Name)))

		//if song.Band == nil {
		//	song.BandID = user.BandID
		//	song, err = updateHandler.songService.UpdateOne(*song)
		//	if err != nil {
		//		return nil, err
		//	}
		//}

		if song.BandID == user.BandID {
			keyboard.Keyboard = append(keyboard.Keyboard, helpers.SongActionsKeyboard...)
		} else {
			keyboard.Keyboard = append(keyboard.Keyboard, helpers.RestrictedSongActionsKeyboard...)
		}
		msg.ReplyMarkup = keyboard

		sendToUserResponse, err := updateHandler.bot.Send(msg)
		if err != nil {
			reader, err := updateHandler.driveFileService.DownloadOneByID(song.DriveFileID)
			if err != nil {
				return nil, err
			}

			msg = tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FileReader{
				Name:   song.DriveFile.Name + ".pdf",
				Reader: *reader,
			})
			msg.ReplyMarkup = keyboard

			sendToUserResponse, err = updateHandler.bot.Send(msg)
			if err != nil {
				return nil, err
			}
		}

		//song, err = updateHandler.songService.FindOrCreateOneByDriveFileID(song.DriveFileID)
		//if err != nil {
		//	return nil, err
		//}

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
		user.State.Context.CurrentDriveFileID = song.DriveFileID

		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		song, err := updateHandler.songService.FindOneByDriveFileID(user.State.Context.CurrentDriveFileID)
		if err != nil {
			return nil, err
		}

		switch update.Message.Text {
		case helpers.Back:
			user.State = user.State.Prev

		case helpers.Voices:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.GetVoicesState,
				Context: entities.Context{
					CurrentDriveFileID: user.State.Context.CurrentDriveFileID,
				},
				Prev: user.State,
			}

		case helpers.Audios:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Функция еще не реалилованна. В будущем планируется хранинть тут аудиозаписи песни в нужной тональности.")
			_, err := updateHandler.bot.Send(msg)
			return &user, err

		case helpers.Transpose:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.TransposeSongState,
				Context: entities.Context{
					CurrentDriveFileID: user.State.Context.CurrentDriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case helpers.Style:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.StyleSongState,
				Context: entities.Context{
					CurrentDriveFileID: user.State.Context.CurrentDriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case helpers.CopyToMyBand:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.CopySongState,
				Context: entities.Context{
					CurrentDriveFileID: user.State.Context.CurrentDriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case helpers.Delete:
			user.State = &entities.State{
				Name: helpers.DeleteSongState,
				Context: entities.Context{
					CurrentDriveFileID: user.State.Context.CurrentDriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case song.DriveFile.Name:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, song.DriveFile.WebViewLink)
			_, err := updateHandler.bot.Send(msg)
			return &user, err

		default:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SearchSongState,
			}
		}

		return updateHandler.enterStateHandler(update, user)
	})

	return helpers.SongActionsState, handleFuncs
}

func transposeSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери новую тональность:")
		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.ResizeKeyboard = true
		keyboard.Keyboard = append(keyboard.Keyboard, helpers.KeysKeyboard.Keyboard...)
		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
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
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.AppendSection),
				))

			sectionsNumber, err := updateHandler.driveFileService.GetSectionsNumber(user.State.Context.CurrentDriveFileID)
			if err != nil {
				return nil, err
			}

			for i := 0; i < sectionsNumber; i++ {
				keyboard.Keyboard = append(keyboard.Keyboard,
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(fmt.Sprintf("Вместо %d-й секции", i+1)),
					))
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

			re := regexp.MustCompile("[1-9]+")
			sectionIndex, err := strconv.Atoi(re.FindString(update.Message.Text))
			if err != nil {
				sectionIndex = -1
			} else {
				sectionIndex--
			}

			driveFile, err := updateHandler.driveFileService.TransposeOne(user.State.Context.CurrentDriveFileID, user.State.Context.Key, sectionIndex)
			if err != nil {
				return nil, err
			}

			song, err := updateHandler.songService.FindOneByDriveFileID(driveFile.Id)
			if err != nil {
				return nil, err
			}

			fakeTime, _ := time.Parse("2006", "2006")
			song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)

			_, err = updateHandler.songService.UpdateOne(*song)
			if err != nil {
				return nil, err
			}

			user.State = user.State.Prev
			user.State.Context.CurrentDriveFileID = driveFile.Id
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

		driveFile, err := updateHandler.driveFileService.StyleOne(user.State.Context.CurrentDriveFileID)
		if err != nil {
			return nil, err
		}

		song, err := updateHandler.songService.FindOneByDriveFileID(driveFile.Id)
		if err != nil {
			return nil, err
		}

		fakeTime, _ := time.Parse("2006", "2006")
		song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)

		_, err = updateHandler.songService.UpdateOne(*song)
		if err != nil {
			return nil, err
		}

		user.State = user.State.Prev
		user.State.Context.CurrentDriveFileID = driveFile.Id
		return updateHandler.enterStateHandler(update, user)
	})
	return helpers.StyleSongState, handleFuncs
}

func copySongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		_, _ = updateHandler.bot.Send(chatAction)

		file, err := updateHandler.driveFileService.FindOneByID(user.State.Context.CurrentDriveFileID)
		if err != nil {
			return nil, err
		}

		file = &drive.File{
			Name:    file.Name,
			Parents: []string{user.Band.DriveFolderID},
		}

		copiedSong, err := updateHandler.driveFileService.CloneOne(user.State.Context.CurrentDriveFileID, file)
		if err != nil {
			return nil, err
		}

		user.State = user.State.Prev
		user.State.Context.CurrentDriveFileID = copiedSong.Id

		return updateHandler.enterStateHandler(update, user)
	})

	return helpers.CopySongState, handleFuncs
}

func createSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправь название:")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Cancel),
			),
		)
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		default:
			user.State.Context.CreateSongPayload.Name = update.Message.Text

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправь слова:")
			msg.ReplyMarkup = helpers.CancelOrSkipKeyboard
			_, err := updateHandler.bot.Send(msg)

			user.State.Index++
			return &user, err
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.Lyrics = update.Message.Text
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери или отправь тональность:")
		keyboard := tgbotapi.NewReplyKeyboard(helpers.KeysKeyboard.Keyboard...)
		keyboard.Keyboard = append(keyboard.Keyboard, helpers.CancelOrSkipKeyboard.Keyboard...)
		msg.ReplyMarkup = keyboard

		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.Key = update.Message.Text
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправь темп:")
		msg.ReplyMarkup = helpers.CancelOrSkipKeyboard

		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.BPM = update.Message.Text
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери или отправь размер:")
		keyboard := tgbotapi.NewReplyKeyboard(helpers.TimesKeyboard.Keyboard...)
		keyboard.Keyboard = append(keyboard.Keyboard, helpers.CancelOrSkipKeyboard.Keyboard...)
		msg.ReplyMarkup = keyboard

		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.Time = update.Message.Text
		}

		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		file := &drive.File{
			Name:     user.State.Context.CreateSongPayload.Name,
			Parents:  []string{user.Band.DriveFolderID},
			MimeType: "application/vnd.google-apps.document",
		}
		newFile, err := updateHandler.driveFileService.CreateOne(
			file,
			user.State.Context.CreateSongPayload.Lyrics,
			user.State.Context.CreateSongPayload.Key,
			user.State.Context.CreateSongPayload.BPM,
			user.State.Context.CreateSongPayload.Time,
		)

		if err != nil {
			return nil, err
		}

		newFile, err = updateHandler.driveFileService.StyleOne(newFile.Id)
		if err != nil {
			return nil, err
		}

		user.State = &entities.State{
			Index: 0,
			Name:  helpers.SongActionsState,
			Context: entities.Context{
				CurrentDriveFileID: newFile.Id,
			},
		}

		return &user, err
	})

	return helpers.CreateSongState, handleFuncs
}

func deleteSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	// TODO: allow deleting Song only if it belongs to the User's Band.
	// TODO: delete from channel.
	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		if user.Role == helpers.Admin {
			err := updateHandler.songService.DeleteOneByID(user.State.Context.CurrentDriveFileID)
			if err != nil {
				return nil, err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Удалено.")
			_, _ = updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Name: helpers.MainMenuState,
			}
			return updateHandler.enterStateHandler(update, user)

		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для удаления песни нужно быть администратором группы.")
			_, _ = updateHandler.bot.Send(msg)
		}

		user.State = user.State.Prev
		return updateHandler.enterStateHandler(update, user)
	})

	return helpers.DeleteSongState, handleFuncs
}

func getVoicesHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		song, err := updateHandler.songService.FindOneByDriveFileID(user.State.Context.CurrentDriveFileID)
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
			song, err := updateHandler.songService.FindOneByDriveFileID(user.State.Context.CurrentDriveFileID)
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
				msg := tgbotapi.NewVoice(update.Message.Chat.ID, tgbotapi.FileID(voices[foundIndex].FileID))
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
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Cancel),
			),
		)
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
				user.State.Index--
				return updateHandler.enterStateHandler(update, user)
			default:
				chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
				_, _ = updateHandler.bot.Send(chatAction)

				var driveFiles []*drive.File
				var err error

				driveFiles, _, err = updateHandler.driveFileService.FindSomeByNameAndFolderID(update.Message.Text, user.Band.DriveFolderID, "")
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
				song, err := updateHandler.songService.FindOneByDriveFileID(driveFiles[foundIndex].Id)
				if err != nil {
					return nil, err
				}

				user.State.Context.CurrentDriveFileID = song.DriveFileID

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

			song, err := updateHandler.songService.FindOneByDriveFileID(user.State.Context.CurrentDriveFileID)
			if err != nil {
				return nil, err
			}

			user.State.Context.CurrentVoice.SongID = song.ID

			_, err = updateHandler.voiceService.UpdateOne(*user.State.Context.CurrentVoice)
			if err != nil {
				return &user, err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добавление завершено.")
			_, err = updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SongActionsState,
				Context: entities.Context{
					CurrentDriveFileID: user.State.Context.CurrentDriveFileID,
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

		driveFiles, _, err := updateHandler.driveFileService.FindSomeByNameAndFolderID(currentSongName, user.Band.DriveFolderID, "")
		if err != nil {
			return nil, err
		}

		if len(driveFiles) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("По запросу \"%s\" ничего не найдено. Напиши новое название или пропусти эту песню.", currentSongName))
			msg.ReplyMarkup = helpers.CancelOrSkipKeyboard

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

		keyboard.Keyboard = append(keyboard.Keyboard, helpers.CancelOrSkipKeyboard.Keyboard...)

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
			user.State.Context.FoundDriveFileIDs = append(user.State.Context.FoundDriveFileIDs, driveFiles[foundIndex].Id)
		} else {
			user.State.Context.Setlist = append([]string{update.Message.Text}, user.State.Context.Setlist...)
		}

		user.State.Index = 0
		return updateHandler.enterStateHandler(update, user)
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		foundSongIDs := user.State.Context.FoundDriveFileIDs

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(foundSongIDs))
		documents := make([]interface{}, len(foundSongIDs))
		for i := range user.State.Context.FoundDriveFileIDs {
			go func(i int) {
				defer waitGroup.Done()

				song, err := updateHandler.songService.FindOneByDriveFileID(foundSongIDs[i])
				if err != nil {
					err = nil
					driveFile, err := updateHandler.driveFileService.FindOneByID(foundSongIDs[i])
					if err != nil {
						return
					}

					song = &entities.Song{
						DriveFileID: driveFile.Id,
					}

					for _, parentFolderID := range driveFile.Parents {
						band, err := updateHandler.bandService.FindOneByDriveFolderID(parentFolderID)
						if err == nil {
							song.BandID = band.ID
							break
						}
					}

					song, err = updateHandler.songService.UpdateOne(*song)
					if err != nil {
						return
					}
				}

				if song.HasOutdatedPDF() {
					reader, err := updateHandler.driveFileService.DownloadOneByID(song.DriveFileID)
					if err != nil {
						return
					}

					documents[i] = tgbotapi.NewInputMediaDocument(tgbotapi.FileReader{
						Name:   song.DriveFile.Name + ".pdf",
						Reader: *reader,
					})
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

			// TODO: check for bugs.
			if err != nil {
				fromIndex := 0
				toIndex := 0 + len(chunk)

				if i-1 < len(chunks) {
					fromIndex = i * len(chunks[i-1])
					toIndex = fromIndex + len(chunks[i])
				}

				foundDriveFileIDs := user.State.Context.FoundDriveFileIDs[fromIndex:toIndex]

				var waitGroup sync.WaitGroup
				waitGroup.Add(len(foundDriveFileIDs))
				documents := make([]interface{}, len(foundDriveFileIDs))
				for i := range foundDriveFileIDs {
					go func(i int) {
						defer waitGroup.Done()
						reader, err := updateHandler.driveFileService.DownloadOneByID(foundDriveFileIDs[i])
						if err != nil {
							return
						}

						driveFile, err := updateHandler.driveFileService.FindOneByID(foundDriveFileIDs[i])
						if err != nil {
							return
						}

						documents[i] = tgbotapi.NewInputMediaDocument(tgbotapi.FileReader{
							Name:   driveFile.Name + ".pdf",
							Reader: *reader,
						})
					}(i)
				}
				waitGroup.Wait()

				responses, err = updateHandler.bot.SendMediaGroup(tgbotapi.NewMediaGroup(update.Message.Chat.ID, documents))
				if err != nil {
					continue
				}
			}

			for j := range responses {
				foundDriveFileID := user.State.Context.FoundDriveFileIDs[j+(i*len(chunk))]

				song, err := updateHandler.songService.FindOneByDriveFileID(foundDriveFileID)
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

		for _, messageID := range user.State.Context.MessagesToDelete {
			msg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, messageID)
			updateHandler.bot.Send(msg)
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
