package handlers

import (
	"fmt"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
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

				songs, _, err := updateHandler.songService.FindByName(update.Message.Text, "")
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено. Попробуй еще раз.")
					msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
					_, err = updateHandler.bot.Send(msg)

					return user, err
				}

				keyboard := tgbotapi.NewReplyKeyboard()
				keyboard.OneTimeKeyboard = false
				keyboard.ResizeKeyboard = true

				// TODO: some sort of pagination.
				for _, song := range songs {

					songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.Name))
					keyboard.Keyboard = append(keyboard.Keyboard, songButton)
				}

				keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери песню:")
				msg.ReplyMarkup = keyboard
				_, _ = updateHandler.bot.Send(msg)

				user.State.Context.Songs = songs
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

			songs := user.State.Context.Songs
			foundIndex := len(songs)
			for i := range songs {
				if songs[i].Name == update.Message.Text {
					foundIndex = i
					break
				}
			}

			if foundIndex != len(songs) {
				if user.State.Next != nil {
					user.State = user.State.Next
					song, err := updateHandler.songService.GetWithActualTgFileID(songs[foundIndex])
					if err != nil {
						user.State.Context.CurrentSong = &songs[foundIndex]
					} else {
						user.State.Context.CurrentSong = &song
					}
				} else {
					user.State = &entities.State{
						Index: 0,
						Name:  helpers.SongActionsState,
						Context: entities.Context{
							CurrentSong: &songs[foundIndex],
						},
					}
				}

				return updateHandler.enterStateHandler(update, user)
			} else {
				user.State.Index = 0
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

		foundSong := *user.State.Context.CurrentSong

		cachedSong, err := updateHandler.songService.GetWithActualTgFileID(foundSong)

		keyboard := helpers.GetSongOptionsKeyboard()
		keyboard = append([][]tgbotapi.KeyboardButton{{{Text: foundSong.Name}}}, keyboard...)

		if err != nil { // Song not found in cache - upload from my server.
			err = nil

			fileReader, err := updateHandler.songService.DownloadPDF(foundSong)
			if err != nil {
				return user, err
			}

			msg := tgbotapi.NewDocument(update.Message.Chat.ID, fileReader)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

			res, err := updateHandler.bot.Send(msg)
			if err != nil {
				return user, fmt.Errorf("failed to send file %v", err)
			}

			foundSong.TgFileID = res.Document.FileID
			cachedSong, err = updateHandler.songService.Cache(foundSong)
			if err != nil {
				return user, fmt.Errorf("failed to cache file %v", err)
			}
		} else { // Found in cache.
			msg := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FileID(cachedSong.TgFileID))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

			_, err := updateHandler.bot.Send(msg)
			if err != nil {
				fileReader, err := updateHandler.songService.DownloadPDF(foundSong)
				if err != nil {
					return user, err
				}

				msg := tgbotapi.NewDocument(update.Message.Chat.ID, fileReader)
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

				res, err := updateHandler.bot.Send(msg)
				if err != nil {
					return user, fmt.Errorf("failed to send file %v", err)
				}

				foundSong.TgFileID = res.Document.FileID
				cachedSong, err = updateHandler.songService.Cache(foundSong)
				if err != nil {
					return user, fmt.Errorf("failed to cache file %v", err)
				}
			}
		}

		user.State.Index++
		user.State.Context.CurrentSong = &cachedSong

		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case user.State.Context.CurrentSong.Name:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, user.State.Context.CurrentSong.WebViewLink)
			_, err := updateHandler.bot.Send(msg)
			return user, err

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
