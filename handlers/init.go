package handlers

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/telebot/v3"
)

type HandlerFunc = func(h *Handler, c telebot.Context, user *entities.User) error

var handlers = make(map[string][]HandlerFunc, 0)

// Register your handlers here.
func init() {
	registerHandlers(
		mainMenuHandler,
		scheduleHandler,
		//addBandAdminHandler,
		createSongHandler,
		copySongHandler,
		//deleteSongHandler,
		createBandHandler,
		chooseBandHandler,
		styleSongHandler,
		searchSongHandler,
		setlistHandler,
		songActionsHandler,
		getVoicesHandler,
		uploadVoiceHandler,
		transposeSongHandler,
		getEventsHandler,
		createEventHandler,
		eventActionsHandler,
		createRoleHandler,
		addEventMemberHandler,
		addEventSongHandler,
	)
}

func registerHandlers(funcs ...func() (name string, funcs []HandlerFunc)) {
	for _, f := range funcs {
		name, hf := f()
		handlers[name] = hf
	}
}
