package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scalaChordsBot/configs"
	"scalaChordsBot/entities"
)

func songActionsHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		foundSong := user.CurrentState().Context.CurrentSong

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

		user.CurrentState().NextIndex()
		user.CurrentState().Context.CurrentSong = cachedSong

		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		var err error

		switch update.Message.Text {
		case user.CurrentState().Context.CurrentSong.Name:
			_, err = updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, user.CurrentState().Context.CurrentSong.WebViewLink))

		case configs.Menu:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Основное меню:")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(configs.MainMenuKeyboard...)
			_, err = updateHandler.bot.Send(msg)

			user.CurrentState().ChangeTo(configs.SongSearchState)

		case configs.Voices:
			user.AppendState(configs.SongVoicesState, entities.Context{
				CurrentSong: user.CurrentState().Context.CurrentSong,
			})
			return enterStateHandler(updateHandler, update, user)

		default:
			user.CurrentState().ChangeTo(configs.SongSearchState)
			return enterStateHandler(updateHandler, update, user)
		}

		return user, err
	})

	return configs.SongActionsState, handleFuncs
}
