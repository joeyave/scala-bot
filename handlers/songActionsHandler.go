package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scala-chords-bot/configs"
	"scala-chords-bot/entities"
)

func songActionsHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		foundSong := *user.State.Context.CurrentSong

		cachedSong, err := updateHandler.SongService.GetWithActualTgFileID(foundSong)

		keyboard := configs.GetSongOptionsKeyboard()
		keyboard = append([][]tgbotapi.KeyboardButton{{{Text: foundSong.Name}}}, keyboard...)

		if err != nil { // Song not found in cache - upload from my server.
			err = nil

			fileReader, err := updateHandler.SongService.DownloadPDF(foundSong)
			if err != nil {
				return user, err
			}

			msg := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, *fileReader)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

			res, err := updateHandler.bot.Send(msg)
			if err != nil {
				return user, fmt.Errorf("failed to send file %v", err)
			}

			foundSong.TgFileID = res.Document.FileID
			cachedSong, err = updateHandler.SongService.UpdateOne(foundSong)
			if err != nil {
				return user, fmt.Errorf("failed to cache file %v", err)
			}
		} else { // Found in cache.
			msg := tgbotapi.NewDocumentShare(update.Message.Chat.ID, cachedSong.TgFileID)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

			_, err := updateHandler.bot.Send(msg)
			if err != nil {
				fileReader, err := updateHandler.SongService.DownloadPDF(foundSong)
				if err != nil {
					return user, err
				}

				msg := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, *fileReader)
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

				res, err := updateHandler.bot.Send(msg)
				if err != nil {
					return user, fmt.Errorf("failed to send file %v", err)
				}

				foundSong.TgFileID = res.Document.FileID
				cachedSong, err = updateHandler.SongService.UpdateOne(foundSong)
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
		var err error

		switch update.Message.Text {
		case user.State.Context.CurrentSong.Name:
			_, err = updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, user.State.Context.CurrentSong.WebViewLink))

		case configs.Menu:
			user.State = &entities.State{
				Index: 0,
				Name:  configs.MainMenuState,
			}
			return updateHandler.enterStateHandler(update, user)

		case configs.Voices:
			user.State = &entities.State{
				Index: 0,
				Name:  configs.GetVoicesState,
				Context: entities.Context{
					CurrentSong: user.State.Context.CurrentSong,
				},
				Prev: user.State,
			}
			return updateHandler.enterStateHandler(update, user)

		default:
			user.State = &entities.State{
				Index: 0,
				Name:  configs.SongSearchState,
			}
			return updateHandler.enterStateHandler(update, user)
		}

		return user, err
	})

	return configs.SongActionsState, handleFuncs
}
