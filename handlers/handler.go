package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/scala-chords-bot/services"
	"github.com/joeyave/telebot/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type Handler struct {
	bot               *telebot.Bot
	userService       *services.UserService
	driveFileService  *services.DriveFileService
	songService       *services.SongService
	voiceService      *services.VoiceService
	bandService       *services.BandService
	membershipService *services.MembershipService
	eventService      *services.EventService
	roleService       *services.RoleService
}

func NewHandler(
	bot *telebot.Bot,
	userService *services.UserService,
	driveFileService *services.DriveFileService,
	songService *services.SongService,
	voiceService *services.VoiceService,
	bandService *services.BandService,
	membershipService *services.MembershipService,
	eventService *services.EventService,
	roleService *services.RoleService,
) *Handler {

	return &Handler{
		bot:               bot,
		userService:       userService,
		driveFileService:  driveFileService,
		songService:       songService,
		voiceService:      voiceService,
		bandService:       bandService,
		membershipService: membershipService,
		eventService:      eventService,
		roleService:       roleService,
	}
}

func (h *Handler) OnText(c telebot.Context) error {

	user, err := h.userService.FindOneByID(c.Chat().ID)
	if err != nil {
		return err
	}

	// Handle buttons.
	switch c.Text() {
	case helpers.Cancel, helpers.Back:
		if user.State.Prev != nil {
			user.State = user.State.Prev
			user.State.Index = 0
		} else {
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			}
		}

	case helpers.Menu:
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.MainMenuState,
		}
	}

	err = h.enter(c, user)
	if err != nil {
		return err
	}

	_, err = h.userService.UpdateOne(*user)

	return err
}

func (h *Handler) OnVoice(c telebot.Context) error {

	user, err := h.userService.FindOneByID(c.Chat().ID)
	if err != nil {
		return err
	}

	user.State = &entities.State{
		Index: 0,
		Name:  helpers.UploadVoiceState,
		Context: entities.Context{
			Voice: &entities.Voice{
				FileID: c.Message().Voice.FileID,
			},
		},
		Prev: user.State,
	}

	err = h.enter(c, user)
	if err != nil {
		return err
	}

	_, err = h.userService.UpdateOne(*user)

	return err
}

func (h *Handler) OnError(botErr error, c telebot.Context) {
	c.Send("Произошла ошибка. Поправим.")

	user, err := h.userService.FindOneByID(c.Chat().ID)
	if err != nil {
		h.bot.Send(telebot.ChatID(helpers.LogsChannelID), fmt.Sprintf("<code>%v</code>", botErr), telebot.ModeHTML)
		return
	}

	bytes, _ := json.MarshalIndent(user, "", "\t")

	h.bot.Send(telebot.ChatID(helpers.LogsChannelID), fmt.Sprintf("<code>%v</code>\n\n<code>%v</code>", botErr, string(bytes)), telebot.ModeHTML)

	user.State = &entities.State{Name: helpers.MainMenuState}
	_, err = h.userService.UpdateOne(*user)
	if err != nil {
		return
	}
}

func (h *Handler) RegisterUserMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		user, err := h.userService.FindOneByID(c.Chat().ID)
		if err != nil {
			user = &entities.User{
				ID: c.Chat().ID,
				State: &entities.State{
					Index: 0,
					Name:  helpers.MainMenuState,
				},
			}
		}

		if user.Name == "" {
			user.Name = strings.TrimSpace(fmt.Sprintf("%s %s", c.Chat().FirstName, c.Chat().LastName))
		}

		if user.BandID == primitive.NilObjectID && user.State.Name != helpers.ChooseBandState && user.State.Name != helpers.CreateBandState {
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.ChooseBandState,
			}
		}

		_, err = h.userService.UpdateOne(*user)
		return next(c)
	}
}

func (h *Handler) enter(c telebot.Context, user *entities.User) error {
	handlerFuncs, ok := handlers[user.State.Name]

	if ok == false || user.State.Index < 0 || user.State.Index >= len(handlerFuncs) {
		user.State = &entities.State{Name: helpers.MainMenuState}
		handlerFuncs = handlers[user.State.Name]
	}

	return handlerFuncs[user.State.Index](h, c, user)
}
