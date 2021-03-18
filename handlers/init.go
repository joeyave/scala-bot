package handlers

import (
	"github.com/joeyave/scala-chords-bot/entities"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
)

var stateHandlers = make(map[string][]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

// Register your handlers here.
func init() {
	name, funcs := mainMenuHandler()
	stateHandlers[name] = funcs

	name, funcs = addBandAdminHandler()
	stateHandlers[name] = funcs

	name, funcs = createSongHandler()
	stateHandlers[name] = funcs

	name, funcs = copySongHandler()
	stateHandlers[name] = funcs

	name, funcs = deleteSongHandler()
	stateHandlers[name] = funcs

	name, funcs = scheduleHandler()
	stateHandlers[name] = funcs

	name, funcs = createBandHandler()
	stateHandlers[name] = funcs

	name, funcs = chooseBandHandler()
	stateHandlers[name] = funcs

	name, funcs = styleSongHandler()
	stateHandlers[name] = funcs

	name, funcs = searchSongHandler()
	stateHandlers[name] = funcs

	name, funcs = setlistHandler()
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
