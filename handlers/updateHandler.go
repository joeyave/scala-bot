package handlers

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/scala-chords-bot/services"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type UpdateHandler struct {
	bot              *tgbotapi.BotAPI
	userService      *services.UserService
	driveFileService *services.DriveFileService
	songService      *services.SongService
	voiceService     *services.VoiceService
	bandService      *services.BandService
}

func NewHandler(bot *tgbotapi.BotAPI, userService *services.UserService, driveFileService *services.DriveFileService, songService *services.SongService, voiceService *services.VoiceService, bandService *services.BandService) *UpdateHandler {
	return &UpdateHandler{
		bot:              bot,
		userService:      userService,
		driveFileService: driveFileService,
		songService:      songService,
		voiceService:     voiceService,
		bandService:      bandService,
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
			ID: update.Message.Chat.ID,
			State: &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			},
		}
	}

	user.Name = strings.TrimSpace(fmt.Sprintf("%s %s", update.Message.Chat.FirstName, update.Message.Chat.LastName))

	if user.BandID == primitive.NilObjectID && user.State.Name != helpers.ChooseBandState && user.State.Name != helpers.CreateBandState {
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.ChooseBandState,
		}
	}

	user, err = u.userService.UpdateOne(*user)
	if err != nil {
		return err
	}

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
					FileID: update.Message.Voice.FileID,
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
		_, err = u.userService.UpdateOne(backupUser)
	} else {
		_, err = u.userService.UpdateOne(*user)
	}

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
