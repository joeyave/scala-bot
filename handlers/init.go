package handlers

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/telebot/v3"
)

type HandlerFunc = func(h *Handler, c telebot.Context, user *entities.User) error

var handlers = make(map[int][]HandlerFunc, 0)

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
		deleteEventHandler,
		changeSongOrderHandler,
		addEventMemberHandler,
		deleteEventMemberHandler,
	)
}

func registerHandlers(funcs ...func() (name int, funcs []HandlerFunc)) {
	for _, f := range funcs {
		name, hf := f()
		handlers[name] = hf
	}
}
