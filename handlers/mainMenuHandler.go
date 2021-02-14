package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scala-chords-bot/entities"
	"scala-chords-bot/helpers"
)

func mainMenuHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Основное меню:")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(helpers.MainMenuKeyboard...)
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case "":
			_, err := updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Действие не поддерживается."))
			return user, err
		case helpers.Help:
			_, err := updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Для поиска документа, отправь боту название.\n\nРедактировать документ можно на гугл диске. Теперь не нужно отправлять файл боту, он сам обновит его.\n\nДля добавления партии, отправь боту голосовое сообщение."))
			return user, err
		default:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SearchSongState,
				Prev:  user.State,
			}

			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.MainMenuState, handleFuncs
}
