package controller

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
)

func (c *BotController) BandCreate_AskForName(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	markup := &gotgbot.ReplyKeyboardMarkup{
		Keyboard:       [][]gotgbot.KeyboardButton{{{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode)}}},
		ResizeKeyboard: true,
	}

	_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.sendBandName", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	user.State = entity.State{
		Name:  state.BandCreate,
		Index: 0,
	}

	_, err = c.UserService.UpdateOne(*user)
	if err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

func (c *BotController) BandCreate(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.BandCreate {
			user.State = entity.State{
				Index: index,
				Name:  state.BandCreate,
			}
			user.Cache = entity.Cache{}
		}

		switch index {
		case 0:
			{
				user.Cache.Band = &entity.Band{
					Name: ctx.EffectiveMessage.Text,
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					Keyboard:       [][]gotgbot.KeyboardButton{{{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode)}}},
					ResizeKeyboard: true,
				}

				_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.sendEmail", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
					ReplyMarkup: markup,
				})
				if err != nil {
					return err
				}

				user.State.Index = 1
				return nil
			}
		case 1:
			{
				re := regexp.MustCompile(`(/folders/|id=)(.*?)(/|\?|$)`)
				matches := re.FindStringSubmatch(ctx.EffectiveMessage.Text)
				if len(matches) < 3 {
					return c.BandCreate(0)(bot, ctx)
				}

				user.Cache.Band.DriveFolderID = matches[2]
				user.Role = entity.AdminRole // todo
				band, err := c.BandService.UpdateOne(*user.Cache.Band)
				if err != nil {
					return err
				}

				user.BandID = band.ID
				user.BandIDs = append(user.BandIDs, band.ID)

				text := txt.Get("text.addedToBandAsAdmin", ctx.EffectiveUser.LanguageCode, band.Name)
				_, err = ctx.EffectiveChat.SendMessage(bot, text, nil)
				if err != nil {
					return err
				}

				return c.Menu(bot, ctx)
			}
		}
		return c.Menu(bot, ctx)
	}
}

func (c *BotController) RoleCreate_AskForName(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	markup := &gotgbot.ReplyKeyboardMarkup{
		Keyboard:       [][]gotgbot.KeyboardButton{{{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode)}}},
		ResizeKeyboard: true,
	}

	_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.sendRoleName", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	user.State = entity.State{
		Name:  state.RoleCreate_ChoosePosition,
		Index: 0,
	}

	_, err = c.UserService.UpdateOne(*user)
	if err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

func (c *BotController) RoleCreate_ChoosePosition(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	markup := &gotgbot.InlineKeyboardMarkup{}

	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: "В начало", CallbackData: util.CallbackData(state.RoleCreate, fmt.Sprintf("%s:%d", ctx.EffectiveMessage.Text, 0))}})
	for _, role := range user.Band.Roles {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: role.Name, CallbackData: util.CallbackData(state.RoleCreate, fmt.Sprintf("%s:%d", ctx.EffectiveMessage.Text, role.Priority+1))}})
	}
	//markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode), CallbackData: "todo"}})

	_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.roleIndex", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
	if err != nil {
		return err
	}

	user.State = entity.State{}

	return nil
}

func (c *BotController) RoleCreate(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	priority, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}

	for _, role := range user.Band.Roles {
		if role.Priority >= priority {
			role.Priority++
			_, err := c.RoleService.UpdateOne(*role)
			if err != nil {
				return err
			}
		}
	}

	role, err := c.RoleService.UpdateOne(
		entity.Role{
			Name:     split[0],
			BandID:   user.BandID,
			Priority: priority,
		})
	if err != nil {
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditText(bot, txt.Get("text.roleAdded", ctx.EffectiveUser.LanguageCode, role.Name), nil)
	if err != nil {
		return err
	}

	return c.Menu(bot, ctx)
}
