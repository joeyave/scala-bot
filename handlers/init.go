package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scalaChordsBot/entities"
)

var stateHandlers = make(map[string][]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

func init() {
	name, funcs := songSearchHandler()
	stateHandlers[name] = funcs

	name, funcs = songActionsHandler()
	stateHandlers[name] = funcs

	name, funcs = songVoicesHandler()
	stateHandlers[name] = funcs
}
