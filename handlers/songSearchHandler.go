package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scala-chords-bot/configs"
	"scala-chords-bot/entities"
	"sort"
)

func songSearchHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	// Print list of found songs.
	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		{
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			_, _ = updateHandler.bot.Send(chatAction)

			songs, err := updateHandler.SongService.FindByName(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено. Попробуй еще раз.")
				//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				_, err = updateHandler.bot.Send(msg)

				return user, err
			}

			songsKeyboard := tgbotapi.NewReplyKeyboard()
			songsKeyboard.OneTimeKeyboard = false
			songsKeyboard.ResizeKeyboard = true

			// TODO: some sort of pagination.
			const pageSize = 100
			for i, song := range songs {
				if i == pageSize {
					break
				}

				songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.Name))
				songsKeyboard.Keyboard = append(songsKeyboard.Keyboard, songButton)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери песню:")
			msg.ReplyMarkup = songsKeyboard
			_, _ = updateHandler.bot.Send(msg)

			user.State.Context.Songs = songs
			user.State.Index++
			return user, err
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		songs := user.State.Context.Songs

		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		sort.Slice(songs, func(i, j int) bool {
			return songs[i].Name <= songs[j].Name
		})

		foundIndex := sort.Search(len(songs), func(i int) bool {
			return songs[i].Name >= update.Message.Text
		})

		if foundIndex != len(songs) {
			if user.State.Next != nil {
				user.State = user.State.Next
				user.State.Context.CurrentSong = &songs[foundIndex]
			} else {
				user.State = &entities.State{
					Index:   0,
					Name:    configs.SongActionsState,
					Context: entities.Context{CurrentSong: &songs[foundIndex]},
				}
			}

			return updateHandler.enterStateHandler(update, user)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено.")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			_, err := updateHandler.bot.Send(msg)
			if err != nil {
				return user, err
			}

			user.State.Index--
		}

		return user, nil
	})

	return configs.SongSearchState, handleFuncs
}
