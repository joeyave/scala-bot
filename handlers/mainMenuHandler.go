package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scala-chords-bot/configs"
	"scala-chords-bot/entities"
)

func mainMenuHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Основное меню:")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(configs.MainMenuKeyboard...)
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		if update.Message.Text != "" {
			user.State = &entities.State{Index: 0, Name: configs.SongSearchState, Prev: user.State}

			return updateHandler.enterStateHandler(update, user)
		} else {
			_, err := updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Действие не поддерживается."))
			return user, err
		}
	})

	return configs.MainMenuState, handleFuncs
}
