package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scala-chords-bot/configs"
	"scala-chords-bot/entities"
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
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(configs.Back)))

			msg.ReplyMarkup = keyboard

			_, err = updateHandler.bot.Send(msg)
			user.State.Index++
			return user, err
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case configs.Back:
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
					tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: configs.Delete}),
					tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: configs.Back}),
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
		case configs.Back:
			user.State.Index = 0
			return updateHandler.enterStateHandler(update, user)
		case configs.Delete:
			return user, nil
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я тебя не понимаю. Нажми на кнопку.")
			_, err := updateHandler.bot.Send(msg)
			return user, err
		}
	})

	return configs.GetVoicesState, handleFuncs
}

func uploadVoiceHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введи название песни, к которой ты хочешь прикрепить эту партию:")
		keyboard := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(configs.Cancel)))
		keyboard.ResizeKeyboard = true
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case configs.Cancel:
			user.State = user.State.Prev
			user.State.Index = 0
		default:
			user.State = &entities.State{
				Index: 0,
				Name:  configs.SongSearchState,
				Prev:  user.State,
				Next: &entities.State{
					Index:   2,
					Name:    configs.UploadVoiceState,
					Context: user.State.Context,
				},
			}
		}
		return updateHandler.enterStateHandler(update, user)
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправь мне название этой партии:")
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		if update.Message.Text != "" {
			user.State.Context.CurrentVoice.Caption = update.Message.Text

			user.State.Context.CurrentSong.Voices =
				append(user.State.Context.CurrentSong.Voices, user.State.Context.CurrentVoice)
			_, err := updateHandler.SongService.UpdateOne(*user.State.Context.CurrentSong)
			if err != nil {
				return user, err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добавление завершено.")
			_, err = updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Index:   0,
				Name:    configs.SongActionsState,
				Context: entities.Context{CurrentSong: user.State.Context.CurrentSong},
			}

			return updateHandler.enterStateHandler(update, user)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Название для партии должно быть текстом! Попробуй еще раз.")
			_, err := updateHandler.bot.Send(msg)

			return user, err
		}
	})
	return configs.UploadVoiceState, handleFuncs
}
