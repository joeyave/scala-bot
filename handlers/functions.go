package handlers

import (
	"errors"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/telebot/v3"
	"sync"
)

func SendSongToUser(h *Handler, c telebot.Context, user *entities.User, song *entities.Song) (*telebot.Message, error) {
	markup := &telebot.ReplyMarkup{
		ReplyKeyboard:  [][]telebot.ReplyButton{{{Text: song.PDF.Name}}},
		ResizeKeyboard: true,
	}

	markup.ReplyKeyboard = append(markup.ReplyKeyboard, helpers.GetSongActionsKeyboard(*user, *song)...)

	sendDocumentByReader := func() (*telebot.Message, error) {
		reader, err := h.driveFileService.DownloadOneByID(song.DriveFileID)
		if err != nil {
			return nil, err
		}

		return h.bot.Send(c.Recipient(), &telebot.Document{
			File:     telebot.FromReader(*reader),
			MIME:     "application/pdf",
			FileName: fmt.Sprintf("%s.pdf", song.PDF.Name),
		}, markup)
	}

	sendDocumentByFileID := func() (*telebot.Message, error) {
		return h.bot.Send(c.Recipient(), &telebot.Document{
			File:     telebot.File{FileID: song.PDF.TgFileID},
			MIME:     "application/pdf",
			FileName: fmt.Sprintf("%s.pdf", song.PDF.Name),
		}, markup)
	}

	var msg *telebot.Message
	var err error
	if song.PDF.TgFileID == "" {
		msg, err = sendDocumentByReader()
	} else {
		msg, err = sendDocumentByFileID()
		if err != nil {
			msg, err = sendDocumentByReader()
		}
	}

	return msg, err
}

func SendSongToChannel(h *Handler, c telebot.Context, user *entities.User, song *entities.Song) (*telebot.Message, error) {
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
			})
	}

	var msg *telebot.Message
	var err error
	if song.PDF.TgChannelMessageID == 0 {
		msg, err = send()
	} else {
		msg, err = edit()
		if err != nil {
			if !errors.Is(err, telebot.ErrSameMessageContent) {
				msg, err = send()
			}
		}
	}

	return msg, err
}

func sendSongsAlbum(h *Handler, c telebot.Context, user *entities.User, driveFileIDs []string) error {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(driveFileIDs))
	documents := make([]telebot.InputMedia, len(driveFileIDs))
	for i := range driveFileIDs {
		go func(i int) {
			defer waitGroup.Done()

			song, err := h.songService.FindOrCreateOneByDriveFileID(driveFileIDs[i])
			if err != nil {
				return
			}

			if song.PDF.TgFileID == "" {
				reader, err := h.driveFileService.DownloadOneByID(song.DriveFileID)
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

		for j := range responses {
			foundDriveFileID := driveFileIDs[j+(i*len(album))]

			song, err := h.songService.FindOrCreateOneByDriveFileID(foundDriveFileID)
			if err != nil {
				return err
			}

			song.PDF.TgFileID = responses[j].Document.FileID
			msg, err := SendSongToChannel(h, c, user, song)
			if err == nil {
				song.PDF.TgChannelMessageID = msg.ID
			}

			_, _ = h.songService.UpdateOne(*song)
		}
	}

	return nil
}
