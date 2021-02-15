package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"sort"
)

func getVoicesHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		voices := user.State.Context.CurrentSong.Voices

		var err error
		if voices == nil || len(voices) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У этой песни нет партий. Чтобы добавить, отправь мне голосовое сообщение.")
			_, err = updateHandler.bot.Send(msg)

			user.State = user.State.Prev
			return user, err
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
			return user, err
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case helpers.Back:
			user.State = user.State.Prev
			user.State.Index = 0
			return updateHandler.enterStateHandler(update, user)
		default:
			voices := user.State.Context.CurrentSong.Voices

			sort.Slice(voices, func(i, j int) bool {
				return voices[i].Caption <= voices[j].Caption
			})

			foundIndex := sort.Search(len(voices), func(i int) bool {
				return voices[i].Caption >= update.Message.Text
			})

			if foundIndex != len(voices) {
				msg := tgbotapi.NewVoiceShare(update.Message.Chat.ID, voices[foundIndex].TgFileID)
				msg.Caption = voices[foundIndex].Caption
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: helpers.Delete}),
					tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: helpers.Back}),
				)
				keyboard.ResizeKeyboard = true
				msg.ReplyMarkup = keyboard

				_, err := updateHandler.bot.Send(msg)

				user.State.Index++
				return user, err
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет партии c таким названием. Попробуй еще раз.")
				_, err := updateHandler.bot.Send(msg)
				return user, err
			}
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case helpers.Back:
			user.State.Index = 0
			return updateHandler.enterStateHandler(update, user)
		case helpers.Delete:
			// TODO: handle delete
			return user, nil
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я тебя не понимаю. Нажми на кнопку.")
			_, err := updateHandler.bot.Send(msg)
			return user, err
		}
	})

	return helpers.GetVoicesState, handleFuncs
}

func uploadVoiceHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введи название песни, к которой ты хочешь прикрепить эту партию:")
		keyboard := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
		keyboard.ResizeKeyboard = true
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case helpers.Cancel:
			if user.State.Prev != nil {
				user.State = user.State.Prev
				user.State.Index = 0
			} else {
				user.State = &entities.State{
					Index: 0,
					Name:  helpers.MainMenuState,
				}
			}
		default:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SearchSongState,
				Prev:  user.State.Prev,
				Next: &entities.State{
					Index:   2,
					Name:    helpers.UploadVoiceState,
					Context: user.State.Context,
				},
			}
		}
		return updateHandler.enterStateHandler(update, user)
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправь мне название этой партии:")
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: helpers.Cancel}),
		)
		keyboard.ResizeKeyboard = true
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case "":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Название для партии должно быть текстом! Попробуй еще раз.")
			_, err := updateHandler.bot.Send(msg)

			return user, err

		case helpers.Cancel:
			if user.State.Prev != nil {
				user.State = user.State.Prev
				user.State.Index = 0
			} else {
				user.State = &entities.State{
					Index: 0,
					Name:  helpers.MainMenuState,
				}
			}
			return updateHandler.enterStateHandler(update, user)

		default:
			user.State.Context.CurrentVoice.Caption = update.Message.Text

			user.State.Context.CurrentSong.Voices =
				append(user.State.Context.CurrentSong.Voices, user.State.Context.CurrentVoice)
			_, err := updateHandler.songService.UpdateOne(*user.State.Context.CurrentSong)
			if err != nil {
				return user, err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добавление завершено.")
			_, err = updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Index:   0,
				Name:    helpers.SongActionsState,
				Context: entities.Context{CurrentSong: user.State.Context.CurrentSong},
			}
			return updateHandler.enterStateHandler(update, user)
		}
	})
	return helpers.UploadVoiceState, handleFuncs
}
