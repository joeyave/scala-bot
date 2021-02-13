package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scalaChordsBot/configs"
	"scalaChordsBot/entities"
)

func songVoicesHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		voices := user.CurrentState().Context.CurrentSong.Voices

		var err error
		if voices == nil || len(voices) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У этой песни нет партий. Чтобы добавить, отправь мне голосовое сообщение.")
			_, err = updateHandler.bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери партию:")

			keyboard := tgbotapi.NewReplyKeyboard()
			keyboard.OneTimeKeyboard = false
			keyboard.ResizeKeyboard = true

			for _, voice := range voices {
				keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(voice.Caption)))
			}

			msg.ReplyMarkup = keyboard

			_, err = updateHandler.bot.Send(msg)
		}

		user.CurrentState().NextIndex()
		return user, err
	})

	return configs.SongVoicesState, handleFuncs
}
