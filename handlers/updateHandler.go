package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scala-chords-bot/configs"
	"scala-chords-bot/entities"
	"scala-chords-bot/services"
)

type UpdateHandler struct {
	bot         *tgbotapi.BotAPI
	userService *services.UserService
	SongService *services.SongService
}

func NewHandler(bot *tgbotapi.BotAPI, userService *services.UserService, songService *services.SongService) *UpdateHandler {
	return &UpdateHandler{
		bot:         bot,
		userService: userService,
		SongService: songService,
	}
}

func (u *UpdateHandler) HandleUpdate(update *tgbotapi.Update) error {
	user, err := u.userService.FindOrCreate(update.Message.Chat.ID)

	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Что-то пошло не так.")
		_, _ = u.bot.Send(msg)
		return fmt.Errorf("couldn't get User's state %v", err)
	}

	if update.Message.Voice != nil {
		user.State = &entities.State{
			Index: 0,
			Name:  configs.UploadVoiceState,
			Context: entities.Context{
				CurrentVoice: &entities.Voice{
					TgFileID: update.Message.Voice.FileID,
					Caption:  "",
				},
			},
			Prev: user.State,
		}
	}

	user, err = u.enterStateHandler(update, user)

	if err == nil {
		user, err = u.userService.UpdateOne(user)
	}

	return err
}

func (u *UpdateHandler) enterStateHandler(update *tgbotapi.Update, user entities.User) (entities.User, error) {
	handleFuncs, ok := stateHandlers[user.State.Name]

	if ok == false || user.State.Index >= len(handleFuncs) || user.State.Index < 0 {
		user.State.Index = 0
		user.State.Name = configs.MainMenuState
		handleFuncs = stateHandlers[user.State.Name]
	}

	return handleFuncs[user.State.Index](u, update, user)
}
