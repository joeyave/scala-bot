package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"scalaChordsBot/configs"
	"scalaChordsBot/entities"
	"scalaChordsBot/services"
	"sort"
)

type Handler struct {
	bot         *tgbotapi.BotAPI
	userService *services.UserService
	songService *services.SongService
}

func NewHandler(bot *tgbotapi.BotAPI, userService *services.UserService, songService *services.SongService) *Handler {
	return &Handler{
		bot:         bot,
		userService: userService,
		songService: songService,
	}
}

func (h *Handler) Handle(update *tgbotapi.Update) error {
	user, err := h.userService.FindOrCreate(update.Message.Chat.ID)

	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Что-то пошло не так.")
		_, _ = h.bot.Send(msg)
		return fmt.Errorf("couldn't get User's state %v", err)
	}

	state := getUserState(&user)

	switch state.Name {
	case configs.HandleSongSearchState:
		err = h.handleSongSearchState(update, user)
	default:
		if update.Message.Text != "" {
			err = h.handleSongSearchState(update, user)
		}
	}

	return err
}

func (h *Handler) handleSongSearchState(update *tgbotapi.Update, user entities.User) error {
	state := getUserState(&user)

	switch state.Index {
	case 0:
		{
			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			_, _ = h.bot.Send(chatAction)

			songs, err := h.songService.FindByName(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено.")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				_, _ = h.bot.Send(msg)
				return fmt.Errorf("couldn't find Song %v", err)
			}

			songsKeyboard := tgbotapi.NewReplyKeyboard()
			songsKeyboard.OneTimeKeyboard = false
			songsKeyboard.ResizeKeyboard = true

			// TODO: some sort of pagination.
			const pageSize = 100
			for i, song := range songs {
				if i == pageSize {
					break
				}

				songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.Name))
				songsKeyboard.Keyboard = append(songsKeyboard.Keyboard, songButton)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери песню:")
			msg.ReplyMarkup = songsKeyboard
			_, _ = h.bot.Send(msg)

			state.Index++
			state.Name = configs.HandleSongSearchState
			state.Context.Songs = songs

			user.States = append(user.States, *state)
			user, err = h.userService.UpdateOne(user)
			if err != nil {
				return err
			}
		}
	case 1:
		{
			songs := state.Context.Songs

			chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
			_, _ = h.bot.Send(chatAction)

			sort.Slice(songs, func(i, j int) bool {
				return songs[i].Name <= songs[j].Name
			})

			foundIndex := sort.Search(len(songs), func(i int) bool {
				return songs[i].Name >= update.Message.Text
			})

			if foundIndex != len(songs) {
				fileReader, err := h.songService.DownloadPDF(songs[foundIndex])
				if err != nil {
					return err
				}

				msg := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, *fileReader)

				keyboard := configs.GetSongOptionsKeyboard()
				keyboard = append([][]tgbotapi.KeyboardButton{{{Text: songs[foundIndex].Name}}}, keyboard...)
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

				_, err = h.bot.Send(msg)
				if err != nil {
					return err
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ничего не найдено.")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				_, err := h.bot.Send(msg)
				if err != nil {
					return err
				}
			}

			user.States = user.States[:len(user.States)-1]
			_, err := h.userService.UpdateOne(user)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getUserState(user *entities.User) *entities.State {
	var state *entities.State
	if user.States != nil && len(user.States) > 0 {
		state = &user.States[len(user.States)-1]
	} else {
		state = &entities.State{
			Index:   0,
			Name:    "",
			Context: entities.Context{},
		}
	}

	return state
}
