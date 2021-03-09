package handlers

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
)

func mainMenuHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Основное меню:")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(helpers.MainMenuKeyboard...)
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			_, err := updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Действие не поддерживается."))
			return &user, err

		case helpers.Help:
			_, err := updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Для поиска документа, отправь боту название.\n\nРедактировать документ можно на гугл диске. Теперь не нужно отправлять файл боту, он сам обновит его.\n\nДля добавления партии, отправь боту голосовое сообщение."))
			return &user, err

		case helpers.Schedule:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.ScheduleState,
			}
			return updateHandler.enterStateHandler(update, user)

		default:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SearchSongState,
			}
			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.MainMenuState, handleFuncs
}

func scheduleHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		_, _ = updateHandler.bot.Send(chatAction)

		var allBandsEvents []*entities.Event
		for _, band := range user.Bands {
			events, err := updateHandler.bandService.GetTodayOrAfterEvents(*band)
			if err != nil {
				return nil, err
			}

			allBandsEvents = append(allBandsEvents, events...)
		}

		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.OneTimeKeyboard = false
		keyboard.ResizeKeyboard = true

		for _, event := range allBandsEvents {
			songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(event.GetAlias()))
			keyboard.Keyboard = append(keyboard.Keyboard, songButton)
		}

		keyboard.Keyboard = append(keyboard.Keyboard,
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери собрание:")
		msg.ReplyMarkup = keyboard
		_, err := updateHandler.bot.Send(msg)

		user.State.Context.Events = allBandsEvents
		user.State.Index++

		return &user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		_, _ = updateHandler.bot.Send(chatAction)

		events := user.State.Context.Events

		foundIndex := len(events)
		for i := range events {
			if events[i].GetAlias() == update.Message.Text {
				foundIndex = i
				break
			}
		}

		if foundIndex != len(events) {
			messageText := ""
			for i, pageID := range events[foundIndex].SetlistPageIDs {

				page, err := updateHandler.songService.FindNotionPageByID(pageID)
				if err != nil {
					continue
				}

				songTitleProp := page.GetTitle()
				if len(songTitleProp) < 1 {
					continue
				}

				messageText += fmt.Sprintf("%d. %s\n", i+1, songTitleProp[0].Text)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
			_, _ = updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			}
		} else {
			user.State.Index--
		}
		return updateHandler.enterStateHandler(update, user)
	})

	return helpers.ScheduleState, handleFuncs
}
