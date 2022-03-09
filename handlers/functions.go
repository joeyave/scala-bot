package handlers

import (
	"fmt"
	"github.com/joeyave/scala-bot/entities"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/klauspost/lctime"
	"gopkg.in/telebot.v3"
	"strings"
	"sync"
	"time"
)

func SendDriveFileToUser(h *Handler, c telebot.Context, user *entities.User, driveFileID string) error {

	q := user.State.CallbackData.Query()
	q.Set("driveFileId", driveFileID)
	user.State.CallbackData.RawQuery = q.Encode()

	song, driveFile, err := h.songService.FindOrCreateOneByDriveFileID(driveFileID)
	if err != nil {
		return err
	}

	markup := &telebot.ReplyMarkup{}

	markup.InlineKeyboard = helpers.GetSongInitKeyboard(user, song)

	sendDocumentByReader := func() (*telebot.Message, error) {
		reader, err := h.driveFileService.DownloadOneByID(driveFile.Id)
		if err != nil {
			return nil, err
		}

		if c.Callback() != nil {
			return h.bot.EditMedia(
				&telebot.Message{
					ID:   c.Callback().Message.ID,
					Chat: c.Callback().Message.Chat,
				}, &telebot.Document{
					File:     telebot.FromReader(*reader),
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
					Caption:  helpers.AddCallbackData(song.Caption()+"\n"+strings.Join(song.Tags, ", "), user.State.CallbackData.String()),
				}, markup, telebot.ModeHTML)
		} else {
			return h.bot.Send(
				c.Recipient(),
				&telebot.Document{
					File:     telebot.FromReader(*reader),
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
					Caption:  helpers.AddCallbackData(song.Caption()+"\n"+strings.Join(song.Tags, ", "), user.State.CallbackData.String()),
				}, markup, telebot.ModeHTML)
		}
	}

	sendDocumentByFileID := func() (*telebot.Message, error) {
		if c.Callback() != nil {
			return h.bot.EditMedia(
				&telebot.Message{
					ID:   c.Callback().Message.ID,
					Chat: c.Callback().Message.Chat,
				},
				&telebot.Document{
					File:     telebot.File{FileID: song.PDF.TgFileID},
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
					Caption:  helpers.AddCallbackData(song.Caption()+"\n"+strings.Join(song.Tags, ", "), user.State.CallbackData.String()),
				}, markup, telebot.ModeHTML)
		} else {
			return h.bot.Send(
				c.Recipient(),
				&telebot.Document{
					File:     telebot.File{FileID: song.PDF.TgFileID},
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
					Caption:  helpers.AddCallbackData(song.Caption()+"\n"+strings.Join(song.Tags, ", "), user.State.CallbackData.String()),
				}, markup, telebot.ModeHTML)
		}
	}

	var msg *telebot.Message
	if song.PDF.TgFileID == "" {
		msg, err = sendDocumentByReader()
	} else {
		msg, err = sendDocumentByFileID()
		if err != nil {
			msg, err = sendDocumentByReader()
		}
	}
	if err != nil {
		return err
	}

	song.PDF.TgFileID = msg.Document.FileID
	err = SendSongToChannel(h, c, user, song)
	if err != nil {
		return err
	}

	song, err = h.songService.UpdateOne(*song)

	return err
}

func SendSongToChannel(h *Handler, c telebot.Context, user *entities.User, song *entities.Song) error {
	send := func() (*telebot.Message, error) {
		return h.bot.Send(
			telebot.ChatID(helpers.FilesChannelID),
			&telebot.Document{
				File: telebot.File{FileID: song.PDF.TgFileID},
			},
			telebot.Silent)
	}

	edit := func() (*telebot.Message, error) {
		return h.bot.EditMedia(
			&telebot.Message{
				ID:   song.PDF.TgChannelMessageID,
				Chat: &telebot.Chat{ID: helpers.FilesChannelID},
			}, &telebot.Document{
				File: telebot.File{FileID: song.PDF.TgFileID},
				MIME: "application/pdf",
			},
		)
	}

	var msg *telebot.Message
	var err error
	if song.PDF.TgChannelMessageID == 0 {
		msg, err = send()
		if err != nil {
			return err
		}
		song.PDF.TgChannelMessageID = msg.ID
	} else {
		msg, err = edit()
		if err != nil {
			if fmt.Sprint(err) == "telegram unknown: Bad Request: MESSAGE_ID_INVALID (400)" {
				msg, err = send()
				if err != nil {
					return err
				}
				song.PDF.TgChannelMessageID = msg.ID
			}
		}
	}

	return nil
}

func sendDriveFilesAlbum(h *Handler, c telebot.Context, user *entities.User, driveFileIDs []string) error {

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(driveFileIDs))
	bigAlbum := make(telebot.Album, len(driveFileIDs))

	for i := range driveFileIDs {
		go func(i int) {
			defer waitGroup.Done()

			song, driveFile, err := h.songService.FindOrCreateOneByDriveFileID(driveFileIDs[i])
			if err != nil {
				return
			}

			if song.PDF.TgFileID == "" {
				reader, err := h.driveFileService.DownloadOneByID(driveFile.Id)
				if err != nil {
					return
				}

				bigAlbum[i] = &telebot.Document{
					File:     telebot.FromReader(*reader),
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", song.PDF.Name),
					Caption:  song.Caption(),
				}
			} else {
				bigAlbum[i] = &telebot.Document{
					File:    telebot.File{FileID: song.PDF.TgFileID},
					Caption: song.Caption(),
				}
			}
		}(i)
	}
	waitGroup.Wait()

	const chunkSize = 10
	chunks := chunkAlbumBy(bigAlbum, chunkSize)

	for i, album := range chunks {
		_, err := h.bot.SendAlbum(c.Recipient(), album)

		// TODO: check for bugs.
		if err != nil {
			fromIndex := 0
			toIndex := 0 + len(album)

			if i-1 > 0 && i-1 < len(chunks) {
				fromIndex = i * len(chunks[i-1])
				toIndex = fromIndex + len(chunks[i])
			}

			foundDriveFileIDs := driveFileIDs[fromIndex:toIndex]

			var waitGroup sync.WaitGroup
			waitGroup.Add(len(foundDriveFileIDs))
			bigAlbum := make(telebot.Album, len(foundDriveFileIDs))

			for i := range foundDriveFileIDs {
				go func(i int) {
					defer waitGroup.Done()
					reader, err := h.driveFileService.DownloadOneByID(foundDriveFileIDs[i])
					if err != nil {
						return
					}

					// driveFile, err := h.driveFileService.FindOneByID(foundDriveFileIDs[i])
					// if err != nil {
					// 	return
					// }

					song, driveFile, err := h.songService.FindOrCreateOneByDriveFileID(driveFileIDs[i])
					if err != nil {
						return
					}

					bigAlbum[i] = &telebot.Document{
						File:     telebot.FromReader(*reader),
						MIME:     "application/pdf",
						FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
						Caption:  song.Caption(),
					}
				}(i)
			}
			waitGroup.Wait()

			_, err = h.bot.SendAlbum(c.Recipient(), bigAlbum)
			if err != nil {
				continue
			}
		}

		// for j := range responses {
		//	foundDriveFileID := driveFileIDs[j+(i*len(album))]
		//
		//	song, err := h.songService.FindOneByDriveFileID(foundDriveFileID)
		//	if err != nil {
		//		continue
		//	}
		//
		//	song.PDF.TgFileID = responses[j].Document.FileID
		//	err = SendSongToChannel(h, c, user, song)
		//	if err != nil {
		//		continue
		//	}
		//
		//	_, _ = h.songService.UpdateOne(*song)
		// }
	}

	return nil
}

func GetCalendarMarkup(now, monthFirstDayDate, monthLastDayDate time.Time) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}

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

	return markup
}
