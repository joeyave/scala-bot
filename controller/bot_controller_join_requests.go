package controller

import (
	"errors"
	"fmt"
	"html"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/joeyave/scala-bot/service"
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
	if !c.BandService.IsUserAdmin(admin, band) {
		_, _ = ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "Недостаточно прав.",
			ShowAlert: true,
		})
		return nil
	}

	if approve {
		request, _, err = c.JoinRequestService.Approve(requestID, admin.ID)
	} else {
		request, err = c.JoinRequestService.Decline(requestID, admin.ID)
	}
	if errors.Is(err, service.ErrInvalidOperation) {
		_, _ = ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Запрос уже обработан.",
		})
		return nil
	}
	if err != nil {
		return err
	}

	statusText := "✅ Запрос одобрен"
	userNotification := fmt.Sprintf("Ваш запрос на вступление в группу %s одобрен!", request.BandName)
	if !approve {
		statusText = "❌ Запрос отклонен"
		userNotification = fmt.Sprintf("Ваш запрос на вступление в группу %s отклонен.", request.BandName)
	}

	_, _, editErr := ctx.EffectiveMessage.EditText(bot, fmt.Sprintf(
		"%s\n\n👤 <b>%s</b>\nГруппа: <b>%s</b>",
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
