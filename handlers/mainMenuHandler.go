package handlers

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"regexp"
	"strings"
)

func mainMenuHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Основное меню:")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(helpers.MainMenuKeyboard...)
		_, err := updateHandler.bot.Send(msg)

		user.State.Index++
		return user, err
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		switch update.Message.Text {
		case "":
			_, err := updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Действие не поддерживается."))
			return user, err
		case helpers.Help:
			_, err := updateHandler.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Для поиска документа, отправь боту название.\n\nРедактировать документ можно на гугл диске. Теперь не нужно отправлять файл боту, он сам обновит его.\n\nДля добавления партии, отправь боту голосовое сообщение."))
			return user, err
		default:
			numbersRegex := regexp.MustCompile("\\(.*?\\)|[1-9.()_]*")
			update.Message.Text = numbersRegex.ReplaceAllString(update.Message.Text, "")
			newLinesRegex := regexp.MustCompile(`\s*[\t\r\n]+`)
			songNames := strings.Split(newLinesRegex.ReplaceAllString(update.Message.Text, "\n"), "\n")
			for _, songName := range songNames {
				songName = strings.TrimSpace(songName)
			}

			if len(songNames) > 1 {
				user.State = &entities.State{
					Index:   0,
					Name:    helpers.SetlistState,
					Prev:    user.State,
					Context: user.State.Context,
				}
				user.State.Context.Setlist = songNames

			} else if len(songNames) == 1 {
				update.Message.Text = songNames[0]
				user.State = &entities.State{
					Index:   0,
					Name:    helpers.SearchSongState,
					Context: user.State.Context,
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Из запроса удаляются все числа, дифизы и скобки вместе с тем, что в них.")
				_, err := updateHandler.bot.Send(msg)
				return user, err
			}

			return updateHandler.enterStateHandler(update, user)
		}
	})

	return helpers.MainMenuState, handleFuncs
}
