package handlers

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"gopkg.in/tucnak/telebot.v3"
)

type HandlerFunc = func(h *Handler, c telebot.Context, user *entities.User) error

var handlers = make(map[int][]HandlerFunc, 0)

// Register your handlers here.
func init() {
	registerHandlers(
		mainMenuHandler,
		// addBandAdminHandler,
		createSongHandler,
		copySongHandler,
		// deleteSongHandler,
		createBandHandler,
		chooseBandHandler,
		styleSongHandler,
		changeSongBPMHandler,
		searchSongHandler,
		setlistHandler,
		songActionsHandler,
		getVoicesHandler,
		uploadVoiceHandler,
		transposeSongHandler,
		getEventsHandler,
		createEventHandler,
		eventActionsHandler,
		settingsHandler,
		createRoleHandler,
		addEventMemberHandler,
		addEventSongHandler,
		changeEventNotesHandler,
		deleteEventHandler,
		changeSongOrderHandler,
		deleteEventMemberHandler,
		deleteEventSongHandler,
		addBandAdminHandler,
		deleteSongHandler,
		getSongsFromMongoHandler,
		changeEventDateHandler,
		editInlineKeyboardHandler,
	)
}

func registerHandlers(funcs ...func() (name int, funcs []HandlerFunc)) {
	for _, f := range funcs {
		name, hf := f()
		handlers[name] = hf
	}
}
