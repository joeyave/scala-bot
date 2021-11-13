package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/joeyave/scala-bot/entities"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/joeyave/scala-bot/services"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/tucnak/telebot.v3"
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

	user, err := h.userService.FindOneByID(c.Chat().ID)
	if err != nil {
		return err
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

	if user.State.Name != helpers.UploadVoiceState {
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.UploadVoiceState,
			Context: entities.Context{
				Voice: &entities.Voice{
					FileID: c.Media().MediaFile().FileID,
				},
			},
			Prev: user.State,
		}
	}

	err = h.enter(c, user)
	if err != nil {
		return err
	}

	_, err = h.userService.UpdateOne(*user)

	return err
}

func (h *Handler) OnCallback(c telebot.Context) error {
	user, err := h.userService.FindOneByID(c.Chat().ID)
	if err != nil {
		return err
	}

	err = h.enter(c, user)
	if err != nil {
		return err
	}

	_, err = h.userService.UpdateOne(*user)

	return nil
}

func (h *Handler) OnError(botErr error, c telebot.Context) {
	c.Send("Произошла ошибка. Поправим.")
	log.Error().Err(botErr).Msgf("Bot error:")

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

func (h *Handler) LogInputMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		messageBytes, err := json.Marshal(c.Message())
		if err == nil {
			log.Info().RawJSON("msg", messageBytes).Msg("Input:")
		}

		return next(c)
	}
}

func (h *Handler) RegisterUserMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		start := time.Now()
		user, err := h.userService.FindOneOrCreateByID(c.Chat().ID)
		if err != nil {
			return err
		}
		log.Printf("getting user took %v", time.Since(start))

		// if user.Name == "" {
		// }
		user.Name = strings.TrimSpace(fmt.Sprintf("%s %s", c.Chat().FirstName, c.Chat().LastName))

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
