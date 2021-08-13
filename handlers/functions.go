package handlers

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/telebot/v3"
	"github.com/klauspost/lctime"
	"github.com/rs/zerolog/log"
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

	//markup.InlineKeyboard = helpers.GetSongActionsKeyboard(*user, *song, *driveFile)
	markup.InlineKeyboard = [][]telebot.InlineButton{
		{
			{Text: "Кнопочки", Data: helpers.AggregateCallbackData(helpers.SongActionsState, 1, "")},
		},
	}

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
					Caption:  helpers.AddCallbackData(song.Caption(), user.State.CallbackData.String()),
				}, markup, telebot.ModeHTML)
		} else {
			return h.bot.Send(
				c.Recipient(),
				&telebot.Document{
					File:     telebot.FromReader(*reader),
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
					Caption:  helpers.AddCallbackData(song.Caption(), user.State.CallbackData.String()),
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
					Caption:  helpers.AddCallbackData(song.Caption(), user.State.CallbackData.String()),
				}, markup, telebot.ModeHTML)
		} else {
			return h.bot.Send(
				c.Recipient(),
				&telebot.Document{
					File:     telebot.File{FileID: song.PDF.TgFileID},
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
					Caption:  helpers.AddCallbackData(song.Caption(), user.State.CallbackData.String()),
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
	documents := make([]telebot.InputMedia, len(driveFileIDs))

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

				documents[i] = &telebot.Document{
					File:     telebot.FromReader(*reader),
					MIME:     "application/pdf",
					FileName: fmt.Sprintf("%s.pdf", song.PDF.Name),
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

			foundDriveFileIDs := driveFileIDs[fromIndex:toIndex]

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

		for _, response := range responses {
			log.Debug().Msgf("%s", response.Document.FileName)
		}
		//for j := range responses {
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
		//}
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
