package handlers

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"regexp"
)

func chooseBandHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		bands, _ := updateHandler.bandService.FindAll()

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери свою группу:")

		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.OneTimeKeyboard = false
		keyboard.ResizeKeyboard = true

		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.CreateBand)))
		for i := range bands {
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(bands[i].Name)))
		}
		msg.ReplyMarkup = keyboard
		_, _ = updateHandler.bot.Send(msg)

		user.State.Context.Bands = bands
		user.State.Index++
		return &user, nil
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		case helpers.CreateBand:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.CreateBandState,
			}
			return updateHandler.enterStateHandler(update, user)
		default:
			bands := user.State.Context.Bands
			foundIndex := len(bands)
			for i := range bands {
				if bands[i].Name == update.Message.Text {
					foundIndex = i
					break
				}
			}

			if foundIndex != len(bands) {
				user.BandID = bands[foundIndex].ID
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ты добавлен в группу %s.", bands[foundIndex].Name))
				_, _ = updateHandler.bot.Send(msg)

				user.State = &entities.State{
					Index: 0,
					Name:  helpers.MainMenuState,
				}
			} else {
				user.State.Index--
			}

			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.ChooseBandState, handleFuncs
}

func createBandHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введи название своей группы:")
		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.ResizeKeyboard = true
		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
		msg.ReplyMarkup = keyboard
		_, _ = updateHandler.bot.Send(msg)

		user.State.Index++
		return &user, nil
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		default:
			user.State.Context.CurrentBand = &entities.Band{
				Name: update.Message.Text,
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Теперь добавь имейл scala-drive@scala-chords-bot.iam.gserviceaccount.com в папку на Гугл Диске как редактора. После этого отправь мне ссылку на эту папку.")
			keyboard := tgbotapi.NewReplyKeyboard()
			keyboard.ResizeKeyboard = true
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)))
			msg.ReplyMarkup = keyboard
			_, _ = updateHandler.bot.Send(msg)

			user.State.Index++
			return &user, nil
		}
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		default:
			re := regexp.MustCompile(`(/folders/|id=)(.*?)(/|$)`)
			matches := re.FindStringSubmatch(update.Message.Text)
			if matches == nil || len(matches) < 3 {
				user.State.Index--
				return updateHandler.enterStateHandler(update, user)
			}
			user.State.Context.CurrentBand.DriveFolderID = matches[2]
			user.Role = helpers.Admin
			band, err := updateHandler.bandService.UpdateOne(*user.State.Context.CurrentBand)
			if err != nil {
				return &user, err
			}

			user.BandID = band.ID

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ты добавлен в группу \"%s\" как администратор.", band.Name))
			_, _ = updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			}
			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.CreateBandState, handleFuncs
}

func addBandAdminHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери пользователя, которого ты хочешь сделать администратором:")
		keyboard := tgbotapi.NewReplyKeyboard()

		band, err := updateHandler.bandService.FindOneByID(user.BandID)
		if err != nil {
			return nil, err
		}

		users, err := updateHandler.userService.FindMultipleByBandID(band.ID)
		if err != nil {
			return nil, err
		}

		for _, bandUser := range users {
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(bandUser.Name),
			))
		}

		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Cancel),
		))

		msg.ReplyMarkup = keyboard

		_, err = updateHandler.bot.Send(msg)
		if err != nil {
			return nil, err
		}

		user.State.Index++
		user.State.Context.CurrentBandID = band.ID
		return &user, nil
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (*entities.User, error) {
		switch update.Message.Text {
		case "":
			user.State.Index--
			return updateHandler.enterStateHandler(update, user)
		default:
			band, err := updateHandler.bandService.FindOneByID(user.State.Context.CurrentBandID)
			if err != nil {
				return nil, err
			}

			users, err := updateHandler.userService.FindMultipleByBandID(band.ID)
			if err != nil {
				return nil, err
			}

			var foundUser *entities.User
			for _, bandUser := range users {
				if bandUser.Name == update.Message.Text {
					foundUser = bandUser
				}
			}

			if foundUser == nil {
				return updateHandler.enterStateHandler(update, user)
			}
			foundUser.Role = helpers.Admin
			_, err = updateHandler.userService.UpdateOne(*foundUser)
			if err != nil {
				return nil, err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Пользователь %s повышен до администратора группы %s.",
				foundUser.Name, band.Name))
			updateHandler.bot.Send(msg)

			user.State = &entities.State{
				Name: helpers.MainMenuState,
			}

			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.AddBandAdminState, handleFuncs
}
