package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scalaChordsBot/configs"
	"scalaChordsBot/entities"
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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено.")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				_, _ = updateHandler.bot.Send(msg)
				return entities.User{}, fmt.Errorf("couldn't find Song %v", err)
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

			user.CurrentState().Context.Songs = songs
			user.CurrentState().NextIndex()
			return user, err
		}
	})

	//
	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		songs := user.CurrentState().Context.Songs

		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		sort.Slice(songs, func(i, j int) bool {
			return songs[i].Name <= songs[j].Name
		})

		foundIndex := sort.Search(len(songs), func(i int) bool {
			return songs[i].Name >= update.Message.Text
		})

		if foundIndex != len(songs) {
			user.CurrentState().ChangeTo(configs.SongActionsState)
			user.CurrentState().Context.CurrentSong = songs[foundIndex]
			return enterStateHandler(updateHandler, update, user)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено.")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			_, err := updateHandler.bot.Send(msg)
			if err != nil {
				return user, err
			}

			user.CurrentState().PrevIndex()
		}

		return user, nil
	})

	return configs.SongSearchState, handleFuncs
}
