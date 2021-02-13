package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scalaChordsBot/configs"
	"scalaChordsBot/entities"
	"scalaChordsBot/services"
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

	//handleFuncs, ok := stateHandlers[user.CurrentState().Name]
	//
	//if ok == false || user.CurrentState().Index > len(handleFuncs) || user.CurrentState().Index < 0 {
	//	user.CurrentState().Index = 0
	//	user.CurrentState().Name = configs.SongSearchState
	//	handleFuncs = stateHandlers[user.CurrentState().Name]
	//}
	//
	//user, err = handleFuncs[user.CurrentState().Index](u, update, user)

	user, err = enterStateHandler(u, update, user)

	if err == nil {
		user, err = u.userService.UpdateOne(user)
	}

	return err
}

func enterStateHandler(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
	handleFuncs, ok := stateHandlers[user.CurrentState().Name]

	if ok == false || user.CurrentState().Index >= len(handleFuncs) || user.CurrentState().Index < 0 {
		user.CurrentState().Index = 0
		user.CurrentState().Name = configs.SongSearchState
		handleFuncs = stateHandlers[user.CurrentState().Name]
	}

	return handleFuncs[user.CurrentState().Index](updateHandler, update, user)
}
