package controller

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/exp/slices"
	"google.golang.org/api/googleapi"
)

func (c *BotController) SettingsChooseBand(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*entity.User)

	hex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	bandID, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	// band, err := c.BandService.FindOneByID(bandID)

	index := slices.IndexFunc(user.BandIDs, func(id bson.ObjectID) bool {
		return id == bandID
	})

	if index == -1 { // If adding.
		user.BandID = bandID
		user.BandIDs = append(user.BandIDs, bandID)

		// _, _, err = bot.EditMessageText(txt.Get("text.addedToBand", ctx.EffectiveUser.LanguageCode, band.Name), &gotgbot.EditMessageTextOpts{
		//	ChatId:    ctx.EffectiveChat.Id,
		//	MessageId: ctx.EffectiveMessage.MessageId,
		// })
		//if err != nil {
		//	return err
		//}
	} else { // If removing.
		user.BandIDs = slices.Delete(user.BandIDs, index, index+1)

		if len(user.BandIDs) > 0 {
			user.BandID = user.BandIDs[0]
		} else {
			user.BandID = bson.NilObjectID
		}

		// _, _, err = bot.EditMessageText(txt.Get("text.removedFromBand", ctx.EffectiveUser.LanguageCode, band.Name), &gotgbot.EditMessageTextOpts{
		//	ChatId:    ctx.EffectiveChat.Id,
		//	MessageId: ctx.EffectiveMessage.MessageId,
		// })
		//if err != nil {
		//	return err
		//}
	}

	return c.SettingsBands(bot, ctx)
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

	_, _ = ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

func (c *BotController) SettingsBands(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*entity.User)

	markup := gotgbot.InlineKeyboardMarkup{}

	bands, err := c.BandService.FindAll()
	if err != nil {
		return err
	}
	for _, band := range bands {
		text := band.Name
		contains := slices.ContainsFunc(user.BandIDs, func(id bson.ObjectID) bool {
			return id == band.ID
		})
		if contains || user.BandID == band.ID {
			text = "✔️ " + text
		}

		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.SettingsChooseBand, band.ID.Hex())}})
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

	_, _ = ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

func (c *BotController) SettingsBandMembers(bot *gotgbot.Bot, ctx *ext.Context) error {
	// user := ctx.Data["user"].(*entity.User)

	hex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	bandID, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	return c.settingsBandMembers(bot, ctx, bandID)
}

func (c *BotController) SettingsCleanupDatabase(bot *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*entity.User)

	// hex := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	// bandID, err := bson.ObjectIDFromHex(hex)
	//if err != nil {
	//	return err
	//}

	songs, err := c.SongService.FindAll()
	if err != nil {
		return err
	}

	builder := strings.Builder{}
	builder.WriteString(txt.Get("text.cleanupDatabase", ctx.EffectiveUser.LanguageCode))

	msg, err := ctx.EffectiveChat.SendMessage(bot, builder.String(), nil)
	if err != nil {
		return err
	}

	for _, song := range songs {
		driveFile, err := c.DriveFileService.FindOneByID(song.DriveFileID)

		var gErr *googleapi.Error
		if errors.As(err, &gErr) && gErr.Code == 404 {
			deleted, _ := c.SongService.DeleteOneByDriveFileIDFromDatabase(song.DriveFileID)
			if deleted {
				builder.WriteString(fmt.Sprintf("\nDeleted: <a href=\"%s\">%s</a>", song.PDF.WebViewLink, song.PDF.Name))
				var voiceIDs []bson.ObjectID
				for _, voice := range song.Voices {
					voiceIDs = append(voiceIDs, voice.ID)
				}
				if len(voiceIDs) > 0 {
					err := c.VoiceService.DeleteManyByIDs(voiceIDs)
					if err != nil {
						builder.WriteString(fmt.Sprintf("\nError deleting voices for song %s:%s", song.ID.String(), err.Error()))
					}
				}
			} else {
				builder.WriteString(fmt.Sprintf("\nCould not delete: <a href=\"%s\">%s</a> | %s", song.PDF.WebViewLink, song.PDF.Name, song.ID.String()))
			}
		} else if slices.Contains(driveFile.Parents, user.Band.ArchiveFolderID) && !song.IsArchived {
			_, err := c.SongService.Archive(song.ID)
			if err != nil {
				builder.WriteString(fmt.Sprintf("\nCouldn't archive: <a href=\"%s\">%s</a>", song.PDF.WebViewLink, song.PDF.Name))
			}
			builder.WriteString(fmt.Sprintf("\nArchived: <a href=\"%s\">%s</a>", song.PDF.WebViewLink, song.PDF.Name))
		}

		_, _, _ = msg.EditText(bot, builder.String(), &gotgbot.EditMessageTextOpts{
			ParseMode: "HTML",
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})
	}

	builder.WriteString("\n\nDone!")

	_, _, _ = msg.EditText(bot, builder.String(), &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return nil
}

func (c *BotController) settingsBandMembers(bot *gotgbot.Bot, ctx *ext.Context, bandID bson.ObjectID) error {
	// user := ctx.Data["user"].(*entity.User)

	members, err := c.UserService.FindMultipleByBandID(bandID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	for _, member := range members {
		text := member.Name
		if member.Role == entity.AdminRole {
			text = "✔️ " + text
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.SettingsBandAddAdmin, fmt.Sprintf("%s:%d:delete", bandID.Hex(), member.ID))}})
		} else {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: text, CallbackData: util.CallbackData(state.SettingsBandAddAdmin, fmt.Sprintf("%s:%d:add", bandID.Hex(), member.ID))}})
		}
	}

	markup.InlineKeyboard = util.SplitInlineKeyboardToColumns(markup.InlineKeyboard, 2)

	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SettingsCB, "")}})

	text := txt.Get("text.chooseMemberToMakeAdmin", ctx.EffectiveUser.LanguageCode)
	_, _, err = ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(bot, nil)

	return nil
}

// todo: BUG! move role from User to Band

func (c *BotController) SettingsBandAddAdmin(bot *gotgbot.Bot, ctx *ext.Context) error {
	// user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	bandID, err := bson.ObjectIDFromHex(split[0])
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
