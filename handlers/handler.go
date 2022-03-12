package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/joeyave/scala-bot/entities"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/joeyave/scala-bot/services"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/telebot.v3"
	"net/url"
	"regexp"
	"strings"
	"time"
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

	//user, err := h.userService.FindOneByID(c.Chat().ID)
	//if err != nil {
	//	return err
	//}

	user, ok := c.Get("user").(*entities.User)
	if !ok {
		return errors.New("error getting user from context")
	}

	// Handle buttons.
	switch c.Text() {
	case helpers.Cancel, helpers.Back:

		// user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, c.Message().ID)
		// for _, messageID := range user.State.Context.MessagesToDelete {
		// 	h.bot.Delete(&telebot.Message{
		// 		ID:   messageID,
		// 		Chat: c.Chat(),
		// 	})
		// }
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

	err := h.enter(c, user)
	if err != nil {
		return err
	}

	_, err = h.userService.UpdateOne(*user)

	return err
}

func (h *Handler) OnVoice(c telebot.Context) error {

	user, ok := c.Get("user").(*entities.User)
	if !ok {
		return errors.New("error getting user from context")
	}

	if user.State.Name != helpers.UploadVoiceState {
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.UploadVoiceState,
			Context: entities.Context{
				Voice: &entities.Voice{
					FileID: c.Message().Media().MediaFile().FileID,
				},
			},
			Prev: user.State,
		}
	}

	err := h.enter(c, user)
	if err != nil {
		return err
	}

	_, err = h.userService.UpdateOne(*user)

	return err
}

func (h *Handler) OnCallback(c telebot.Context) error {
	user, ok := c.Get("user").(*entities.User)
	if !ok {
		return errors.New("error getting user from context")
	}
	err := h.enter(c, user)
	if err != nil {
		return err
	}

	_, err = h.userService.UpdateOne(*user)

	return nil
}

func (h *Handler) OnError(botErr error, c telebot.Context) {
	c.Send("Произошла ошибка. Поправим.")

	requestID := c.Get("requestId").(string)

	user, err := h.userService.FindOneByID(c.Chat().ID)
	if err != nil {
		log.Error().Err(botErr).Str("requestId", requestID).Msg("Error!")
		return
	}

	log.Error().Err(botErr).Str("requestId", requestID).Msg("Error!")

	user.State = &entities.State{Name: helpers.MainMenuState}
	_, err = h.userService.UpdateOne(*user)
	if err != nil {
		log.Error().Err(botErr).Str("requestId", requestID).Msg("Error!")
		return
	}
}

func (h *Handler) RegisterUserMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {

		start := time.Now()

		requestID := uuid.New().String()
		c.Set("requestId", requestID)

		user, err := h.userService.FindOneOrCreateByID(c.Chat().ID)
		if err != nil {
			return err
		}

		update := c.Update()

		updateForLogs := helpers.Update{
			Update:   &update,
			Message:  helpers.MapMessage(update.Message),
			Callback: helpers.MapCallback(update.Callback),
		}

		updateBytes, _ := json.Marshal(updateForLogs)
		userBytes, _ := json.Marshal(user)

		log.Info().
			Str("requestId", c.Get("requestId").(string)).
			RawJSON("update", updateBytes).
			RawJSON("user", userBytes).
			Msg("Input:")

		// if user.Name == "" {
		// }
		user.Name = strings.TrimSpace(fmt.Sprintf("%s %s", c.Chat().FirstName, c.Chat().LastName))

		if user.BandID == primitive.NilObjectID && user.State.Name != helpers.ChooseBandState && user.State.Name != helpers.CreateBandState {
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.ChooseBandState,
			}
		}

		c.Set("user", user)

		//_, err = h.userService.UpdateOne(*user)

		err = next(c)
		err = errors.New("test err")
		if err != nil {
			return err
		}

		user, err = h.userService.FindOneByID(user.ID)
		if err != nil {
			log.Error().Err(err).Msg("error getting user for output log")
		}

		userBytes, _ = json.Marshal(user)

		log.Info().
			Str("requestId", c.Get("requestId").(string)).
			RawJSON("user", userBytes).
			Str("latency", time.Since(start).String()).
			Msg("Output:")

		return nil
	}
}

func (h *Handler) NotifyUser() {
	for range time.Tick(time.Hour * 2) {
		events, err := h.eventService.FindAllFromToday()
		if err != nil {
			return
		}

		for _, event := range events {
			if event.Time.Add(time.Hour*8).Sub(time.Now()).Hours() < 48 {
				for _, membership := range event.Memberships {
					if membership.Notified == true {
						continue
					}

					eventString := h.eventService.ToHtmlStringByEvent(*event)
					_, err := h.bot.Send(telebot.ChatID(membership.UserID),
						fmt.Sprintf("Привет. Ты учавствуешь в собрании через несколько дней! "+
							"Вот план:\n\n%s", eventString), telebot.ModeHTML, telebot.NoPreview)
					if err != nil {
						continue
					}

					membership.Notified = true
					h.membershipService.UpdateOne(*membership)
				}
			}
		}
	}
}

func (h *Handler) enter(c telebot.Context, user *entities.User) error {

	if user.State.CallbackData == nil {
		user.State.CallbackData, _ = url.Parse("t.me/callbackData")
	}

	if c.Callback() != nil {
		return h.enterInlineHandler(c, user)
	} else {
		return h.enterReplyHandler(c, user)
	}
}

func (h *Handler) enterInlineHandler(c telebot.Context, user *entities.User) error {

	re := regexp.MustCompile(`t\.me/callbackData.*`)

	for _, entity := range c.Callback().Message.CaptionEntities {
		if entity.Type == telebot.EntityTextLink {
			matches := re.FindStringSubmatch(entity.URL)

			if len(matches) > 0 {
				u, err := url.Parse(matches[0])
				if err != nil {
					return err
				}

				user.State.CallbackData = u
				break
			}
		}
	}

	for _, entity := range c.Callback().Message.Entities {
		if entity.Type == telebot.EntityTextLink {
			matches := re.FindStringSubmatch(entity.URL)

			if len(matches) > 0 {
				u, err := url.Parse(matches[0])
				if err != nil {
					return err
				}

				user.State.CallbackData = u
				break
			}
		}
	}

	state, index, _ := helpers.ParseCallbackData(c.Callback().Data)

	// Handle error.
	handlerFuncs, _ := handlers[state]

	return handlerFuncs[index](h, c, user)
}

func (h *Handler) enterReplyHandler(c telebot.Context, user *entities.User) error {
	handlerFuncs, ok := handlers[user.State.Name]

	if ok == false || user.State.Index < 0 || user.State.Index >= len(handlerFuncs) {
		user.State = &entities.State{Name: helpers.MainMenuState}
		handlerFuncs = handlers[user.State.Name]
	}

	return handlerFuncs[user.State.Index](h, c, user)
}
