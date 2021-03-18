package handlers

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/scala-chords-bot/services"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"strings"
)

type UpdateHandler struct {
	bot         *tgbotapi.BotAPI
	userService *services.UserService
	songService *services.SongService
	bandService *services.BandService
}

func NewHandler(bot *tgbotapi.BotAPI, userService *services.UserService, songService *services.SongService, bandService *services.BandService) *UpdateHandler {
	return &UpdateHandler{
		bot:         bot,
		userService: userService,
		songService: songService,
		bandService: bandService,
	}
}

func (u *UpdateHandler) HandleUpdate(update *tgbotapi.Update) error {
	defer func() {
		if r := recover(); r != nil {
			helpers.LogError(update, u.bot, r)
		}
	}()

	user, err := u.userService.FindOneByID(update.Message.Chat.ID)
	if err != nil {
		user = &entities.User{
			ID:   update.Message.Chat.ID,
			Name: strings.TrimSpace(fmt.Sprintf("%s %s", update.Message.Chat.FirstName, update.Message.Chat.LastName)),
			State: &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			},
		}

		user, err = u.userService.CreateOne(*user)
		if err != nil {
			return err
		}
	}

	if user.Band == nil &&
		user.State.Name != helpers.ChooseBandState && user.State.Name != helpers.CreateBandState {
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.ChooseBandState,
		}

		user, err = u.userService.UpdateOne(*user)
		if err != nil {
			return err
		}
	}

	user.Name = strings.TrimSpace(fmt.Sprintf("%s %s", update.Message.Chat.FirstName, update.Message.Chat.LastName))

	backupUser := *user

	// Handle buttons.
	switch update.Message.Text {
	case helpers.Cancel:
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

	// Catch voice anywhere.
	if update.Message.Voice != nil {
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.UploadVoiceState,
			Context: entities.Context{
				CurrentVoice: &entities.Voice{
					TgFileID: update.Message.Voice.FileID,
				},
			},
			Prev: user.State,
		}
	}

	user, err = u.enterStateHandler(update, *user)
	if err != nil {
		backupUser.State = &entities.State{
			Index: 0,
			Name:  helpers.MainMenuState,
		}
		u.userService.UpdateOne(backupUser)

		return err
	}

	_, err = u.userService.UpdateOne(*user)
	return err
}

func (u *UpdateHandler) enterStateHandler(update *tgbotapi.Update, user entities.User) (*entities.User, error) {
	handleFuncs, ok := stateHandlers[user.State.Name]

	if ok == false || user.State.Index >= len(handleFuncs) || user.State.Index < 0 {
		user.State.Index = 0
		user.State.Name = helpers.MainMenuState
		handleFuncs = stateHandlers[user.State.Name]
	}

	return handleFuncs[user.State.Index](u, update, user)
}
