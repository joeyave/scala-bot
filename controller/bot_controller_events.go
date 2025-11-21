package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *BotController) event(bot *gotgbot.Bot, ctx *ext.Context, event *entity.Event) error {

	user := ctx.Data["user"].(*entity.User)

	html := c.EventService.ToHtmlStringByEvent(*event, ctx.EffectiveUser.LanguageCode) // todo: refactor

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.EventInit(event, user, ctx.EffectiveUser.LanguageCode),
	}

	msg, err := ctx.EffectiveChat.SendMessage(bot, html, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	user.CallbackCache.ChatID = msg.Chat.Id
	user.CallbackCache.MessageID = msg.MessageId
	text := user.CallbackCache.AddToText(html)

	_, _, err = msg.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	return err
}

func (c *BotController) CreateEvent(bot *gotgbot.Bot, ctx *ext.Context) error {

	var event *entity.Event
	err := json.Unmarshal([]byte(ctx.EffectiveMessage.WebAppData.Data), &event)
	if err != nil {
		return err
	}

	user := ctx.Data["user"].(*entity.User)

	event.BandID = user.BandID

	createdEvent, err := c.EventService.UpdateOne(*event)
	if err != nil {
		return err
	}

	// todo: remove when added this as setting to band.
	if createdEvent.Band.Timezone == "" {
		createdEvent.Band.Timezone = event.Timezone
		_, err := c.BandService.UpdateOne(*createdEvent.Band)
		if err != nil {
			log.Info().Msgf("Error updating band timezone: %v", err)
		}
	}

	user.State.Index = 0
	err = c.event(bot, ctx, createdEvent)
	if err != nil {
		return err
	}
	err = c.GetEvents(0)(bot, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *BotController) GetEvents(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.GetEvents {
			user.State = entity.State{
				Index: index,
				Name:  state.GetEvents,
			}
			user.Cache = entity.Cache{}
		}

		switch index {
		case 0:
			{
				events, err := c.EventService.FindManyFromTodayByBandID(user.BandID, user.Band.GetLocation())
				if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
					return err
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
				}

				user.Cache.Buttons = keyboard.GetEventsStateFilterButtons(events, ctx.EffectiveUser.LanguageCode)
				markup.Keyboard = append(markup.Keyboard, user.Cache.Buttons)
				markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.createEvent", ctx.EffectiveUser.LanguageCode), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/webapp-react/#/events/create?bandId=%s&bandTimezone=%s&driveFolderId=%s&archiveFolderId=%s", os.Getenv("BOT_DOMAIN"), user.Band.ID.Hex(), user.Band.Timezone, user.Band.DriveFolderID, user.Band.ArchiveFolderID)}}})

				for _, event := range events {
					markup.Keyboard = append(markup.Keyboard, keyboard.EventButton(event, user, ctx.EffectiveUser.LanguageCode, false))
				}

				markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}})

				_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseEvent", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
				if err != nil {
					return err
				}

				user.State.Index = 1

				return nil
			}
		case 1:
			{
				switch ctx.EffectiveMessage.Text {
				case txt.Get("button.next", ctx.EffectiveUser.LanguageCode), txt.Get("button.prev", ctx.EffectiveUser.LanguageCode):
					return c.GetEvents(0)(bot, ctx)

				case txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode), txt.Get("button.archive", ctx.EffectiveUser.LanguageCode):
					return c.filterEvents(0)(bot, ctx)

				default:
					if keyboard.IsWeekdayButton(ctx.EffectiveMessage.Text) {
						return c.filterEvents(0)(bot, ctx)
					}
				}

				_, _ = ctx.EffectiveChat.SendAction(bot, "typing", nil)

				eventName, eventTime, err := keyboard.ParseEventButton(ctx.EffectiveMessage.Text, user.Band.GetLocation())
				if err != nil {
					return c.search(0)(bot, ctx)
				}

				foundEvent, err := c.EventService.FindOneByNameAndTimeAndBandID(eventName, eventTime, user.BandID)
				if err != nil {
					return c.search(0)(bot, ctx)
				}

				err = c.event(bot, ctx, foundEvent)
				return err
			}
		}
		return c.Menu(bot, ctx)
	}
}

func (c *BotController) filterEvents(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.FilterEvents {
			user.State = entity.State{
				Index: index,
				Name:  state.FilterEvents,
			}
			user.Cache = entity.Cache{
				Buttons: user.Cache.Buttons,
			}
		}

		switch index {
		case 0:
			{
				_, _ = ctx.EffectiveChat.SendAction(bot, "typing", nil)

				// todo: refactor - extract to func
				if (ctx.EffectiveMessage.Text == txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode) || ctx.EffectiveMessage.Text == txt.Get("button.archive", ctx.EffectiveUser.LanguageCode) ||
					keyboard.IsWeekdayButton(ctx.EffectiveMessage.Text)) && user.Cache.Filter != txt.Get("button.archive", ctx.EffectiveUser.LanguageCode) {
					user.Cache.Filter = ctx.EffectiveMessage.Text
				}

				var (
					events []*entity.Event
					err    error
				)

				bandLoc := user.Band.GetLocation()

				if user.Cache.Filter == txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode) {
					events, err = c.EventService.FindManyFromTodayByBandIDAndUserID(user.BandID, bandLoc, user.ID, user.Cache.PageIndex)
				} else if user.Cache.Filter == txt.Get("button.archive", ctx.EffectiveUser.LanguageCode) {
					if keyboard.IsWeekdayButton(ctx.EffectiveMessage.Text) {
						events, err = c.EventService.FindManyUntilTodayByBandIDAndWeekdayAndPageNumber(user.BandID, bandLoc, keyboard.ParseWeekdayButton(ctx.EffectiveMessage.Text), user.Cache.PageIndex)
						user.Cache.Query = ctx.EffectiveMessage.Text
					} else if keyboard.IsWeekdayButton(user.Cache.Query) && (ctx.EffectiveMessage.Text == txt.Get("button.next", ctx.EffectiveUser.LanguageCode) || ctx.EffectiveMessage.Text == txt.Get("button.prev", ctx.EffectiveUser.LanguageCode)) {
						events, err = c.EventService.FindManyUntilTodayByBandIDAndWeekdayAndPageNumber(user.BandID, bandLoc, keyboard.ParseWeekdayButton(user.Cache.Query), user.Cache.PageIndex)
					} else if ctx.EffectiveMessage.Text == txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode) {
						events, err = c.EventService.FindManyUntilTodayByBandIDAndUserIDAndPageNumber(user.BandID, bandLoc, user.ID, user.Cache.PageIndex)
						user.Cache.Query = ctx.EffectiveMessage.Text
					} else if user.Cache.Query == txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode) && (ctx.EffectiveMessage.Text == txt.Get("button.next", ctx.EffectiveUser.LanguageCode) || ctx.EffectiveMessage.Text == txt.Get("button.prev", ctx.EffectiveUser.LanguageCode)) {
						events, err = c.EventService.FindManyUntilTodayByBandIDAndUserIDAndPageNumber(user.BandID, bandLoc, user.ID, user.Cache.PageIndex)
					} else {
						events, err = c.EventService.FindManyUntilTodayByBandIDAndPageNumber(user.BandID, bandLoc, user.Cache.PageIndex)
						user.Cache.Buttons = keyboard.GetEventsStateFilterButtons(events, ctx.EffectiveUser.LanguageCode)
					}
				} else if keyboard.IsWeekdayButton(user.Cache.Filter) {
					events, err = c.EventService.FindManyFromTodayByBandIDAndWeekday(user.BandID, keyboard.ParseWeekdayButton(user.Cache.Filter), bandLoc)
				}
				if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
					return err
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
				}

				var buttons []gotgbot.KeyboardButton
				for _, button := range user.Cache.Buttons {

					if button.Text == user.Cache.Filter ||
						(button.Text == ctx.EffectiveMessage.Text && user.Cache.Filter == txt.Get("button.archive", ctx.EffectiveUser.LanguageCode)) ||
						(button.Text == user.Cache.Query && user.Cache.Filter == txt.Get("button.archive", ctx.EffectiveUser.LanguageCode) && (ctx.EffectiveMessage.Text == txt.Get("button.next", ctx.EffectiveUser.LanguageCode) || ctx.EffectiveMessage.Text == txt.Get("button.prev", ctx.EffectiveUser.LanguageCode))) {
						button = keyboard.SelectedButton(button.Text)
					}

					buttons = append(buttons, button)
				}

				markup.Keyboard = append(markup.Keyboard, buttons)
				markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.createEvent", ctx.EffectiveUser.LanguageCode), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/webapp-react/#/events/create?bandId=%s&bandTimezone=%s&driveFolderId=%s&archiveFolderId=%s", os.Getenv("BOT_DOMAIN"), user.Band.Timezone, user.Band.ID.Hex(), user.Band.DriveFolderID, user.Band.ArchiveFolderID)}}})

				for _, event := range events {
					if user.Cache.Filter == txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode) {
						markup.Keyboard = append(markup.Keyboard, keyboard.EventButton(event, user, ctx.EffectiveUser.LanguageCode, true))
					} else {
						markup.Keyboard = append(markup.Keyboard, keyboard.EventButton(event, user, ctx.EffectiveUser.LanguageCode, false))
					}
				}

				if user.Cache.PageIndex != 0 {
					markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.prev", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.next", ctx.EffectiveUser.LanguageCode)}})
				} else {
					markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.next", ctx.EffectiveUser.LanguageCode)}})
				}

				_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseEvent", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
				if err != nil {
					return err
				}

				user.State.Index = 1

				return nil
			}
		case 1:
			{
				switch ctx.EffectiveMessage.Text {
				case txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode), txt.Get("button.archive", ctx.EffectiveUser.LanguageCode):
					user.Cache.PageIndex = 0
					return c.filterEvents(0)(bot, ctx)
				case txt.Get("button.next", ctx.EffectiveUser.LanguageCode):
					user.Cache.PageIndex++
					return c.filterEvents(0)(bot, ctx)
				case txt.Get("button.prev", ctx.EffectiveUser.LanguageCode):
					user.Cache.PageIndex--
					return c.filterEvents(0)(bot, ctx)
				default:
					if keyboard.IsWeekdayButton(ctx.EffectiveMessage.Text) {
						user.Cache.PageIndex = 0
						return c.filterEvents(0)(bot, ctx)
					}
				}

				if keyboard.IsSelectedButton(ctx.EffectiveMessage.Text) {
					if user.Cache.Filter == txt.Get("button.archive", ctx.EffectiveUser.LanguageCode) {
						if keyboard.IsWeekdayButton(keyboard.ParseSelectedButton(ctx.EffectiveMessage.Text)) ||
							keyboard.ParseSelectedButton(ctx.EffectiveMessage.Text) == txt.Get("button.eventsWithMe", ctx.EffectiveUser.LanguageCode) {
							return c.filterEvents(0)(bot, ctx)
						} else {
							return c.GetEvents(0)(bot, ctx)
						}
					} else {
						return c.GetEvents(0)(bot, ctx)
					}
				}

				_, _ = ctx.EffectiveChat.SendAction(bot, "typing", nil)

				eventName, eventTime, err := keyboard.ParseEventButton(ctx.EffectiveMessage.Text, user.Band.GetLocation())
				if err != nil {
					return c.search(0)(bot, ctx)
				}

				foundEvent, err := c.EventService.FindOneByNameAndTimeAndBandID(eventName, eventTime, user.BandID)
				if err != nil {
					return c.GetEvents(0)(bot, ctx)
				}

				err = c.event(bot, ctx, foundEvent)
				return err
			}
		}

		return c.Menu(bot, ctx)
	}
}

// ------- Callback controllers -------

func (c *BotController) EventSetlistDocs(bot *gotgbot.Bot, ctx *ext.Context) error {

	eventIDHex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	eventID, err := primitive.ObjectIDFromHex(eventIDHex)
	if err != nil {
		return err
	}
	event, err := c.EventService.GetEventWithSongs(eventID)
	if err != nil {
		return err
	}

	var driveFileIDs []string
	for _, song := range event.Songs {
		driveFileIDs = append(driveFileIDs, song.DriveFileID)
	}

	if len(driveFileIDs) == 0 {
		_, err := ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: txt.Get("noSongs", ctx.EffectiveUser.LanguageCode),
		})
		if err != nil {
			return err
		}
		return nil
	}

	songs, _, err := c.SongService.FindOrCreateManyByDriveFileIDs(driveFileIDs)
	if err != nil {
		return err
	}

	err = c.songsAlbum(bot, ctx, songs)
	if err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

func (c *BotController) EventCB(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	hex := split[0]
	eventID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	html, event, err := c.EventService.ToHtmlStringByID(eventID, ctx.EffectiveUser.LanguageCode)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	if len(split) > 1 {
		switch split[1] {
		case "edit":
			markup.InlineKeyboard = keyboard.EventEdit(event, user, user.CallbackCache.ChatID, user.CallbackCache.MessageID, ctx.EffectiveUser.LanguageCode)
		default:
			markup.InlineKeyboard = keyboard.EventInit(event, user, ctx.EffectiveUser.LanguageCode)
		}
	}

	user.CallbackCache = entity.CallbackCache{
		MessageID: user.CallbackCache.MessageID,
		ChatID:    user.CallbackCache.ChatID,
	}
	text := user.CallbackCache.AddToText(html)

	_, _, err = ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: markup,
	})

	_, _ = ctx.CallbackQuery.Answer(bot, nil)
	return err
}

func (c *BotController) eventMembers(bot *gotgbot.Bot, ctx *ext.Context, event *entity.Event, memberships []*entity.Membership) error {

	user := ctx.Data["user"].(*entity.User)

	markup := gotgbot.InlineKeyboardMarkup{}

	for _, membership := range memberships {
		isDeleted := true
		for _, eventMembership := range event.Memberships {
			if eventMembership.ID == membership.ID {
				isDeleted = false
				break
			}
		}

		text := fmt.Sprintf("%s (%s)", membership.User.Name, membership.Role.Name)
		if isDeleted {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.EventMembersDeleteOrRecoverMember, event.ID.Hex()+":"+membership.ID.Hex()+":recover")}})
		} else {
			text += " ✅"
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.EventMembersDeleteOrRecoverMember, event.ID.Hex()+":"+membership.ID.Hex()+":delete")}})
		}
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.addMember", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.EventMembersAddMemberChooseRole, event.ID.Hex())}})
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.EventCB, event.ID.Hex()+":edit")}})

	text := fmt.Sprintf("<b>%s</b>\n\n%s:", event.Alias(ctx.EffectiveUser.LanguageCode), txt.Get("button.members", ctx.EffectiveUser.LanguageCode))
	text = user.CallbackCache.AddToText(text)

	_, _, err := ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(bot, nil)
	return nil
}

func (c *BotController) EventMembers(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	hex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	eventID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	event, err := c.EventService.FindOneByID(eventID)
	if err != nil {
		return err
	}

	membershipsJson, err := json.Marshal(event.Memberships)
	if err != nil {
		return err
	}

	user.CallbackCache.JsonString = string(membershipsJson)

	return c.eventMembers(bot, ctx, event, event.Memberships)
}

func (c *BotController) EventMembersDeleteOrRecoverMember(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	eventID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	membershipID, err := primitive.ObjectIDFromHex(split[1])
	if err != nil {
		return err
	}

	var cachedMemberships []*entity.Membership
	err = json.Unmarshal([]byte(user.CallbackCache.JsonString), &cachedMemberships)
	if err != nil {
		return err
	}

	switch split[2] {
	case "delete":
		membership, err := c.MembershipService.FindOneByID(membershipID)
		if err != nil {
			return err
		}

		// todo: return deleted membership
		err = c.MembershipService.DeleteOneByID(membershipID)
		if err != nil {
			return err
		}

		go c.notifyDeleted(bot, user, membership)
	case "recover":
		var membershipToRecover *entity.Membership
		for _, cachedMembership := range cachedMemberships {
			if membershipID == cachedMembership.ID {
				membershipToRecover = cachedMembership
				break
			}
		}

		membership, err := c.MembershipService.UpdateOne(*membershipToRecover)
		if err != nil {
			return err
		}

		go c.notifyAdded(bot, user, membership)
	}

	event, err := c.EventService.FindOneByID(eventID)
	if err != nil {
		return err
	}

	return c.eventMembers(bot, ctx, event, cachedMemberships)
}

func (c *BotController) EventMembersAddMemberChooseRole(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	hex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	eventID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	event, err := c.EventService.FindOneByID(eventID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	for _, role := range event.Band.Roles {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: role.Name, CallbackData: util.CallbackData(state.EventMembersAddMemberChooseUser, event.ID.Hex()+":"+role.ID.Hex())}})
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.createRole", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.RoleCreate_AskForName, user.Band.ID.Hex())}})
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.EventMembers, event.ID.Hex())}})

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>\n\n", event.Alias(ctx.EffectiveUser.LanguageCode))
	rolesString := event.RolesString()
	if rolesString != "" {
		fmt.Fprintf(&b, "%s\n\n", rolesString)
	}
	b.WriteString(txt.Get("text.chooseRoleForNewMember", ctx.EffectiveUser.LanguageCode))

	text := user.CallbackCache.AddToText(b.String())

	_, _, err = ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}
	_, _ = ctx.CallbackQuery.Answer(bot, nil)
	return nil
}

func (c *BotController) EventMembersAddMemberChooseUser(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	eventID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	roleID, err := primitive.ObjectIDFromHex(split[1])
	if err != nil {
		return err
	}

	loadMore := len(split) > 2 && split[2] == "more"

	err = c.eventMembersAddMemberChooseUser(bot, ctx, eventID, roleID, loadMore)
	if err != nil {
		return err
	}
	return nil
}

func (c *BotController) eventMembersAddMemberChooseUser(bot *gotgbot.Bot, ctx *ext.Context, eventID primitive.ObjectID, roleID primitive.ObjectID, loadMore bool) error {

	user := ctx.Data["user"].(*entity.User)

	event, err := c.EventService.FindOneByID(eventID)
	if err != nil {
		return err
	}

	role, err := c.RoleService.FindOneByID(roleID)
	if err != nil {
		return err
	}

	now := event.Band.GetNowTime()
	fromDate := time.Date(
		now.Year(), time.January, 1,
		0, 0, 0, 0,
		now.Location())
	usersWithEvents, err := c.UserService.FindManyByBandIDAndRoleID(event.BandID, roleID, fromDate)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	if !loadMore {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.loadMore", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.EventMembersAddMemberChooseUser, eventID.Hex()+":"+roleID.Hex()+":more")}})
	}

	for _, u := range usersWithEvents {
		var text string
		if len(u.Events) == 0 {
			text = u.Name
		} else {
			text = u.NameWithStats()
		}

		isMember := false
		var membership *entity.Membership
		for _, eventMembership := range event.Memberships {
			if eventMembership.RoleID == roleID && eventMembership.UserID == u.ID {
				isMember = true
				membership = eventMembership
				break
			}
		}

		if (len(u.Events) > 0 && u.Events[0].TimeUTC.After(time.Now().AddDate(0, -4, 0))) || loadMore {
			if isMember {
				text += " ✅"
				markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.EventMembersDeleteMember, roleID.Hex()+":"+membership.ID.Hex())}})
			} else {
				markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.EventMembersAddMember, roleID.Hex()+":"+strconv.FormatInt(u.ID, 10))}})
			}
		}
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.EventMembersAddMemberChooseRole, eventID.Hex())}})

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>\n\n", event.Alias(ctx.EffectiveUser.LanguageCode))
	rolesString := event.RolesString()
	if rolesString != "" {
		fmt.Fprintf(&b, "%s\n\n", rolesString)
	}
	b.WriteString(txt.Get("text.chooseNewMember", ctx.EffectiveUser.LanguageCode, role.Name))

	user.CallbackCache.EventIDHex = eventID.Hex()
	text := user.CallbackCache.AddToText(b.String())

	_, _, err = ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(bot, nil)
	return nil
}

func (c *BotController) EventMembersAddMember(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	roleID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil {
		return err
	}

	eventID, err := primitive.ObjectIDFromHex(user.CallbackCache.EventIDHex)
	if err != nil {
		return err
	}

	membership, err := c.MembershipService.UpdateOne(entity.Membership{
		EventID: eventID,
		UserID:  userID,
		RoleID:  roleID,
	})
	if err != nil {
		return err
	}

	go c.notifyAdded(bot, user, membership)

	return c.eventMembersAddMemberChooseUser(bot, ctx, eventID, roleID, false)
}

func (c *BotController) EventMembersDeleteMember(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	roleID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	membershipID, err := primitive.ObjectIDFromHex(split[1])
	if err != nil {
		return err
	}

	eventID, err := primitive.ObjectIDFromHex(user.CallbackCache.EventIDHex)
	if err != nil {
		return err
	}

	membership, err := c.MembershipService.FindOneByID(membershipID)
	if err != nil {
		return err
	}

	err = c.MembershipService.DeleteOneByID(membershipID)
	if err != nil {
		return err
	}

	go c.notifyDeleted(bot, user, membership)

	return c.eventMembersAddMemberChooseUser(bot, ctx, eventID, roleID, false)
}

func (c *BotController) EventDeleteConfirm(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	eventID, err := primitive.ObjectIDFromHex(payload)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	markup.InlineKeyboard = [][]gotgbot.InlineKeyboardButton{
		{
			{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.EventCB, eventID.Hex()+":edit")},
			{Text: txt.Get("button.yes", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.EventDelete, eventID.Hex())},
		},
	}

	text := user.CallbackCache.AddToText(txt.Get("text.eventDeleteConfirm", ctx.EffectiveUser.LanguageCode))

	_, _, err = ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ParseMode:   "HTML",
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *BotController) EventDelete(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	eventID, err := primitive.ObjectIDFromHex(payload)
	if err != nil {
		return err
	}

	err = c.EventService.DeleteOneByID(eventID)
	if err != nil {
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditText(bot, txt.Get("text.eventDeleted", ctx.EffectiveUser.LanguageCode), nil)
	if err != nil {
		return err
	}

	return c.GetEvents(0)(bot, ctx)
}

func (c *BotController) notifyAdded(bot *gotgbot.Bot, user *entity.User, membership *entity.Membership) {

	if user.ID == membership.UserID {
		return
	}

	// todo
	//time.Sleep(5 * time.Second)
	//
	//_, err := c.MembershipService.FindOneByID(membership.ID)
	//if err != nil {
	//	return
	//}

	event, err := c.EventService.FindOneByID(membership.EventID)
	if err != nil {
		return
	}

	todayStartUTC := helpers.GetStartOfDayInLocUTC(event.Band.GetLocation())
	if event.TimeUTC.After(todayStartUTC) {

		markup := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{{
				Text:         txt.Get("button.moreInfo", membership.User.LanguageCode),
				CallbackData: util.CallbackData(state.EventCB, event.ID.Hex()+":init"),
			}}},
		}

		text := txt.Get("text.memberAddedNotification", membership.User.LanguageCode,
			user.Name, membership.Role.Name, event.Alias(membership.User.LanguageCode))

		_, err := bot.SendMessage(membership.UserID, text, &gotgbot.SendMessageOpts{
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		})
		if err != nil {
			return
		}
	}
}

func (c *BotController) notifyDeleted(bot *gotgbot.Bot, user *entity.User, membership *entity.Membership) {

	if user.ID == membership.UserID {
		return
	}

	// todo
	//time.Sleep(5 * time.Second)
	//
	//_, err := c.MembershipService.FindSimilar(membership)
	//if err == nil {
	//	return
	//}

	event, err := c.EventService.FindOneByID(membership.EventID)
	if err != nil {
		return
	}

	todayStartUTC := helpers.GetStartOfDayInLocUTC(event.Band.GetLocation())
	if event.TimeUTC.After(todayStartUTC) {

		markup := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{{
				Text:         txt.Get("button.moreInfo", membership.User.LanguageCode),
				CallbackData: util.CallbackData(state.EventCB, event.ID.Hex()+":init"),
			}}},
		}

		text := txt.Get("text.memberRemovedNotification", membership.User.LanguageCode,
			user.Name, membership.Role.Name, event.Alias(membership.User.LanguageCode))

		_, err := bot.SendMessage(membership.UserID, text, &gotgbot.SendMessageOpts{
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		})
		if err != nil {
			return
		}
	}
}
