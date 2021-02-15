package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joeyave/scala-chords-bot/entities"
)

var stateHandlers = make(map[string][]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

// Register your handlers here.
func init() {
	name, funcs := mainMenuHandler()
	stateHandlers[name] = funcs

	name, funcs = searchSongHandler()
	stateHandlers[name] = funcs

	name, funcs = songActionsHandler()
	stateHandlers[name] = funcs

	name, funcs = getVoicesHandler()
	stateHandlers[name] = funcs

	name, funcs = uploadVoiceHandler()
	stateHandlers[name] = funcs

	name, funcs = transposeSongHandler()
	stateHandlers[name] = funcs
}
