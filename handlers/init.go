package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scala-chords-bot/entities"
)

var stateHandlers = make(map[string][]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

// Register your handlers here.
func init() {
	name, funcs := mainMenuHandler()
	stateHandlers[name] = funcs

	name, funcs = songSearchHandler()
	stateHandlers[name] = funcs

	name, funcs = songActionsHandler()
	stateHandlers[name] = funcs

	name, funcs = getVoicesHandler()
	stateHandlers[name] = funcs

	name, funcs = uploadVoiceHandler()
	stateHandlers[name] = funcs
}
