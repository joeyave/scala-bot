package handlers

// func SendSongToChannel(h *Handler, c *ext.Context, user *entities.User, song *entities.Song) error {
//	send := func() (*telebot.Message, error) {
//		return h.bot.Send(
//			telebot.ChatID(helpers.FilesChannelID),
//			&telebot.Document{
//				File: telebot.File{FileID: song.PDF.TgFileID},
//			},
//			telebot.Silent)
//	}
//
//	edit := func() (*telebot.Message, error) {
//		return h.bot.EditMedia(
//			&telebot.Message{
//				ID:   song.PDF.TgChannelMessageID,
//				Chat: &telebot.Chat{ID: helpers.FilesChannelID},
//			}, &telebot.Document{
//				File: telebot.File{FileID: song.PDF.TgFileID},
//				MIME: "application/pdf",
//			},
//		)
//	}
//
//	var msg *telebot.Message
//	var err error
//	if song.PDF.TgChannelMessageID == 0 {
//		msg, err = send()
//		if err != nil {
//			return err
//		}
//		song.PDF.TgChannelMessageID = msg.ID
//	} else {
//		msg, err = edit()
//		if err != nil {
//			if fmt.Sprint(err) == "telegram unknown: Bad Request: MESSAGE_ID_INVALID (400)" {
//				msg, err = send()
//				if err != nil {
//					return err
//				}
//				song.PDF.TgChannelMessageID = msg.ID
//			}
//		}
//	}
//
//	return nil
//}
