package handlers

import (
	"errors"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/telebot/v3"
	"github.com/klauspost/lctime"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/drive/v3"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func mainMenuHandler() (int, []HandlerFunc) {
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

		case helpers.Schedule:
			user.State = &entities.State{
				Name: helpers.GetEventsState,
			}

		case helpers.Songs:
			return c.Send(helpers.Songs+":", &telebot.ReplyMarkup{
				ReplyKeyboard: [][]telebot.ReplyButton{
					{
						{Text: helpers.AllSongs},
					},
					{
						{Text: helpers.SongsByLastDateOfPerforming},
					},
					{
						{Text: helpers.SongsByNumberOfPerforming},
					},
					{
						{Text: helpers.CreateDoc},
					},
				},
				ResizeKeyboard: true,
			})

		case helpers.AllSongs:
			user.State = &entities.State{
				Name: helpers.SearchSongState,
			}

		case helpers.CreateDoc:
			user.State = &entities.State{
				Name: helpers.CreateSongState,
			}

		case helpers.Members:
			users, err := h.userService.FindManyExtraByBandID(user.BandID)
			if err != nil {
				return err
			}

			usersStr := ""
			event, err := h.eventService.FindOneOldestByBandID(user.BandID)
			if err == nil {
				usersStr = fmt.Sprintf("Статистика ведется с %s", lctime.Strftime("%d %B, %Y", event.Time))
			}

			for _, user := range users {
				if user.User == nil || user.User.Name == "" {
					continue
				}

				usersStr = fmt.Sprintf("%s\n\n<b><a href=\"tg://user?id=%d\">%s</a></b>\nВсего участий: %d", usersStr, user.User.ID, user.User.Name, len(user.Events))

				if len(user.Events) > 0 {
					usersStr = fmt.Sprintf("%s\nИз них:", usersStr)
				}

				mp := map[entities.Role]int{}

				for _, event := range user.Events {
					for _, membership := range event.Memberships {
						if membership.UserID == user.User.ID {
							mp[*membership.Role]++
							break
						}
					}
				}

				for role, num := range mp {
					usersStr = fmt.Sprintf("%s\n - %s: %d", usersStr, role.Name, num)
				}
			}

			return c.Send(usersStr, telebot.ModeHTML)

		case helpers.Settings:
			return c.Send(helpers.Settings+":", &telebot.ReplyMarkup{
				ReplyKeyboard:  helpers.SettingsKeyboard,
				ResizeKeyboard: true,
			})

		case helpers.BandSettings:
			err := c.Send(helpers.BandSettings+":", &telebot.ReplyMarkup{
				ResizeKeyboard: true,
				ReplyKeyboard: [][]telebot.ReplyButton{
					{
						{Text: helpers.CreateRole}, {Text: helpers.AddAdmin},
					},
					{{Text: helpers.Back}},
				},
			})
			if err != nil {
				return err
			}
			user.State.Index++
			return nil

		case helpers.ProfileSettings:
			err := c.Send(helpers.ProfileSettings+":", &telebot.ReplyMarkup{
				ResizeKeyboard: true,
				ReplyKeyboard: [][]telebot.ReplyButton{
					{
						{Text: helpers.ChangeBand},
					},
					{{Text: helpers.Back}},
				},
			})
			if err != nil {
				return err
			}
			user.State.Index++
			return nil

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

		case helpers.AddAdmin:
			user.State = &entities.State{
				Name: helpers.AddBandAdminState,
			}
		}

		return h.enter(c, user)
	})

	return helpers.MainMenuState, handlerFuncs
}

func createRoleHandler() (int, []HandlerFunc) {

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

		user.State.Context.Role = &entities.Role{
			Name: c.Text(),
		}

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		if len(user.Band.Roles) == 0 {
			user.State.Context.Role.Priority = 1
			user.State.Index++
			return nil
		}

		for _, role := range user.Band.Roles {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: role.Name}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err := c.Send("После какой роли должна быть эта роль?", markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		if user.State.Context.Role.Priority == 0 {

			var foundRole *entities.Role
			for _, role := range user.Band.Roles {
				if c.Text() == role.Name {
					foundRole = role
					break
				}
			}

			if foundRole == nil {
				user.State.Index--
				return h.enter(c, user)
			}

			user.State.Context.Role.Priority = foundRole.Priority + 1

			for _, role := range user.Band.Roles {
				if role.Priority > foundRole.Priority {
					role.Priority++
					h.roleService.UpdateOne(*role)
				}
			}
		}

		role, err := h.roleService.UpdateOne(
			entities.Role{
				Name:     user.State.Context.Role.Name,
				BandID:   user.BandID,
				Priority: user.State.Context.Role.Priority,
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

func getEventsHandler() (int, []HandlerFunc) {

	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		events, err := h.eventService.FindManyFromTodayByBandID(user.BandID)

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		for _, event := range events {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: event.Alias()}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.GetEventsWithMe}, {Text: helpers.GetAllEvents}})
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

		case helpers.GetEventsWithMe:
			events, err := h.eventService.FindManyFromTodayByBandIDAndUserID(user.BandID, user.ID)
			if err != nil {
				return err
			}

			for _, event := range events {
				eventString, _, err := h.eventService.ToHtmlStringByID(event.ID)
				if err != nil {
					continue
				}

				q := user.State.CallbackData.Query()
				q.Set("eventId", event.ID.Hex())
				user.State.CallbackData.RawQuery = q.Encode()

				err = c.Send(helpers.AddCallbackData(eventString, user.State.CallbackData.String()),
					&telebot.ReplyMarkup{
						InlineKeyboard: helpers.GetEventActionsKeyboard(*user, *event),
					}, telebot.ModeHTML, telebot.NoPreview)
				if err != nil {
					return err
				}
			}

			return nil

		case helpers.GetAllEvents:
			events, err := h.eventService.FindManyFromTodayByBandID(user.BandID)
			if err != nil {
				return err
			}

			for _, event := range events {
				eventString, _, err := h.eventService.ToHtmlStringByID(event.ID)
				if err != nil {
					continue
				}

				q := user.State.CallbackData.Query()
				q.Set("eventId", event.ID.Hex())
				user.State.CallbackData.RawQuery = q.Encode()

				err = c.Send(helpers.AddCallbackData(eventString, user.State.CallbackData.String()),
					&telebot.ReplyMarkup{
						InlineKeyboard: helpers.GetEventActionsKeyboard(*user, *event),
					}, telebot.ModeHTML, telebot.NoPreview)
				if err != nil {
					return err
				}
			}

			return nil

		default:
			c.Notify(telebot.Typing)

			regex := regexp.MustCompile(`.* \| (\d{2}\.\d{2}\.\d{4}) \| (.*)`)
			matches := regex.FindStringSubmatch(c.Text())
			if len(matches) < 3 {
				user.State = &entities.State{
					Name: helpers.SearchSongState,
				}
				return h.enter(c, user)
			}

			eventTime, err := time.Parse("02.01.2006", strings.TrimSpace(matches[1]))
			if err != nil {
				user.State = &entities.State{
					Name: helpers.SearchSongState,
				}
				return h.enter(c, user)
			}

			foundEvent, err := h.eventService.FindOneByNameAndTime(strings.TrimSpace(matches[2]), eventTime)
			if err != nil {
				user.State.Index--
				return h.enter(c, user)
			}

			user.State = &entities.State{
				Name: helpers.EventActionsState,
				Context: entities.Context{
					EventID: foundEvent.ID,
				},
				Prev: user.State,
			}
			user.State.Prev.Index = 1
			return h.enter(c, user)
		}
	})

	return helpers.GetEventsState, handlerFuncs
}

func createEventHandler() (int, []HandlerFunc) {

	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		err := c.Send("Введи название этого собрания:", &telebot.ReplyMarkup{
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

		markup := &telebot.ReplyMarkup{}

		now := time.Now()
		var monthFirstDayDate time.Time
		var monthLastDayDate time.Time

		if c.Callback() != nil {
			_, _, monthFirstDateStr := helpers.ParseCallbackData(c.Callback().Data)

			monthFirstDayDate, _ = time.Parse(time.RFC3339, monthFirstDateStr)
			monthLastDayDate = monthFirstDayDate.AddDate(0, 1, -1)
		} else {
			user.State.Context.Map = map[string]string{"eventName": c.Text()}

			monthFirstDayDate = time.Now().AddDate(0, 0, -now.Day()+1)
			monthLastDayDate = time.Now().AddDate(0, 1, -now.Day())
		}

		currCol := 4
		colNum := 4
		for d := monthFirstDayDate; d.After(monthLastDayDate) == false; d = d.AddDate(0, 0, 1) {
			timeStr := lctime.Strftime("%d %a", d)

			if now.Day() == d.Day() && now.Month() == d.Month() && now.Year() == d.Year() {
				timeStr = helpers.Today
			}
			if currCol == colNum {
				markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{})
				currCol = 0
			}

			markup.InlineKeyboard[len(markup.InlineKeyboard)-1] =
				append(markup.InlineKeyboard[len(markup.InlineKeyboard)-1], telebot.InlineButton{
					Text: timeStr,
					Data: helpers.AggregateCallbackData(helpers.CreateEventState, 2, d.Format(time.RFC3339)),
				})
			currCol++
		}

		prevMonthLastDate := monthFirstDayDate.AddDate(0, 0, -1)
		prevMonthFirstDateStr := prevMonthLastDate.AddDate(0, 0, -prevMonthLastDate.Day()+1).Format(time.RFC3339)
		nextMonthFirstDate := monthLastDayDate.AddDate(0, 0, 1)
		nextMonthFirstDateStr := monthLastDayDate.AddDate(0, 0, 1).Format(time.RFC3339)
		markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{
			{
				Text: lctime.Strftime("%B", prevMonthLastDate),
				Data: helpers.AggregateCallbackData(helpers.CreateEventState, 1, prevMonthFirstDateStr),
			},
			{
				Text: lctime.Strftime("%B", nextMonthFirstDate),
				Data: helpers.AggregateCallbackData(helpers.CreateEventState, 1, nextMonthFirstDateStr),
			},
		})

		msg := fmt.Sprintf("Выбери дату:\n\n<b>%s</b>", lctime.Strftime("%B %Y", monthFirstDayDate))
		if c.Callback() != nil {
			c.Edit(msg, markup, telebot.ModeHTML)
			c.Respond()
		} else {
			c.Send(msg, markup, telebot.ModeHTML)
		}

		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		_, _, eventTime := helpers.ParseCallbackData(c.Callback().Data)

		parsedTime, err := time.Parse(time.RFC3339, eventTime)
		if err != nil {
			user.State = &entities.State{Name: helpers.CreateEventState}
			return h.enter(c, user)
		}

		event, err := h.eventService.UpdateOne(entities.Event{
			Time:   parsedTime,
			Name:   user.State.Context.Map["eventName"],
			BandID: user.BandID,
		})
		if err != nil {
			return err
		}

		c.Callback().Data = helpers.AggregateCallbackData(helpers.EventActionsState, 0, "")
		q := user.State.CallbackData.Query()
		q.Set("eventId", event.ID.Hex())
		user.State.CallbackData.RawQuery = q.Encode()

		h.enterInlineHandler(c, user)

		user.State = &entities.State{
			Name: helpers.GetEventsState,
		}
		return h.enterReplyHandler(c, user)
	})

	return helpers.CreateEventState, handlerFuncs
}

func eventActionsHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		var eventID primitive.ObjectID
		if c.Callback() != nil {
			eventIDFromCallback, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
			if err != nil {
				return err
			}
			eventID = eventIDFromCallback
		} else {
			eventID = user.State.Context.EventID
		}

		eventString, event, err := h.eventService.ToHtmlStringByID(eventID)
		if err != nil {
			return err
		}

		options := &telebot.SendOptions{
			ReplyMarkup: &telebot.ReplyMarkup{
				InlineKeyboard: helpers.GetEventActionsKeyboard(*user, *event),
			},
			DisableWebPagePreview: true,
			ParseMode:             telebot.ModeHTML,
		}

		q := user.State.CallbackData.Query()
		q.Set("eventId", eventID.Hex())
		q.Del("index")
		q.Del("driveFileIds")
		user.State.CallbackData.RawQuery = q.Encode()

		if c.Callback() != nil {
			c.Edit(helpers.AddCallbackData(eventString, user.State.CallbackData.String()), options)
			c.Respond()
			return nil
		} else {
			err := c.Send(helpers.AddCallbackData(eventString, user.State.CallbackData.String()), options)
			if err != nil {
				return err
			}
			if user.State.Next != nil {
				user.State = user.State.Next
				return h.enter(c, user)
			} else {
				user.State = user.State.Prev
				return nil
			}
		}
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		event, err := h.eventService.FindOneByID(eventID)
		if err != nil {
			return err
		}

		var driveFileIDs []string
		for _, song := range event.Songs {
			driveFileIDs = append(driveFileIDs, song.DriveFileID)
		}

		err = sendDriveFilesAlbum(h, c, user, driveFileIDs)
		if err != nil {
			return err
		}

		c.Respond()

		return nil
	})

	return helpers.EventActionsState, handlerFuncs
}

func changeSongOrderHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, chosenDriveFileID := helpers.ParseCallbackData(c.Callback().Data)

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		if user.State.CallbackData.Query().Get("index") == "" {

			event, err := h.eventService.FindOneByID(eventID)
			if err != nil {
				return err
			}

			q := user.State.CallbackData.Query()
			for _, song := range event.Songs {
				q.Add("driveFileIds", song.DriveFileID)
			}

			q.Set("index", "-1")
			user.State.CallbackData.RawQuery = q.Encode()
		}

		songIndex, err := strconv.Atoi(user.State.CallbackData.Query().Get("index"))
		if err != nil {
			return err
		}

		if chosenDriveFileID != "" {
			q := user.State.CallbackData.Query()
			for i, driveFileID := range user.State.CallbackData.Query()["driveFileIds"] {
				if driveFileID == chosenDriveFileID {
					q["driveFileIds"] = append(q["driveFileIds"][:i], q["driveFileIds"][i+1:]...)
					user.State.CallbackData.RawQuery = q.Encode()
					break
				}
			}

			song, _, err := h.songService.FindOrCreateOneByDriveFileID(chosenDriveFileID)
			if err != nil {
				return err
			}

			err = h.eventService.ChangeSongIDPosition(eventID, song.ID, songIndex)
			if err != nil {
				return err
			}
		}

		if len(user.State.CallbackData.Query()["driveFileIds"]) == 0 {
			c.Callback().Data = helpers.AggregateCallbackData(helpers.EventActionsState, 0, "")
			return h.enter(c, user)
		}

		markup := &telebot.ReplyMarkup{}

		start := time.Now()
		songsStr, _, err := h.eventService.GetSongsAsHTMLStringByID(eventID)
		log.Printf("getting songs for event took %v", time.Since(start))

		songs, driveFiles, err := h.songService.FindOrCreateManyByDriveFileIDs(user.State.CallbackData.Query()["driveFileIds"])
		if err != nil {
			return err
		}

		for i, driveFile := range driveFiles {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: fmt.Sprintf("%s (%s)", driveFile.Name, songs[i].Caption()), Data: helpers.AggregateCallbackData(state, index, driveFile.Id)}})
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: helpers.End, Data: helpers.AggregateCallbackData(helpers.EventActionsState, 0, "")}})

		q := user.State.CallbackData.Query()
		q.Set("index", strconv.Itoa(songIndex+1))
		user.State.CallbackData.RawQuery = q.Encode()

		c.Edit(helpers.AddCallbackData(fmt.Sprintf("%s\nВыбери песню номер %d:", songsStr, songIndex+2),
			user.State.CallbackData.String()), markup, telebot.ModeHTML, telebot.NoPreview)
		c.Respond()
		return nil
	})

	return helpers.ChangeSongOrderState, handlerFuncs
}

func addEventMemberHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, payload := helpers.ParseCallbackData(c.Callback().Data)

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		event, err := h.eventService.FindOneByID(eventID)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{}

		for _, role := range event.Band.Roles {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: role.Name, Data: helpers.AggregateCallbackData(state, index+1, fmt.Sprintf("%s", role.ID.Hex()))}})
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: helpers.Back, Data: helpers.AggregateCallbackData(helpers.EventActionsState, 0, payload)}})

		c.Edit(markup)
		c.Respond()
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, roleIDHex := helpers.ParseCallbackData(c.Callback().Data)

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		event, err := h.eventService.FindOneByID(eventID)
		if err != nil {
			return err
		}

		roleID, err := primitive.ObjectIDFromHex(roleIDHex)
		if err != nil {
			return err
		}

		usersExtra, err := h.userService.FindManyByBandIDAndRoleID(event.BandID, roleID)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{}

		for _, userExtra := range usersExtra {
			var buttonText string
			if len(userExtra.Events) == 0 {
				buttonText = userExtra.User.Name
			} else {
				buttonText = fmt.Sprintf("%s | %v | %d", userExtra.User.Name, lctime.Strftime("%d %b", userExtra.Events[0].Time), len(userExtra.Events))
			}
			markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{
				{Text: buttonText, Data: helpers.AggregateCallbackData(state, index+1, fmt.Sprintf("%s:%d", roleIDHex, userExtra.User.ID))},
			})
		}

		// TODO
		markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: helpers.Cancel, Data: helpers.AggregateCallbackData(helpers.EventActionsState, 0, roleIDHex)}})

		c.Edit(markup)
		c.Respond()
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		_, _, payload := helpers.ParseCallbackData(c.Callback().Data)

		parsedPayload := strings.Split(payload, ":")

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		userID, err := strconv.ParseInt(parsedPayload[1], 10, 0)
		if err != nil {
			return err
		}

		roleID, err := primitive.ObjectIDFromHex(parsedPayload[0])
		if err != nil {
			return err
		}

		_, err = h.membershipService.UpdateOne(entities.Membership{
			EventID: eventID,
			UserID:  userID,
			RoleID:  roleID,
		})
		if err != nil {
			return err
		}

		//go func() {
		//	eventString, _ := h.eventService.ToHtmlStringByID(event.ID)
		//	h.bot.Send(telebot.ChatID(foundUser.ID),
		//		fmt.Sprintf("Привет. Ты учавствуешь в собрании! "+
		//			"Вот план:\n\n%s", eventString), telebot.ModeHTML, telebot.NoPreview)
		//}()

		eventString, event, err := h.eventService.ToHtmlStringByID(eventID)
		if err != nil {
			return err
		}

		q := user.State.CallbackData.Query()
		q.Set("eventId", eventID.Hex())
		user.State.CallbackData.RawQuery = q.Encode()

		c.Edit(helpers.AddCallbackData(eventString, user.State.CallbackData.String()), &telebot.ReplyMarkup{
			InlineKeyboard: helpers.GetEventActionsKeyboard(*user, *event),
		}, telebot.ModeHTML, telebot.NoPreview)
		c.Respond()
		return nil
	})

	return helpers.AddEventMemberState, handlerFuncs
}

func deleteEventMemberHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, payload := helpers.ParseCallbackData(c.Callback().Data)

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		event, err := h.eventService.FindOneByID(eventID)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{}

		for _, membership := range event.Memberships {
			user, err := h.userService.FindOneByID(membership.UserID)
			if err != nil {
				continue
			}

			markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: user.Name + " | " + membership.Role.Name, Data: helpers.AggregateCallbackData(state, index+1, fmt.Sprintf("%s", membership.ID.Hex()))}})
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: helpers.Back, Data: helpers.AggregateCallbackData(helpers.EventActionsState, 0, payload)}})

		c.Edit(markup)
		c.Respond()
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		_, _, membershipHex := helpers.ParseCallbackData(c.Callback().Data)

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		membershipID, err := primitive.ObjectIDFromHex(membershipHex)
		if err != nil {
			return err
		}

		err = h.membershipService.DeleteOneByID(membershipID)
		if err != nil {
			return err
		}

		//go func() {
		//	eventString, _ := h.eventService.ToHtmlStringByID(event.ID)
		//	h.bot.Send(telebot.ChatID(foundUser.ID),
		//		fmt.Sprintf("Привет. Ты учавствуешь в собрании! "+
		//			"Вот план:\n\n%s", eventString), telebot.ModeHTML, telebot.NoPreview)
		//}()

		eventString, event, err := h.eventService.ToHtmlStringByID(eventID)
		if err != nil {
			return err
		}

		q := user.State.CallbackData.Query()
		q.Set("eventId", eventID.Hex())
		user.State.CallbackData.RawQuery = q.Encode()

		c.Edit(helpers.AddCallbackData(eventString, user.State.CallbackData.String()), &telebot.ReplyMarkup{
			InlineKeyboard: helpers.GetEventActionsKeyboard(*user, *event),
		}, telebot.ModeHTML, telebot.NoPreview)
		c.Respond()
		return nil
	})

	return helpers.DeleteEventMemberState, handlerFuncs
}

func addEventSongHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		var eventID primitive.ObjectID
		if c.Callback() != nil {
			eventIDFromCallback, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
			if err != nil {
				return err
			}
			eventID = eventIDFromCallback
			c.Respond()
		} else {
			eventID = user.State.Context.EventID
		}

		err := c.Send("Введи название песни:", &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.End}}},
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State = &entities.State{
			Index: 1,
			Name:  helpers.AddEventSongState,
			Context: entities.Context{
				EventID: eventID,
			},
		}
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		if c.Text() == helpers.End {
			user.State = &entities.State{
				Name:    helpers.EventActionsState,
				Context: user.State.Context,
			}
			user.State.Next = &entities.State{
				Name: helpers.GetEventsState,
			}
			return h.enter(c, user)
		}

		c.Notify(telebot.Typing)

		query := helpers.CleanUpQuery(c.Text())

		driveFiles, _, err := h.driveFileService.FindSomeByFullTextAndFolderID(query, user.Band.DriveFolderID, "")
		if err != nil {
			return err
		}

		if len(driveFiles) == 0 {
			return c.Send(fmt.Sprintf("По запросу \"%s\" ничего не найдено.", c.Text()), &telebot.ReplyMarkup{
				ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.End}}},
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
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.End}})

		err = c.Send(fmt.Sprintf("Выбери песню по запросу \"%s\" или введи другое название:", c.Text()), markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		if c.Text() == helpers.End {
			user.State = &entities.State{
				Name:    helpers.EventActionsState,
				Context: user.State.Context,
			}
			user.State.Next = &entities.State{
				Name: helpers.GetEventsState,
			}
			return h.enter(c, user)
		}

		c.Notify(telebot.Typing)

		foundDriveFile, err := h.driveFileService.FindOneByNameAndFolderID(c.Text(), user.Band.DriveFolderID)
		if err != nil {
			user.State.Index--
			return h.enter(c, user)
		}

		song, _, err := h.songService.FindOrCreateOneByDriveFileID(foundDriveFile.Id)
		if err != nil {
			return err
		}

		err = h.eventService.PushSongID(user.State.Context.EventID, song.ID)
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.Send("Вероятнее всего, эта песня уже есть в списке.")
		} else if err != nil {
			return err
		}

		user.State.Index = 0
		return h.enter(c, user)
	})

	return helpers.AddEventSongState, handlerFuncs
}

func deleteEventSongHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, payload := helpers.ParseCallbackData(c.Callback().Data)

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		event, err := h.eventService.FindOneByID(eventID)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{}

		for _, song := range event.Songs {
			driveFile, err := h.driveFileService.FindOneByID(song.DriveFileID)
			if err != nil {
				continue
			}

			markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: driveFile.Name, Data: helpers.AggregateCallbackData(state, index+1, song.ID.Hex())}})
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{{Text: helpers.Back, Data: helpers.AggregateCallbackData(helpers.EventActionsState, 0, payload)}})

		c.Edit(markup)
		c.Respond()
		return nil
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {

		_, _, songIDHex := helpers.ParseCallbackData(c.Callback().Data)

		eventID, err := primitive.ObjectIDFromHex(user.State.CallbackData.Query().Get("eventId"))
		if err != nil {
			return err
		}

		songID, err := primitive.ObjectIDFromHex(songIDHex)
		if err != nil {
			return err
		}

		err = h.eventService.PullSongID(eventID, songID)
		if err != nil {
			return err
		}

		//go func() {
		//	eventString, _ := h.eventService.ToHtmlStringByID(event.ID)
		//	h.bot.Send(telebot.ChatID(foundUser.ID),
		//		fmt.Sprintf("Привет. Ты учавствуешь в собрании! "+
		//			"Вот план:\n\n%s", eventString), telebot.ModeHTML, telebot.NoPreview)
		//}()

		eventString, event, err := h.eventService.ToHtmlStringByID(eventID)
		if err != nil {
			return err
		}

		q := user.State.CallbackData.Query()
		q.Set("eventId", eventID.Hex())
		user.State.CallbackData.RawQuery = q.Encode()

		c.Edit(helpers.AddCallbackData(eventString, user.State.CallbackData.String()), &telebot.ReplyMarkup{
			InlineKeyboard: helpers.GetEventActionsKeyboard(*user, *event),
		}, telebot.ModeHTML, telebot.NoPreview)
		c.Respond()
		return nil
	})

	return helpers.DeleteEventSongState, handlerFuncs
}

func deleteEventHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		err := c.Send("Ты уверен?", &telebot.ReplyMarkup{
			ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: helpers.No}, {Text: helpers.Yes}}},
			ResizeKeyboard: true,
		})
		if err != nil {
			return err
		}

		user.State.Index++
		return err
	})

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
		if c.Text() == helpers.Yes {
			err := h.eventService.DeleteOneByID(user.State.Context.EventID)
			if err != nil {
				return err
			}

			err = c.Send("Удаление завершено.")
			if err != nil {
				return err
			}
			user.State = &entities.State{
				Name: helpers.GetEventsState,
			}
			return h.enter(c, user)
		} else {
			err := c.Send("Удаление отменено.")
			if err != nil {
				return err
			}
			user.State = user.State.Prev
			return h.enter(c, user)
		}
	})

	return helpers.DeleteEventState, handlerFuncs
}

func chooseBandHandler() (int, []HandlerFunc) {
	handlerFuncs := make([]HandlerFunc, 0)

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
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

	handlerFuncs = append(handlerFuncs, func(h *Handler, c telebot.Context, user *entities.User) error {
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

	return helpers.ChooseBandState, handlerFuncs
}

func createBandHandler() (int, []HandlerFunc) {
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

func addBandAdminHandler() (int, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		markup := &telebot.ReplyMarkup{
			ResizeKeyboard: true,
		}

		users, err := h.userService.FindMultipleByBandID(user.BandID)
		if err != nil {
			return err
		}

		for _, user := range users {
			markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: user.Name}})
		}
		markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Cancel}})

		err = c.Send("Выбери пользователя, которого ты хочешь сделать администратором:", markup)
		if err != nil {
			return err
		}

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		chosenUser, err := h.userService.FindOneByName(c.Text())
		if err != nil {
			user.State.Index--
			return h.enter(c, user)
		}

		chosenUser.Role = helpers.Admin
		_, err = h.userService.UpdateOne(*chosenUser)
		if err != nil {
			return err
		}

		err = c.Send(fmt.Sprintf("Пользователь %s повышен до администратора.", chosenUser.Name))
		if err != nil {
			return err
		}

		user.State = &entities.State{
			Name: helpers.MainMenuState,
		}

		return h.enter(c, user)
	})

	return helpers.AddBandAdminState, handlerFunc
}

func searchSongHandler() (int, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	// Print list of found songs.
	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		{
			c.Notify(telebot.Typing)

			var query string
			switch c.Text() {
			case helpers.SearchEverywhere, helpers.AllSongs:
				user.State.Context.QueryType = c.Text()
				query = user.State.Context.Query
			case helpers.PrevPage, helpers.NextPage:
				query = user.State.Context.Query
			default:
				user.State.Context.NextPageToken = nil
				query = c.Text()
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
			var nextPageToken string
			var err error

			if c.Text() == helpers.PrevPage {
				if user.State.Context.NextPageToken != nil &&
					user.State.Context.NextPageToken.PrevPageToken != nil {
					user.State.Context.NextPageToken = user.State.Context.NextPageToken.PrevPageToken.PrevPageToken
				}
			}

			if user.State.Context.NextPageToken == nil {
				user.State.Context.NextPageToken = &entities.NextPageToken{}
			}

			switch user.State.Context.QueryType {
			case helpers.SearchEverywhere:
				_driveFiles, _nextPageToken, _err := h.driveFileService.FindSomeByFullTextAndFolderID(query, "", user.State.Context.NextPageToken.Token)
				driveFiles = _driveFiles
				nextPageToken = _nextPageToken
				err = _err
			case helpers.AllSongs:
				_driveFiles, _nextPageToken, _err := h.driveFileService.FindAllByFolderID(user.Band.DriveFolderID, user.State.Context.NextPageToken.Token)
				driveFiles = _driveFiles
				nextPageToken = _nextPageToken
				err = _err
			default:
				_driveFiles, _nextPageToken, _err := h.driveFileService.FindSomeByFullTextAndFolderID(query, user.Band.DriveFolderID, user.State.Context.NextPageToken.Token)
				driveFiles = _driveFiles
				nextPageToken = _nextPageToken
				err = _err
			}

			if err != nil {
				return err
			}

			user.State.Context.NextPageToken = &entities.NextPageToken{
				Token:         nextPageToken,
				PrevPageToken: user.State.Context.NextPageToken,
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

			for _, song := range driveFiles {
				markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: song.Name}})
			}

			if c.Text() != helpers.SearchEverywhere || c.Text() != helpers.AllSongs {
				markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.SearchEverywhere}})
			}

			if user.State.Context.NextPageToken.Token != "" {
				if user.State.Context.NextPageToken.PrevPageToken != nil && user.State.Context.NextPageToken.PrevPageToken.Token != "" {
					markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.PrevPage}, {Text: helpers.Menu}, {Text: helpers.NextPage}})
				} else {
					markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Menu}, {Text: helpers.NextPage}})
				}
			} else {
				if user.State.Context.NextPageToken.PrevPageToken.Token != "" {
					markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.PrevPage}, {Text: helpers.Menu}})
				} else {
					markup.ReplyKeyboard = append(markup.ReplyKeyboard, []telebot.ReplyButton{{Text: helpers.Menu}, {Text: helpers.NextPage}})
				}
			}

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

		case helpers.SearchEverywhere, helpers.NextPage:
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

func songActionsHandler() (int, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		var driveFileID string

		if c.Callback() != nil {
			driveFileID = user.State.CallbackData.Query().Get("driveFileId")
		} else {
			c.Notify(telebot.UploadingDocument)
			driveFileID = user.State.Context.DriveFileID
		}

		err := SendDriveFileToUser(h, c, user, driveFileID)
		if err != nil {
			return err
		}

		if c.Callback() != nil {
			c.Respond()
		} else {
			if user.State.Next != nil {
				user.State = user.State.Next
				return h.enter(c, user)
			} else {
				user.State = user.State.Prev
				return nil
			}
		}
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		song, driveFile, err :=
			h.songService.FindOrCreateOneByDriveFileID(user.State.CallbackData.Query().Get("driveFileId"))
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{}
		markup.InlineKeyboard = helpers.GetSongActionsKeyboard(*user, *song, *driveFile)

		h.bot.EditReplyMarkup(c.Callback().Message, markup)
		c.Respond()
		return nil
	})

	//handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
	//
	//	song, err := h.songService.FindOneByDriveFileID(user.State.Context.DriveFileID)
	//	if err != nil {
	//		return err
	//	}
	//
	//	driveFile, err := h.driveFileService.FindOneByID(song.DriveFileID)
	//	if err != nil {
	//		return err
	//	}
	//
	//	switch c.Text() {

	//	case helpers.CopyToMyBand:
	//		user.State = &entities.State{
	//			Name: helpers.CopySongState,
	//			Context: entities.Context{
	//				DriveFileID: user.State.Context.DriveFileID,
	//			},
	//			Prev: user.State,
	//		}
	//		user.State.Prev.Index = 0
	//
	//	case helpers.Delete:
	//		user.State = &entities.State{
	//			Name: helpers.DeleteSongState,
	//			Context: entities.Context{
	//				DriveFileID: user.State.Context.DriveFileID,
	//			},
	//			Prev: user.State,
	//		}
	//		user.State.Prev.Index = 0
	//
	//	return h.enter(c, user)
	//})

	return helpers.SongActionsState, handlerFunc
}

func transposeSongHandler() (int, []HandlerFunc) {

	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, _ := helpers.ParseCallbackData(c.Callback().Data)

		err := c.EditCaption(helpers.AddCallbackData("Выбери новую тональность:", user.State.CallbackData.String()),
			&telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						{Text: "C | Am", Data: helpers.AggregateCallbackData(state, index+1, "C")},
						{Text: "C# | A#m", Data: helpers.AggregateCallbackData(state, index+1, "C#")},
						{Text: "Db | Bbm", Data: helpers.AggregateCallbackData(state, index+1, "Db")},
					},
					{
						{Text: "D | Bm", Data: helpers.AggregateCallbackData(state, index+1, "D")},
						{Text: "D# | Cm", Data: helpers.AggregateCallbackData(state, index+1, "D#")},
						{Text: "Eb | Cm", Data: helpers.AggregateCallbackData(state, index+1, "Eb")},
					},
					{
						{Text: "E | C#m", Data: helpers.AggregateCallbackData(state, index+1, "E")},
					},
					{
						{Text: "F | Dm", Data: helpers.AggregateCallbackData(state, index+1, "F")},
						{Text: "F# | D#m", Data: helpers.AggregateCallbackData(state, index+1, "F#")},
						{Text: "Gb | Ebm", Data: helpers.AggregateCallbackData(state, index+1, "Gb")},
					},
					{
						{Text: "G | Em", Data: helpers.AggregateCallbackData(state, index+1, "G")},
						{Text: "G# | Fm", Data: helpers.AggregateCallbackData(state, index+1, "G#")},
						{Text: "Ab | Fm", Data: helpers.AggregateCallbackData(state, index+1, "Ab")},
					},
					{
						{Text: "A | F#m", Data: helpers.AggregateCallbackData(state, index+1, "A")},
						{Text: "A# | Gm", Data: helpers.AggregateCallbackData(state, index+1, "A#")},
						{Text: "Bb | Gm", Data: helpers.AggregateCallbackData(state, index+1, "Bb")},
					},
					{
						{Text: "B | G#m", Data: helpers.AggregateCallbackData(state, index+1, "B")},
					},
					{
						{Text: helpers.Cancel, Data: helpers.AggregateCallbackData(helpers.SongActionsState, 0, "")},
					},
				},
			}, telebot.ModeHTML)
		if err != nil {
			return err
		}

		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, key := helpers.ParseCallbackData(c.Callback().Data)

		q := user.State.CallbackData.Query()
		q.Set("key", key)
		user.State.CallbackData.RawQuery = q.Encode()

		markup := &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					{Text: helpers.AppendSection, Data: helpers.AggregateCallbackData(state, index+1, "-1")},
				},
			},
		}

		sectionsNumber, err := h.driveFileService.GetSectionsNumber(user.State.CallbackData.Query().Get("driveFileId"))
		if err != nil {
			return err
		}

		for i := 0; i < sectionsNumber; i++ {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{
				{Text: fmt.Sprintf("Вместо %d-й секции", i+1), Data: helpers.AggregateCallbackData(state, index+1, fmt.Sprintf("%d", i))},
			})
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{
			{Text: helpers.Cancel, Data: helpers.AggregateCallbackData(helpers.SongActionsState, 0, "")},
		})

		c.EditCaption(helpers.AddCallbackData("Куда ты хочешь вставить новую тональность?", user.State.CallbackData.String()),
			markup, telebot.ModeHTML)

		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		_, _, sectionIndexStr := helpers.ParseCallbackData(c.Callback().Data)

		sectionIndex, _ := strconv.Atoi(sectionIndexStr)

		driveFile, err := h.driveFileService.TransposeOne(
			user.State.CallbackData.Query().Get("driveFileId"),
			user.State.CallbackData.Query().Get("key"),
			sectionIndex)
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

		c.Callback().Data = helpers.AggregateCallbackData(helpers.SongActionsState, 0, "")
		return h.enterInlineHandler(c, user)
	})

	return helpers.TransposeSongState, handlerFunc
}

func styleSongHandler() (int, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	// Print list of found songs.
	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		driveFileID := user.State.CallbackData.Query().Get("driveFileId")

		driveFile, err := h.driveFileService.StyleOne(driveFileID)
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

		//c.Respond()
		c.Callback().Data = helpers.AggregateCallbackData(helpers.SongActionsState, 0, "")
		return h.enterInlineHandler(c, user)
	})
	return helpers.StyleSongState, handlerFunc
}

func copySongHandler() (int, []HandlerFunc) {
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

func createSongHandler() (int, []HandlerFunc) {
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
			Next: &entities.State{
				Name: helpers.MainMenuState,
			},
		}

		return h.enter(c, user)
	})

	return helpers.CreateSongState, handlerFunc
}

func deleteSongHandler() (int, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		if user.Role == helpers.Admin {
			err := h.songService.DeleteOneByDriveFileID(user.State.CallbackData.Query().Get("driveFileId"))
			if err != nil {
				return err
			}

			c.EditCaption("Удалено")
		}

		return nil
	})

	return helpers.DeleteSongState, handlerFunc
}

func getVoicesHandler() (int, []HandlerFunc) {
	handlerFunc := make([]HandlerFunc, 0)

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, _ := helpers.ParseCallbackData(c.Callback().Data)

		song, driveFileID, err := h.songService.FindOrCreateOneByDriveFileID(user.State.CallbackData.Query().Get("driveFileId"))
		if err != nil {
			return err
		}

		if song.Voices == nil || len(song.Voices) == 0 {
			c.EditCaption(helpers.AddCallbackData("У этой песни нет партий. Чтобы добавить, отправь мне голосовое сообщение.",
				user.State.CallbackData.String()), &telebot.ReplyMarkup{
				InlineKeyboard: helpers.GetSongActionsKeyboard(*user, *song, *driveFileID),
			}, telebot.ModeHTML)
			return nil
		} else {
			markup := &telebot.ReplyMarkup{}

			for _, voice := range song.Voices {
				markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{
					{Text: voice.Name, Data: helpers.AggregateCallbackData(state, index+1, voice.ID.Hex())},
				})
			}
			markup.InlineKeyboard = append(markup.InlineKeyboard, []telebot.InlineButton{
				{Text: helpers.Back, Data: helpers.AggregateCallbackData(helpers.SongActionsState, 0, "")},
			})

			c.EditCaption(helpers.AddCallbackData("Выбери партию:", user.State.CallbackData.String()),
				markup, telebot.ModeHTML)

			return nil
		}
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		state, index, voiceIDHex := helpers.ParseCallbackData(c.Callback().Data)

		voiceID, err := primitive.ObjectIDFromHex(voiceIDHex)
		if err != nil {
			return err
		}

		voice, err := h.voiceService.FindOneByID(voiceID)
		if err != nil {
			return err
		}

		markup := &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					{Text: helpers.Back, Data: helpers.AggregateCallbackData(state, index-1, "")},
				},
				//{
				//	{Text: helpers.Delete},
				//},
			},
		}

		song, driveFile, err := h.songService.FindOrCreateOneByDriveFileID(user.State.CallbackData.Query().Get("driveFileId"))
		getPerformer := func() string {
			if driveFile != nil {
				return driveFile.Name
			} else {
				return ""
			}
		}
		getCaption := func() string {
			if song != nil {
				return song.Caption()
			} else {
				return "-"
			}
		}

		if voice.AudioFileID == "" {
			file, err := h.bot.File(&telebot.File{FileID: voice.FileID})
			if err != nil {
				return err
			}

			msg, err := h.bot.EditMedia(
				c.Callback().Message,
				&telebot.Audio{
					File:      telebot.FromReader(file),
					Title:     voice.Name,
					Performer: getPerformer(),
					Caption:   helpers.AddCallbackData(getCaption(), user.State.CallbackData.String()),
				},
				markup, telebot.ModeHTML)
			if err != nil {
				return c.Respond()
			}
			voice.AudioFileID = msg.Audio.FileID
			h.voiceService.UpdateOne(*voice)
		} else {
			h.bot.EditMedia(
				c.Callback().Message,
				&telebot.Audio{
					File:      telebot.File{FileID: voice.AudioFileID},
					Title:     voice.Name,
					Performer: getPerformer(),
					Caption:   helpers.AddCallbackData(getCaption(), user.State.CallbackData.String()),
				},
				markup, telebot.ModeHTML)
		}

		c.Respond()
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		switch c.Text() {
		case helpers.Back:
			user.State.Index = 0
			return h.enter(c, user)
		case helpers.DeleteMember:
			// TODO: handle delete
			return nil
		default:
			return c.Send("Я тебя не понимаю. Нажми на кнопку.")
		}
	})

	return helpers.GetVoicesState, handlerFunc
}

func uploadVoiceHandler() (int, []HandlerFunc) {
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

		driveFiles, _, err := h.driveFileService.FindSomeByFullTextAndFolderID(c.Text(), user.Band.DriveFolderID, "")
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

		user.State.Index++
		return nil
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		c.Notify(telebot.UploadingDocument)

		foundDriveFile, err := h.driveFileService.FindOneByNameAndFolderID(c.Text(), user.Band.DriveFolderID)
		if err != nil {
			user.State.Index--
			return h.enter(c, user)
		}

		song, _, err := h.songService.FindOrCreateOneByDriveFileID(foundDriveFile.Id)
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
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {

		user.State.Context.Voice.Name = c.Text()

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

func setlistHandler() (int, []HandlerFunc) {
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

		driveFiles, _, err := h.driveFileService.FindSomeByFullTextAndFolderID(currentSongName, user.Band.DriveFolderID, "")
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

		foundDriveFile, err := h.driveFileService.FindOneByNameAndFolderID(c.Text(), user.Band.DriveFolderID)
		if err != nil {
			user.State.Context.SongNames = append([]string{c.Text()}, user.State.Context.SongNames...)
		} else {
			user.State.Context.FoundDriveFileIDs = append(user.State.Context.FoundDriveFileIDs, foundDriveFile.Id)
		}

		user.State.Index = 0
		return h.enter(c, user)
	})

	handlerFunc = append(handlerFunc, func(h *Handler, c telebot.Context, user *entities.User) error {
		c.Notify(telebot.UploadingDocument)

		driveFileIDs := user.State.Context.FoundDriveFileIDs

		err := sendDriveFilesAlbum(h, c, user, driveFileIDs)
		if err != nil {
			return err
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
