package controller

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
)

func (c *BotController) SettingsChooseBand(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	hex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	bandID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	band, err := c.BandService.FindOneByID(bandID)

	user.BandID = bandID

	_, _, err = bot.EditMessageText(txt.Get("text.addedToBand", ctx.EffectiveUser.LanguageCode, band.Name), &gotgbot.EditMessageTextOpts{
		ChatId:    ctx.EffectiveChat.Id,
		MessageId: ctx.EffectiveMessage.MessageId,
		//ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
	})
	if err != nil {
		return err
	}

	return c.Menu(bot, ctx)
}

func (c *BotController) Settings(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.Settings(user, ctx.EffectiveUser.LanguageCode),
	}

	text := txt.Get("button.settings", ctx.EffectiveUser.LanguageCode) + ":"
	_, err := ctx.EffectiveChat.SendMessage(bot, text, &gotgbot.SendMessageOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *BotController) SettingsCB(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.Settings(user, ctx.EffectiveUser.LanguageCode),
	}

	text := txt.Get("button.settings", ctx.EffectiveUser.LanguageCode) + ":"
	_, _, err := ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

func (c *BotController) SettingsBands(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	markup := gotgbot.InlineKeyboardMarkup{}

	bands, err := c.BandService.FindAll()
	if err != nil {
		return err
	}
	for _, band := range bands {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: band.Name, CallbackData: util.CallbackData(state.SettingsChooseBand, band.ID.Hex())}})
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.createBand", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.BandCreate_AskForName, "")}})
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SettingsCB, "")}})

	text := txt.Get("text.chooseBand", ctx.EffectiveUser.LanguageCode)
	_, _, err = ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

func (c *BotController) SettingsBandMembers(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	hex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	bandID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	return c.settingsBandMembers(bot, ctx, bandID)
}

func (c *BotController) settingsBandMembers(bot *gotgbot.Bot, ctx *ext.Context, bandID primitive.ObjectID) error {

	//user := ctx.Data["user"].(*entity.User)

	members, err := c.UserService.FindMultipleByBandID(bandID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	for _, member := range members {
		text := member.Name
		if member.Role == entity.AdminRole {
			text += " ✔️"
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.SettingsBandAddAdmin, fmt.Sprintf("%s:%d:delete", bandID.Hex(), member.ID))}})
		} else {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.SettingsBandAddAdmin, fmt.Sprintf("%s:%d:add", bandID.Hex(), member.ID))}})
		}
	}

	markup.InlineKeyboard = util.SplitKeyboardToColumns(markup.InlineKeyboard, 2)

	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SettingsCB, "")}})

	text := txt.Get("text.chooseMemberToMakeAdmin", ctx.EffectiveUser.LanguageCode)
	_, _, err = ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

// todo: BUG! move role from User to Band

func (c *BotController) SettingsBandAddAdmin(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	bandID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil {
		return err
	}

	user, err := c.UserService.FindOneByID(userID)
	if err != nil {
		return err
	}

	switch split[2] {
	case "delete":
		user.Role = ""
	case "add":
		user.Role = entity.AdminRole
	}
	_, err = c.UserService.UpdateOne(*user)
	if err != nil {
		return err
	}

	return c.settingsBandMembers(bot, ctx, bandID)
}
