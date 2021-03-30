package handlers

import (
	"fmt"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/telebot/v3"
	"github.com/kjk/notionapi"
	"github.com/klauspost/lctime"
	"google.golang.org/api/drive/v3"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func mainMenuHandler() (string, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		err := c.Send("Основное меню:", &telebot.ReplyMarkup{
			ReplyKeyboard:  helpers.MainMenuKeyboard,
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}
		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {

		case helpers.Help:
			err := c.Send("Для поиска документа, отправь боту название.\n\nРедактировать документ можно на гугл диске.\n\nДля добавления партии, отправь боту голосовое сообщение.")
			return err

		case helpers.Schedule:
			user.State = &entities.State{
				Name: helpers.ScheduleState,
			}

		// TODO
		case "/schedule_dev":
			user.State = &entities.State{
				Name: helpers.GetEventsState,
			}

		case helpers.Settings:
			err := c.Send("Настройки:", &telebot.ReplyMarkup{
				ReplyKeyboard:  helpers.SettingsKeyboard,
				ResizeKeyboard: true,
			})
			if err != nil {
				return err
			}

			user.State.Index++
			return nil

		case helpers.CreateDoc:
			user.State = &entities.State{
				Name: helpers.CreateSongState,
			}

		case helpers.AddAdmin:
			user.State = &entities.State{
				Name: helpers.AddBandAdminState,
			}

		default:
			user.State = &entities.State{
				Name: helpers.SearchSongState,
			}
		}

		return h.enter(c, user)
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.ChangeBand:
			user.State = &entities.State{
				Name: helpers.ChooseBandState,
			}
		case helpers.CreateRole:
			user.State = &entities.State{
				Name: helpers.CreateRoleState,
			}
		}

		return h.enter(c, user)
	})

	return helpers.MainMenuState, handlerFuncs
}

func createRoleHandler() (string, []HandlerFunc) {

	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		err := c.Send("Отправь название новой роли. Например, лид-вокал, проповедник и т. д.", &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		role, err := h.roleService.UpdateOne(entities.Role{
			Name:   c.Text(),
			BandID: user.BandID,
		})
		if err != nil {
			return err
		}

		err = c.Send(fmt.Sprintf("Добавлена новая роль: %s.", role.Name))
		if err != nil {
			return err
		}

		user.State = &entities.State{Name: helpers.MainMenuState}
		return h.enter(c, user)
	})

	return helpers.CreateRoleState, handlerFuncs
}

func scheduleHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		c.Notify(telebot.Typing)

		events, err := h.bandService.GetTodayOrAfterEvents(*user.Band)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		for _, event := range events {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: event.GetAlias()}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err = c.Send("Выбери собрание:", markup)
		if err != nil {
			return err
		}

		user.State.Context.NotionEvents = events
		user.State.Index++

		return err
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		c.Notify(telebot.Typing)

		events := user.State.Context.NotionEvents

		var foundEvent *entities.NotionEvent
		for _, event := range events {
			if event.GetAlias() == c.Text() {
				foundEvent = event
				break
			}
		}

		if foundEvent != nil {
			messageText := fmt.Sprintf("<b><a href=\"https://www.notion.so/%s\">%s</a></b>\n\n",
				notionapi.ToNoDashID(foundEvent.ID), foundEvent.GetAlias())

			for i, pageID := range foundEvent.SetlistPageIDs {

				page, err := h.songService.FindNotionPageByID(pageID)
				if err != nil {
					continue
				}

				songTitleProp := page.GetTitle()
				if len(songTitleProp) < 1 {
					continue
				}
				songTitle := songTitleProp[0].Text

				songKey := "?"
				songKeyProp := page.GetProperty("OR>-")
				if len(songKeyProp) > 0 {
					songKey = songKeyProp[0].Text
				}

				songBPM := "?"
				songBPMProp := page.GetProperty("j0]A")
				if len(songBPMProp) > 0 {
					songBPM = songBPMProp[0].Text
				}

				user.State.Context.SongNames = append(user.State.Context.SongNames, songTitle)

				messageText += fmt.Sprintf("%d. %s (<a href=\"https://www.notion.so/%s\">%s, %s</a>)\n",
					i+1, songTitle, notionapi.ToNoDashID(pageID), songKey, songBPM)
			}

			err := c.Send(messageText, &telebot.SendOptions{
				ReplyMarkup: &telebot.ReplyMarkup{
					ReplyKeyboard:  helpers.FindChordsKeyboard,
					ResizeKeyboard: true,
				},
				DisableWebPagePreview: true,
				ParseMode:             telebot.ModeHTML,
			})
			if err != nil {
				return err
			}

			user.State.Index++
			return nil
		} else {
			user.State.Index--
			return h.enter(c, user)
		}
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {

		case helpers.Back:
			user.State.Index = 0
			return h.enter(c, user)

		case helpers.FindChords:
			user.State = &entities.State{
				Index:   0,
				Name:    helpers.SearchSongState,
				Context: entities.Context{Query: strings.Join(user.State.Context.SongNames, "\n")},
			}
			return h.enter(c, user)

		default:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.SearchSongState,
			}
			return h.enter(c, user)
		}
	})

	return helpers.ScheduleState, handlerFunc
}

func getEventsHandler() (string, []HandlerFunc) {

	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		events, err := h.eventService.FindMultipleByBandIDFromTodayByBandID(user.BandID)
		user.State.Context.Events = events

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		for _, event := range events {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: event.Alias()}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.GetAllEvents}})
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Back}, {Text: helpers.CreateEvent}})

		err = c.Send("Выбери собрание:", markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.CreateEvent:
			user.State = &entities.State{
				Name: helpers.CreateEventState,
				Prev: user.State,
			}
			user.State.Prev.Index = 0
			return h.enter(c, user)
		case helpers.GetAllEvents:
			for _, event := range user.State.Context.Events {
				eventString, err := h.eventService.ToHtmlStringByID(event.ID)
				if err != nil {
					continue
				}
				err = c.Send(eventString, telebot.ModeHTML)
				if err != nil {
					return err
				}
			}

			return nil

		default:
			events := user.State.Context.Events

			var foundEvent *entities.Event
			for _, event := range events {
				if c.Text() == event.Alias() {
					foundEvent = event
					break
				}
			}

			if foundEvent != nil {
				user.State = &entities.State{
					Name: helpers.EventActionsState,
					Context: entities.Context{
						EventID: foundEvent.ID,
					},
					Prev: user.State,
				}
				user.State.Prev.Index = 0
				return h.enter(c, user)
			} else {
				user.State.Index--
				return h.enter(c, user)
			}
		}
	})

	return helpers.GetEventsState, handlerFuncs
}

func createEventHandler() (string, []HandlerFunc) {

	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		err := lctime.SetLocale("ru_RU")
		if err != nil {
			fmt.Println(err)
		}
		start := time.Now()
		end := start.AddDate(0, 1, 0)

		for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
			timeStr := lctime.Strftime("%A / %d.%m.%Y", d)
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: timeStr}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err = c.Send("Выбери дату:", markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		re := regexp.MustCompile(`(\d{1,2}).(\d{1,2}).(\d{4})`)
		matches := re.FindStringSubmatch(c.Text())

		if len(matches) < 4 {
			return c.Send("Неверный формат. Введи дату в формате 01.02.2021.")
		}

		year, _ := strconv.Atoi(matches[3])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[1])

		parsedTime, err := time.Parse("02-01-2006", fmt.Sprintf("%02d-%02d-%d", day, month, year))
		if err != nil {
			return c.Send("Что-то не так. Введи дату в формате 01.02.2021.")
		}

		user.State.Context.Map = map[string]string{"time": parsedTime.Format(time.RFC3339)}

		err = c.Send("Введи название этого собрания:", &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		parsedTime, err := time.Parse(time.RFC3339, user.State.Context.Map["time"])
		if err != nil {
			user.State = &entities.State{Name: helpers.CreateEventState}
			return h.enter(c, user)
		}

		event, err := h.eventService.UpdateOne(entities.Event{
			Time:   parsedTime,
			Name:   c.Text(),
			BandID: user.BandID,
		})
		if err != nil {
			return err
		}

		user.State = &entities.State{
			Name: helpers.EventActionsState,
			Context: entities.Context{
				EventID: event.ID,
			},
		}
		return h.enter(c, user)
	})

	return helpers.CreateEventState, handlerFuncs
}

func eventActionsHandler() (string, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		eventString, err := h.eventService.ToHtmlStringByID(user.State.Context.EventID)
		if err != nil {
			return err
		}

		err = c.Send(eventString, &telebot.ReplyMarkup{
			ReplyKeyboard:  append(helpers.EventActionsKeyboard, helpers.BackOrMenuKeyboard...),
			ResizeKeyboard: true,
		}, telebot.ModeHTML)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		event, err := h.eventService.FindOneByID(user.State.Context.EventID)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		for _, role := range event.Band.Roles {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: role.Name}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err = c.Send("Кем будет этот участник?", markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		event, err := h.eventService.FindOneByID(user.State.Context.EventID)
		if err != nil {
			return err
		}

		var foundRole *entities.Role
		for _, role := range event.Band.Roles {
			if role.Name == c.Text() {
				foundRole = role
			}
		}

		if foundRole == nil {
			user.State.Index--
			return h.enter(c, user)
		}

		user.State.Context.RoleID = foundRole.ID

		users, err := h.userService.FindMultipleByBandID(event.BandID)
		if err != nil {
			return err
		}
		sort.Slice(users, func(i, j int) bool {
			return len(users[i].Memberships) < len(users[j].Memberships)
		})

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		for _, user := range users {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: user.Name}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err = c.Send("Выбери участника", markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		event, err := h.eventService.FindOneByID(user.State.Context.EventID)
		if err != nil {
			return err
		}

		users, err := h.userService.FindMultipleByBandID(event.BandID)
		if err != nil {
			return err
		}

		var foundUser *entities.User
		for _, user := range users {
			if user.Name == c.Text() {
				foundUser = user
			}
		}

		if foundUser == nil {
			user.State.Index--
			return h.enter(c, user)
		}

		_, err = h.membershipService.UpdateOne(entities.Membership{
			EventID: user.State.Context.EventID,
			UserID:  foundUser.ID,
			RoleID:  user.State.Context.RoleID,
		})
		if err != nil {
			return err
		}

		eventString, _ := h.eventService.ToHtmlStringByID(event.ID)
		h.bot.Send(telebot.ChatID(foundUser.ID),
			fmt.Sprintf("Привет. Ты учавствуешь в собрании! "+
				"Вот план:\n\n%s", eventString), telebot.ModeHTML)

		user.State.Index = 0
		return h.enter(c, user)
	})

	return helpers.EventActionsState, handlerFuncs
}

func chooseBandHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		bands, err := h.bandService.FindAll()
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.CreateBand}})
		for _, band := range bands {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: band.Name}})
		}

		err = c.Send("Выбери свою группу:", markup)
		if err != nil {
			return err
		}

		user.State.Context.Bands = bands
		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.CreateBand:
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.CreateBandState,
			}
			return h.enter(c, user)

		default:
			bands := user.State.Context.Bands
			var foundBand *entities.Band
			for _, band := range bands {
				if band.Name == c.Text() {
					foundBand = band
					break
				}
			}

			if foundBand != nil {
				err := c.Send(fmt.Sprintf("Ты добавлен в группу %s.", foundBand.Name))
				if err != nil {
					return err
				}

				user.BandID = foundBand.ID
				user.State = &entities.State{
					Index: 0,
					Name:  helpers.MainMenuState,
				}
			} else {
				user.State.Index--
			}

			return h.enter(c, user)
		}
	})

	return helpers.ChooseBandState, handlerFunc
}

func createBandHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		err := c.Send("Введи название своей группы:", &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		user.State.Context.Band = &entities.Band{
			Name: c.Text(),
		}

		err := c.Send("Теперь добавь имейл scala-drive@scala-chords-bot.iam.gserviceaccount.com в папку на Гугл Диске как редактора. После этого отправь мне ссылку на эту папку.",
			&telebot.ReplyMarkup{
				ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
				ResizeKeyboard: true,
			})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		re := regexp.MustCompile(`(/folders/|id=)(.*?)(/|\?|$)`)
		matches := re.FindStringSubmatch(c.Text())
		if matches == nil || len(matches) < 3 {
			user.State.Index--
			return h.enter(c, user)
		}
		user.State.Context.Band.DriveFolderID = matches[2]
		user.Role = helpers.Admin
		band, err := h.bandService.UpdateOne(*user.State.Context.Band)
		if err != nil {
			return err
		}

		user.BandID = band.ID

		err = c.Send(fmt.Sprintf("Ты добавлен в группу \"%s\" как администратор.", band.Name))
		if err != nil {
			return err
		}

		user.State = &entities.State{
			Name: helpers.MainMenuState,
		}
		return h.enter(c, user)
	})

	return helpers.CreateBandState, handlerFunc
}

//func addBandAdminHandler() (string, []HandlerFunc) {
//	handlerFunc := make([]HandlerFunc, 0)
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
//		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери пользователя, которого ты хочешь сделать администратором:")
//		keyboard := tgbotapi.NewReplyKeyboard()
//
//		band, err := h.bandService.FindOneByID(user.BandID)
//		if err != nil {
//			return nil, err
//		}
//
//		users, err := h.userService.FindMultipleByBandID(band.ID)
//		if err != nil {
//			return nil, err
//		}
//
//		for _, bandUser := range users {
//			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(
//				tgbotapi.NewKeyboardButton(bandUser.Role),
//			))
//		}
//
//		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(
//			tgbotapi.NewKeyboardButton(helpers.Cancel),
//		))
//
//		msg.ReplyMarkup = keyboard
//
//		_, err = h.bot.Send(msg)
//		if err != nil {
//			return nil, err
//		}
//
//		user.State.Index++
//		user.State.Context.BandID = band.ID
//		return nil
//	})
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
//		switch update.Message.Text {
//		case "":
//			user.State.Index--
//			return h.enter(c, user)
//		default:
//			band, err := h.bandService.FindOneByID(user.State.Context.BandID)
//			if err != nil {
//				return nil, err
//			}
//
//			users, err := h.userService.FindMultipleByBandID(band.ID)
//			if err != nil {
//				return nil, err
//			}
//
//			var foundUser *entities.User
//			for _, bandUser := range users {
//				if bandUser.Role == update.Message.Text {
//					foundUser = bandUser
//				}
//			}
//
//			if foundUser == nil {
//				return h.enter(c, user)
//			}
//			foundUser.Role = helpers.Admin
//			_, err = h.userService.UpdateOne(*foundUser)
//			if err != nil {
//				return nil, err
//			}
//
//			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Пользователь %s повышен до администратора группы %s.",
//				foundUser.Role, band.Role))
//			h.bot.Send(msg)
//
//			user.State = &entities.State{
//				Role: helpers.MainMenuState,
//			}
//
//			return h.enter(c, user)
//		}
//	})
//
//	return helpers.AddBandAdminState, handlerFunc
//}

func searchSongHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	// Print list of found songs.
	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		{
			c.Notify(telebot.Typing)

			query := c.Text()

			if query == helpers.SearchEverywhere || query == helpers.Back || query == helpers.FindChords {
				query = user.State.Context.Query
			}

			query = helpers.CleanUpQuery(query)
			songNames := helpers.SplitQueryByNewlines(query)

			if len(songNames) > 1 {
				user.State = &entities.State{
					Index: 0,
					Name:  helpers.SetlistState,
					Prev: &entities.State{
						Index: 0,
						Name:  helpers.MainMenuState,
					},
					Context: user.State.Context,
				}
				user.State.Context.SongNames = songNames
				return h.enter(c, user)

			} else if len(songNames) == 1 {
				query = songNames[0]
				user.State.Context.Query = query
			} else {
				err := c.Send("Из запроса удаляются все числа, дифизы и скобки вместе с тем, что в них.")
				if err != nil {
					return err
				}

				user.State = &entities.State{
					Name: helpers.MainMenuState,
				}

				return h.enter(c, user)
			}

			var driveFiles []*drive.File
			var err error
			if c.Text() == helpers.SearchEverywhere {
				driveFiles, _, err = h.driveFileService.FindSomeByNameAndFolderID(query, "", "")
			} else {
				driveFiles, _, err = h.driveFileService.FindSomeByNameAndFolderID(query, user.Band.DriveFolderID, "")
			}
			if err != nil {
				return err
			}

			if len(driveFiles) == 0 {
				return c.Send("Ничего не найдено. Попробуй еще раз.", &telebot.ReplyMarkup{
					ReplyKeyboard:  helpers.SearchEverywhereKeyboard,
					ResizeKeyboard: true,
				})
			}

			markup := &telebot.ReplyMarkup{
				ResizeKeyboard: true,
			}

			// TODO: some sort of pagination.
			for _, song := range driveFiles {
				markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: song.Name}})
			}

			markup.ReplyKeyboard = append(markup.ReplyKeyboard, helpers.SearchEverywhereKeyboard...)

			err = c.Send("Выбери песню:", markup)
			if err != nil {
				return err
			}

			user.State.Context.DriveFiles = driveFiles
			user.State.Index++
			return nil
		}
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {

		case helpers.SearchEverywhere:
			user.State.Index--
			return h.enter(c, user)

		default:
			c.Notify(telebot.UploadingDocument)

			driveFiles := user.State.Context.DriveFiles
			var foundDriveFile *drive.File
			for _, driveFile := range driveFiles {
				if driveFile.Name == c.Text() {
					foundDriveFile = driveFile
					break
				}
			}

			if foundDriveFile != nil {
				user.State = &entities.State{
					Name: helpers.SongActionsState,
					Context: entities.Context{
						DriveFileID: foundDriveFile.Id,
					},
					Prev: user.State,
				}
				return h.enter(c, user)
			} else {
				user.State.Index--
				return h.enter(c, user)
			}
		}
	})

	return helpers.SearchSongState, handlerFunc
}

func songActionsHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		c.Notify(telebot.UploadingDocument)

		driveFileID := user.State.Context.DriveFileID

		song, err := h.songService.FindOrCreateOneByDriveFileID(driveFileID)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: song.DriveFile.Name}}},
			ResizeKeyboard: true,
		}

		if song.BandID == user.BandID {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, helpers.SongActionsKeyboard...)
		} else {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, helpers.RestrictedSongActionsKeyboard...)
		}

		sendDocumentByReader := func() (*telebot.Message, error) {
			reader, err := h.driveFileService.DownloadOneByID(driveFileID)
			if err != nil {
				return nil, err
			}

			msg, err := h.bot.Send(c.Recipient(), &telebot.Document{
				File:     telebot.FromReader(*reader),
				MIME:     "application/pdf",
				FileName: fmt.Sprintf("%s.pdf", song.DriveFile.Name),
			}, markup)
			if err != nil {
				return nil, err
			}
			return msg, nil
		}

		var sentMessage *telebot.Message
		if song.HasOutdatedPDF() {
			sentMessage, err = sendDocumentByReader()
			if err != nil {
				return err
			}
		} else {
			sentMessage, err = h.bot.Send(c.Recipient(), &telebot.Document{
				File:     telebot.File{FileID: song.PDF.TgFileID},
				MIME:     "application/pdf",
				FileName: fmt.Sprintf("%s.pdf", song.DriveFile.Name),
			}, markup)
			if err != nil {
				sentMessage, err = sendDocumentByReader()
				if err != nil {
					return err
				}
			}
		}

		song.PDF.TgFileID = sentMessage.Document.FileID

		if song.HasOutdatedPDF() || song.PDF.TgChannelMessageID == 0 {
			song = helpers.SendToChannel(h.bot, song)
		}

		song.PDF.ModifiedTime = song.DriveFile.ModifiedTime

		song, err = h.songService.UpdateOne(*song)
		if err != nil {
			return fmt.Errorf("failed to cache file %v", err)
		}

		user.State.Index++
		user.State.Context.DriveFileID = song.DriveFileID

		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		song, err := h.songService.FindOneByDriveFileID(user.State.Context.DriveFileID)
		if err != nil {
			return err
		}

		switch c.Text() {
		case helpers.Back:
			user.State = user.State.Prev

		case helpers.Voices:
			user.State = &entities.State{
				Name: helpers.GetVoicesState,
				Context: entities.Context{
					DriveFileID: user.State.Context.DriveFileID,
				},
				Prev: user.State,
			}

		case helpers.Audios:
			return c.Send("Функция еще не реалилованна. В будущем планируется хранинть тут аудиозаписи песни в нужной тональности.")

		case helpers.Transpose:
			user.State = &entities.State{
				Name: helpers.TransposeSongState,
				Context: entities.Context{
					DriveFileID: user.State.Context.DriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case helpers.Style:
			user.State = &entities.State{
				Name: helpers.StyleSongState,
				Context: entities.Context{
					DriveFileID: user.State.Context.DriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case helpers.CopyToMyBand:
			user.State = &entities.State{
				Name: helpers.CopySongState,
				Context: entities.Context{
					DriveFileID: user.State.Context.DriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case helpers.Delete:
			user.State = &entities.State{
				Name: helpers.DeleteSongState,
				Context: entities.Context{
					DriveFileID: user.State.Context.DriveFileID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 0

		case song.DriveFile.Name:
			return c.Send(song.DriveFile.WebViewLink)

		default:
			user.State = &entities.State{
				Name: helpers.SearchSongState,
			}
		}

		return h.enter(c, user)
	})

	return helpers.SongActionsState, handlerFunc
}

func transposeSongHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		err := c.Send("Выбери новую тональность:", &telebot.ReplyMarkup{
			ReplyKeyboard:  append(helpers.KeysKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}}),
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		_, err := transposer.ParseChord(c.Text())
		if err != nil {
			user.State.Index--
			return h.enter(c, user)
		}
		user.State.Context.Key = c.Text()

		markup := &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.AppendSection}}},
			ResizeKeyboard: true,
		}

		sectionsNumber, err := h.driveFileService.GetSectionsNumber(user.State.Context.DriveFileID)
		if err != nil {
			return err
		}

		for i := 0; i < sectionsNumber; i++ {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: fmt.Sprintf("Вместо %d-й секции", i+1)}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err = c.Send("Куда ты хочешь вставить новую тональность?", markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return err
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		c.Notify(telebot.UploadingDocument)

		re := regexp.MustCompile("[1-9]+")
		sectionIndex, err := strconv.Atoi(re.FindString(c.Text()))
		if err != nil {
			sectionIndex = -1
		} else {
			sectionIndex--
		}

		driveFile, err := h.driveFileService.TransposeOne(user.State.Context.DriveFileID, user.State.Context.Key, sectionIndex)
		if err != nil {
			return err
		}

		song, err := h.songService.FindOneByDriveFileID(driveFile.Id)
		if err != nil {
			return err
		}

		fakeTime, _ := time.Parse("2006", "2006")
		song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)

		_, err = h.songService.UpdateOne(*song)
		if err != nil {
			return err
		}

		user.State = user.State.Prev
		user.State.Context.DriveFileID = driveFile.Id
		return h.enter(c, user)
	})

	return helpers.TransposeSongState, handlerFunc
}

func styleSongHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	// Print list of found songs.
	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		c.Notify(telebot.Typing)

		driveFile, err := h.driveFileService.StyleOne(user.State.Context.DriveFileID)
		if err != nil {
			return err
		}

		song, err := h.songService.FindOneByDriveFileID(driveFile.Id)
		if err != nil {
			return err
		}

		fakeTime, _ := time.Parse("2006", "2006")
		song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)

		_, err = h.songService.UpdateOne(*song)
		if err != nil {
			return err
		}

		user.State = user.State.Prev
		user.State.Context.DriveFileID = driveFile.Id
		return h.enter(c, user)
	})
	return helpers.StyleSongState, handlerFunc
}

func copySongHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		c.Notify(telebot.Typing)

		file, err := h.driveFileService.FindOneByID(user.State.Context.DriveFileID)
		if err != nil {
			return err
		}

		file = &drive.File{
			Name:    file.Name,
			Parents: []string{user.Band.DriveFolderID},
		}

		copiedSong, err := h.driveFileService.CloneOne(user.State.Context.DriveFileID, file)
		if err != nil {
			return err
		}

		user.State = user.State.Prev
		user.State.Context.DriveFileID = copiedSong.Id

		return h.enter(c, user)
	})

	return helpers.CopySongState, handlerFunc
}

func createSongHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		err := c.Send("Отправь название:", &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		user.State.Context.CreateSongPayload.Name = c.Text()
		err := c.Send("Отправь слова:", &telebot.ReplyMarkup{
			ReplyKeyboard:  helpers.CancelOrSkipKeyboard,
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.Lyrics = c.Text()
		}

		err := c.Send("Выбери или отправь тональность:", &telebot.ReplyMarkup{
			ReplyKeyboard:  append(helpers.KeysKeyboard, helpers.CancelOrSkipKeyboard...),
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.Key = c.Text()
		}

		err := c.Send("Отправь темп:", &telebot.ReplyMarkup{
			ReplyKeyboard:  helpers.CancelOrSkipKeyboard,
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.BPM = c.Text()
		}

		err := c.Send("Выбери или отправь размер:", &telebot.ReplyMarkup{
			ReplyKeyboard:  append(helpers.TimesKeyboard, helpers.CancelOrSkipKeyboard...),
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.Skip:
		default:
			user.State.Context.CreateSongPayload.Time = c.Text()
		}

		c.Notify(telebot.UploadingDocument)

		file := &drive.File{
			Name:     user.State.Context.CreateSongPayload.Name,
			Parents:  []string{user.Band.DriveFolderID},
			MimeType: "application/vnd.google-apps.document",
		}
		newFile, err := h.driveFileService.CreateOne(
			file,
			user.State.Context.CreateSongPayload.Lyrics,
			user.State.Context.CreateSongPayload.Key,
			user.State.Context.CreateSongPayload.BPM,
			user.State.Context.CreateSongPayload.Time,
		)

		if err != nil {
			return err
		}

		newFile, err = h.driveFileService.StyleOne(newFile.Id)
		if err != nil {
			return err
		}

		user.State = &entities.State{
			Index: 0,
			Name:  helpers.SongActionsState,
			Context: entities.Context{
				DriveFileID: newFile.Id,
			},
		}

		return h.enter(c, user)
	})

	return helpers.CreateSongState, handlerFunc
}

//func deleteSongHandler() (string, []HandlerFunc) {
//	handlerFunc := make([]HandlerFunc, 0)
//
//	// TODO: allow deleting Song only if it belongs to the User's Band.
//	// TODO: delete from channel.
//	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
//		if user.Role == helpers.Admin {
//			err := h.songService.DeleteOneByID(user.State.Context.DriveFileID)
//			if err != nil {
//				return nil, err
//			}
//
//			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Удалено.")
//			_, _ = h.bot.Send(msg)
//
//			user.State = &entities.State{
//				Role: helpers.MainMenuState,
//			}
//			return h.enter(c, user)
//
//		} else {
//			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для удаления песни нужно быть администратором группы.")
//			_, _ = h.bot.Send(msg)
//		}
//
//		user.State = user.State.Prev
//		return h.enter(c, user)
//	})
//
//	return helpers.DeleteSongState, handlerFunc
//}

func getVoicesHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		song, err := h.songService.FindOneByDriveFileID(user.State.Context.DriveFileID)
		if err != nil {
			return err
		}

		if song.Voices == nil || len(song.Voices) == 0 {
			err := c.Send("У этой песни нет партий. Чтобы добавить, отправь мне голосовое сообщение.")
			if err != nil {
				return err
			}
			user.State = user.State.Prev
			return err
		} else {
			markup := &telebot.ReplyMarkup{
				ResizeKeyboard: true,
			}

			for _, voice := range song.Voices {
				markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: voice.Caption}})
			}
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Back}})

			err := c.Send("Выбери партию:", markup)
			if err != nil {
				return err
			}

			user.State.Index++
			return nil
		}
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.Back:
			user.State = user.State.Prev
			user.State.Index = 0
			return h.enter(c, user)
		default:
			song, err := h.songService.FindOneByDriveFileID(user.State.Context.DriveFileID)
			if err != nil {
				return err
			}

			var foundVoice *entities.Voice
			for _, voice := range song.Voices {
				if voice.Caption == c.Text() {
					foundVoice = voice
				}
			}

			if foundVoice != nil {
				err := c.Send(&telebot.Voice{
					File:    telebot.File{FileID: foundVoice.FileID},
					Caption: foundVoice.Caption,
				}, &telebot.ReplyMarkup{
					ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Back}}, {{Text: helpers.Delete}}},
					ResizeKeyboard: true,
				})
				if err != nil {
					return err
				}

				user.State.Index++
				return nil
			} else {
				return c.Send("Нет партии c таким названием. Попробуй еще раз.")
			}
		}
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.Back:
			user.State.Index = 0
			return h.enter(c, user)
		case helpers.Delete:
			// TODO: handle delete
			return nil
		default:
			return c.Send("Я тебя не понимаю. Нажми на кнопку.")
		}
	})

	return helpers.GetVoicesState, handlerFunc
}

func uploadVoiceHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		err := c.Send("Введи название песни, к которой ты хочешь прикрепить эту партию:", &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		c.Notify(telebot.Typing)

		driveFiles, _, err := h.driveFileService.FindSomeByNameAndFolderID(c.Text(), user.Band.DriveFolderID, "")
		if err != nil {
			return err
		}

		if len(driveFiles) == 0 {
			return c.Send("Ничего не найдено. Попробуй другое название.", &telebot.ReplyMarkup{
				ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
				ResizeKeyboard: true,
			})
		}

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		// TODO: some sort of pagination.
		for _, driveFile := range driveFiles {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: driveFile.Name}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err = c.Send("Выбери песню:", markup)
		if err != nil {
			return err
		}

		user.State.Context.DriveFiles = driveFiles
		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		c.Notify(telebot.UploadingDocument)

		driveFiles := user.State.Context.DriveFiles
		var foundDriveFile *drive.File
		for _, driveFile := range driveFiles {
			if driveFile.Name == c.Text() {
				foundDriveFile = driveFile
				break
			}
		}

		if foundDriveFile != nil {
			song, err := h.songService.FindOrCreateOneByDriveFileID(foundDriveFile.Id)
			if err != nil {
				return err
			}

			user.State.Context.DriveFileID = song.DriveFileID

			err = c.Send("Отправь мне название этой партии:", &telebot.ReplyMarkup{
				ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.Cancel}}},
				ResizeKeyboard: true,
			})
			if err != nil {
				return err
			}

			user.State.Index++
			return nil
		} else {
			user.State.Index--
			return h.enter(c, user)
		}

	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		user.State.Context.Voice.Caption = c.Text()

		song, err := h.songService.FindOneByDriveFileID(user.State.Context.DriveFileID)
		if err != nil {
			return err
		}

		user.State.Context.Voice.SongID = song.ID

		_, err = h.voiceService.UpdateOne(*user.State.Context.Voice)
		if err != nil {
			return err
		}

		c.Send("Добавление завершено.")

		user.State = &entities.State{
			Name: helpers.SongActionsState,
			Context: entities.Context{
				DriveFileID: user.State.Context.DriveFileID,
			},
		}
		return h.enter(c, user)

	})
	return helpers.UploadVoiceState, handlerFunc
}

func setlistHandler() (string, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		if len(user.State.Context.SongNames) < 1 {
			user.State.Index = 2
			return h.enter(c, user)
		}

		songNames := user.State.Context.SongNames

		currentSongName := songNames[0]
		user.State.Context.SongNames = songNames[1:]

		c.Notify(telebot.Typing)

		driveFiles, _, err := h.driveFileService.FindSomeByNameAndFolderID(currentSongName, user.Band.DriveFolderID, "")
		if err != nil {
			return err
		}

		if len(driveFiles) == 0 {
			msg, err := h.bot.Send(c.Recipient(), fmt.Sprintf("По запросу \"%s\" ничего не найдено. Напиши новое название или пропусти эту песню.", currentSongName), &telebot.ReplyMarkup{
				ReplyKeyboard:  helpers.CancelOrSkipKeyboard,
				ResizeKeyboard: true,
			})
			if err != nil {
				return err
			}

			user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, msg.ID)
			user.State.Index++
			return err
		}

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		// TODO: some sort of pagination.
		for _, song := range driveFiles {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: song.Name}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, helpers.CancelOrSkipKeyboard...)

		msg, err := h.bot.Send(c.Recipient(), fmt.Sprintf("Выбери песню по запросу \"%s\" или введи другое название:", currentSongName), markup)
		if err != nil {
			return err
		}

		user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, msg.ID)
		user.State.Context.DriveFiles = driveFiles
		user.State.Index++

		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, c.Message().ID)

		switch c.Text() {
		case helpers.Skip:
			user.State.Index = 0
			return h.enter(c, user)
		}

		driveFiles := user.State.Context.DriveFiles
		var foundDriveFile *drive.File
		for _, driveFile := range driveFiles {
			if driveFile.Name == c.Text() {
				foundDriveFile = driveFile
				break
			}
		}

		if foundDriveFile != nil {
			user.State.Context.FoundDriveFileIDs = append(user.State.Context.FoundDriveFileIDs, foundDriveFile.Id)
		} else {
			user.State.Context.SongNames = append([]string{c.Text()}, user.State.Context.SongNames...)
		}

		user.State.Index = 0
		return h.enter(c, user)
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		c.Notify(telebot.UploadingDocument)

		foundDriveFileIDs := user.State.Context.FoundDriveFileIDs

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(foundDriveFileIDs))
		documents := make([]telebot.InputMedia, len(foundDriveFileIDs))
		for i := range user.State.Context.FoundDriveFileIDs {
			go func(i int) {
				defer waitGroup.Done()

				song, err := h.songService.FindOrCreateOneByDriveFileID(foundDriveFileIDs[i])
				if err != nil {
					return
				}

				if song.HasOutdatedPDF() {
					reader, err := h.driveFileService.DownloadOneByID(song.DriveFileID)
					if err != nil {
						return
					}

					documents[i] = &telebot.Document{
						File:     telebot.FromReader(*reader),
						MIME:     "application/pdf",
						FileName: fmt.Sprintf("%s.pdf", song.DriveFile.Name),
					}
				} else {
					documents[i] = &telebot.Document{
						File: telebot.File{FileID: song.PDF.TgFileID},
					}
				}
			}(i)
		}
		waitGroup.Wait()

		const chunkSize = 10
		chunks := chunkAlbumBy(documents, chunkSize)

		for i, album := range chunks {
			responses, err := h.bot.SendAlbum(c.Recipient(), album)

			// TODO: check for bugs.
			if err != nil {
				fromIndex := 0
				toIndex := 0 + len(album)

				if i-1 > 0 && i-1 < len(chunks) {
					fromIndex = i * len(chunks[i-1])
					toIndex = fromIndex + len(chunks[i])
				}

				foundDriveFileIDs := user.State.Context.FoundDriveFileIDs[fromIndex:toIndex]

				var waitGroup sync.WaitGroup
				waitGroup.Add(len(foundDriveFileIDs))
				documents := make([]telebot.InputMedia, len(foundDriveFileIDs))
				for i := range foundDriveFileIDs {
					go func(i int) {
						defer waitGroup.Done()
						reader, err := h.driveFileService.DownloadOneByID(foundDriveFileIDs[i])
						if err != nil {
							return
						}

						driveFile, err := h.driveFileService.FindOneByID(foundDriveFileIDs[i])
						if err != nil {
							return
						}

						documents[i] = &telebot.Document{
							File:     telebot.FromReader(*reader),
							MIME:     "application/pdf",
							FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
						}
					}(i)
				}
				waitGroup.Wait()

				responses, err = h.bot.SendAlbum(c.Recipient(), documents)
				if err != nil {
					continue
				}
			}

			for j := range responses {
				foundDriveFileID := user.State.Context.FoundDriveFileIDs[j+(i*len(album))]

				song, err := h.songService.FindOneByDriveFileID(foundDriveFileID)
				if err != nil {
					return err
				}

				song.PDF.TgFileID = responses[j].Document.FileID

				if song.HasOutdatedPDF() || song.PDF.TgChannelMessageID == 0 {
					song = helpers.SendToChannel(h.bot, song)
				}

				song.PDF.ModifiedTime = song.DriveFile.ModifiedTime

				_, _ = h.songService.UpdateOne(*song)
			}
		}

		for _, messageID := range user.State.Context.MessagesToDelete {
			h.bot.Delete(&telebot.Message{
				ID:   messageID,
				Chat: c.Chat(),
			})
		}

		user.State = user.State.Prev
		user.State.Index = 0

		return h.enter(c, user)
	})

	return helpers.SetlistState, handlerFunc
}

func chunkAlbumBy(items telebot.Album, chunkSize int) (chunks []telebot.Album) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}
