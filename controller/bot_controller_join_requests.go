package controller

import (
	"errors"
	"html"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/service"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (c *BotController) JoinRequestApprove(bot *gotgbot.Bot, ctx *ext.Context) error {
	return c.decideJoinRequest(bot, ctx, true)
}

func (c *BotController) JoinRequestDecline(bot *gotgbot.Bot, ctx *ext.Context) error {
	return c.decideJoinRequest(bot, ctx, false)
}

func (c *BotController) decideJoinRequest(bot *gotgbot.Bot, ctx *ext.Context, approve bool) error {
	requestID, err := bson.ObjectIDFromHex(util.ParseCallbackPayload(ctx.CallbackQuery.Data))
	if err != nil {
		return err
	}

	request, err := c.JoinRequestService.FindOneByID(requestID)
	if err != nil {
		return err
	}

	band, err := c.BandService.FindOneByID(request.BandID)
	if err != nil {
		return err
	}

	admin, err := c.UserService.FindOneByID(ctx.EffectiveUser.Id)
	if err != nil {
		return err
	}
	adminLang := admin.LanguageCode
	if !c.BandService.IsUserAdmin(admin, band) {
		_, _ = ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text:      txt.Get("text.joinRequestInsufficientRights", adminLang),
			ShowAlert: true,
		})
		return nil
	}

	var requestUser *entity.User
	if approve {
		request, requestUser, err = c.JoinRequestService.Approve(requestID, admin.ID)
	} else {
		request, err = c.JoinRequestService.Decline(requestID, admin.ID)
		if err == nil {
			requestUser, err = c.UserService.FindOneByID(request.UserID)
		}
	}
	if errors.Is(err, service.ErrInvalidOperation) {
		_, _ = ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: txt.Get("text.joinRequestAlreadyProcessed", adminLang),
		})
		return nil
	}
	if err != nil {
		return err
	}

	userLang := ""
	if requestUser != nil {
		userLang = requestUser.LanguageCode
	}

	statusText := txt.Get("text.joinRequestApprovedStatus", adminLang)
	userNotification := txt.Get("text.joinRequestApprovedNotification", userLang, request.BandName)
	if !approve {
		statusText = txt.Get("text.joinRequestDeclinedStatus", adminLang)
		userNotification = txt.Get("text.joinRequestDeclinedNotification", userLang, request.BandName)
	}

	_, _, editErr := ctx.EffectiveMessage.EditText(bot, txt.Get(
		"text.joinRequestDecisionSummary",
		adminLang,
		statusText,
		html.EscapeString(request.UserName),
		html.EscapeString(request.BandName),
	), &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
	})
	if editErr != nil {
		log.Warn().Err(editErr).Str("joinRequestID", request.ID.Hex()).Msg("failed to edit join request admin message")
	}

	if _, err := bot.SendMessage(request.UserID, userNotification, nil); err != nil {
		log.Warn().Err(err).Int64("userID", request.UserID).Str("joinRequestID", request.ID.Hex()).Msg("failed to notify user about join request decision")
	}

	_, _ = ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
		Text: statusText,
	})

	return nil
}
