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
)

func searchSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	// Print list of found songs.
	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
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
				return user, err
			default:
				chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
				_, _ = updateHandler.bot.Send(chatAction)

				var driveFiles []*drive.File
				var err error
				if update.Message.Text == helpers.SearchEverywhere {
					driveFiles, _, err = updateHandler.songService.QueryDrive(user.State.Context.Query, "")
				} else {
					user.State.Context.Query = update.Message.Text
					driveFiles, _, err = updateHandler.songService.QueryDrive(update.Message.Text, "", user.GetFolderIDs()...)
				}

				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено. Попробуй еще раз.")
					keyboard := tgbotapi.NewReplyKeyboard()
					keyboard.Keyboard = append(keyboard.Keyboard,
						tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.SearchEverywhere)),
						tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
					msg.ReplyMarkup = keyboard
					_, err = updateHandler.bot.Send(msg)

					return user, err
				}

				keyboard := tgbotapi.NewReplyKeyboard()
				keyboard.OneTimeKeyboard = false
				keyboard.ResizeKeyboard = true

				// TODO: some sort of pagination.
				for _, song := range driveFiles {

					songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.Name))
					keyboard.Keyboard = append(keyboard.Keyboard, songButton)
				}

				keyboard.Keyboard = append(keyboard.Keyboard,
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.SearchEverywhere)),
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери песню:")
				msg.ReplyMarkup = keyboard
				_, _ = updateHandler.bot.Send(msg)

				user.State.Context.DriveFiles = driveFiles
				user.State.Index++
				return user, err
			}
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case "":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мне нужно название песни.")
			_, err := updateHandler.bot.Send(msg)
			user.State.Index--
			return user, err
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
				song, err := updateHandler.songService.FindOneByDriveFile(*driveFiles[foundIndex])
				if err != nil {
					return entities.User{}, err
				}

				if user.State.Next != nil {
					user.State = user.State.Next
					user.State.Context.CurrentSong = song
				} else {
					user.State = &entities.State{
						Index: 0,
						Name:  helpers.SongActionsState,
						Context: entities.Context{
							CurrentSong: song,
						},
					}
				}

				return updateHandler.enterStateHandler(update, user)
			} else {
				user.State = &entities.State{
					Index:   1,
					Name:    helpers.MainMenuState,
					Context: user.State.Context,
				}
				return updateHandler.enterStateHandler(update, user)
			}
		}
	})

	return helpers.SearchSongState, handleFuncs
}

func songActionsHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		//if user.HasAuthorityToEdit(cachedSong) == false {
		//	buttons := [][]tgbotapi.KeyboardButton{
		//		{{Text: foundSong.Name}},
		//		{{Text: helpers.CopyToMyBand}},
		//		{{Text: helpers.Voices}, {Text: helpers.Audios}},
		//		{{Text: helpers.Menu}},
		//	}
		//	keyboard.Keyboard = append(keyboard.Keyboard, buttons...)
		//} else {
		//	buttons := helpers.GetSongOptionsKeyboard()
		//	buttons = append([][]tgbotapi.KeyboardButton{{{Text: foundSong.Name}}}, buttons...)
		//	keyboard.Keyboard = append(keyboard.Keyboard, buttons...)
		//}

		song := user.State.Context.CurrentSong
		song, err := updateHandler.songService.FindOneByDriveFile(*song.DriveFile)
		if err != nil {
			return entities.User{}, err
		}

		var msg tgbotapi.DocumentConfig

		if song.HasOutdatedPDF() {
			fileReader, err := updateHandler.songService.DownloadPDF(*song.DriveFile)
			if err != nil {
				return entities.User{}, err
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

		res, err := updateHandler.bot.Send(msg)
		if err != nil {
			fileReader, err := updateHandler.songService.DownloadPDF(*song.DriveFile)
			if err != nil {
				return entities.User{}, err
			}

			msg = tgbotapi.NewDocument(update.Message.Chat.ID, fileReader)

			res, err = updateHandler.bot.Send(msg)
			if err != nil {
				return entities.User{}, err
			}
		}

		song.PDF = &entities.PDF{
			TgFileID:     res.Document.FileID,
			ModifiedTime: song.DriveFile.ModifiedTime,
		}

		song, err = updateHandler.songService.UpdateOne(*song)
		if err != nil {
			return user, fmt.Errorf("failed to cache file %v", err)
		}

		user.State.Index++
		user.State.Context.CurrentSong = song

		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case helpers.Menu:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			}
			return updateHandler.enterStateHandler(update, user)

		case helpers.Voices:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.GetVoicesState,
				Context: entities.Context{
					CurrentSong: user.State.Context.CurrentSong,
				},
				Prev: user.State,
			}
			return updateHandler.enterStateHandler(update, user)

		case helpers.Transpose:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.TransposeSongState,
				Context: entities.Context{
					CurrentSong: user.State.Context.CurrentSong,
				},
				Prev: user.State,
			}
			return updateHandler.enterStateHandler(update, user)

		case helpers.Style:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.StyleSongState,
				Context: entities.Context{
					CurrentSong: user.State.Context.CurrentSong,
				},
				Prev: user.State,
			}
			return updateHandler.enterStateHandler(update, user)

		case helpers.CopyToMyBand:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.CopySongState,
				Context: entities.Context{
					CurrentSong: user.State.Context.CurrentSong,
				},
				Prev: user.State,
			}
			return updateHandler.enterStateHandler(update, user)

		case user.State.Context.CurrentSong.DriveFile.Name:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, user.State.Context.CurrentSong.DriveFile.WebViewLink)
			_, err := updateHandler.bot.Send(msg)
			return user, err

		default:
			user.State = &entities.State{
				Index: 1,
				Name:  helpers.MainMenuState,
			}
			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.SongActionsState, handleFuncs
}

func transposeSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери новую тональность:")
		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.ResizeKeyboard = true
		keyboard.Keyboard = append(keyboard.Keyboard, helpers.KeysKeyboard...)
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
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

			sections, err := updateHandler.songService.GetSections(*user.State.Context.CurrentSong)
			if err != nil {
				return user, err
			}

			for i, _ := range sections {
				keyboard.Keyboard = append(keyboard.Keyboard,
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("Вместо %d-й секции", i+1))))
			}
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))

			msg.ReplyMarkup = keyboard
			_, err = updateHandler.bot.Send(msg)

			user.State.Index++
			return user, err
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case "":
			if user.State.Index > 0 {
				user.State.Index--
			}
			return updateHandler.enterStateHandler(update, user)
		default:
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
			_, _ = updateHandler.bot.Send(chatAction)

			sections, err := updateHandler.songService.GetSections(*user.State.Context.CurrentSong)
			if err != nil {
				return user, err
			}

			re := regexp.MustCompile("[1-9]+")
			sectionIndex, err := strconv.Atoi(re.FindString(update.Message.Text))
			if err != nil {
				sections, err = updateHandler.songService.AppendSection(*user.State.Context.CurrentSong)
				if err != nil {
					return user, err
				}

				sectionIndex = len(sections) - 1
			} else {
				sectionIndex--
			}

			if sectionIndex >= len(sections) {
				user.State.Index--
				return updateHandler.enterStateHandler(update, user)
			}

			song, err := updateHandler.songService.Transpose(*user.State.Context.CurrentSong,
				user.State.Context.Key, sectionIndex)
			if err != nil {
				return entities.User{}, err
			}

			user.State = &entities.State{
				Index:   0,
				Name:    helpers.SongActionsState,
				Context: entities.Context{CurrentSong: &song},
			}

			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.TransposeSongState, handleFuncs
}

func styleSongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	// Print list of found songs.
	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		_, _ = updateHandler.bot.Send(chatAction)

		song, err := updateHandler.songService.Style(*user.State.Context.CurrentSong)

		if err != nil {
			return entities.User{}, err
		}

		user.State = &entities.State{
			Index:   0,
			Name:    helpers.SongActionsState,
			Context: entities.Context{CurrentSong: &song},
		}

		return updateHandler.enterStateHandler(update, user)
	})
	return helpers.StyleSongState, handleFuncs
}

func copySongHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
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
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
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
				copiedSong, err := updateHandler.songService.DeepCopyToFolder(*user.State.Context.CurrentSong, user.Bands[foundIndex].DriveFolderID)
				if err != nil {
					return entities.User{}, err
				}

				user.State = &entities.State{
					Index:   0,
					Name:    helpers.SongActionsState,
					Context: entities.Context{CurrentSong: copiedSong},
				}

				return updateHandler.enterStateHandler(update, user)

			} else {
				user.State.Index--
			}

			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.CopySongState, handleFuncs
}
