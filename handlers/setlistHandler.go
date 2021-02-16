package handlers

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"sort"
	"sync"
)

func setlistHandler() (string, []func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error)) {
	handleFuncs := make([]func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error), 0)

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		if len(user.State.Context.Setlist) < 1 {
			user.State.Index = 2
			return updateHandler.enterStateHandler(update, user)
		}

		songNames := user.State.Context.Setlist

		currentSongName := songNames[0]
		user.State.Context.Setlist = songNames[1:]

		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		_, _ = updateHandler.bot.Send(chatAction)

		songs, _, err := updateHandler.songService.FindByName(currentSongName, "")
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("По запросу \"%s\" ничего не найдено. Напиши новое название или пропусти эту песню.", currentSongName))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Skip)),
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)),
			)
			res, err := updateHandler.bot.Send(msg)
			if err != nil {
				return user, err
			}

			user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, res.MessageID)
			user.State.Index++
			return user, err
		}

		keyboard := tgbotapi.NewReplyKeyboard()
		keyboard.OneTimeKeyboard = false
		keyboard.ResizeKeyboard = true

		// TODO: some sort of pagination.
		for _, song := range songs {
			songButton := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(song.Name))
			keyboard.Keyboard = append(keyboard.Keyboard, songButton)
		}

		keyboard.Keyboard = append(keyboard.Keyboard,
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Skip)),
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Cancel)),
		)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Выбери песню по запросу \"%s\" или введи другое название:", currentSongName))
		msg.ReplyMarkup = keyboard
		res, err := updateHandler.bot.Send(msg)
		if err != nil {
			return user, err
		}

		user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, res.MessageID)
		user.State.Context.Songs = songs
		user.State.Index++

		return user, nil
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		user.State.Context.MessagesToDelete = append(user.State.Context.MessagesToDelete, update.Message.MessageID)

		switch update.Message.Text {
		case "":
		case helpers.Skip:
			user.State.Index = 0
			return updateHandler.enterStateHandler(update, user)
		}

		songs := user.State.Context.Songs
		sort.Slice(songs, func(i, j int) bool {
			return songs[i].Name <= songs[j].Name
		})

		foundIndex := sort.Search(len(songs), func(i int) bool {
			return songs[i].Name >= update.Message.Text
		})

		if foundIndex != len(songs) {
			song, err := updateHandler.songService.GetWithActualTgFileID(songs[foundIndex])
			if err != nil {
				user.State.Context.FoundSongs = append(user.State.Context.FoundSongs, songs[foundIndex])
			} else {
				user.State.Context.FoundSongs = append(user.State.Context.FoundSongs, song)
			}
		} else {
			user.State.Context.Setlist = append([]string{update.Message.Text}, user.State.Context.Setlist...)
		}

		user.State.Index = 0
		return updateHandler.enterStateHandler(update, user)
	})

	handleFuncs = append(handleFuncs, func(updateHandler *UpdateHandler, update *tgbotapi.Update, user entities.User) (entities.User, error) {
		chatAction := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatUploadDocument)
		_, _ = updateHandler.bot.Send(chatAction)

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(user.State.Context.FoundSongs))
		var documents []interface{}
		for _, song := range user.State.Context.FoundSongs {
			go func(song entities.Song) {
				defer waitGroup.Done()
				if song.TgFileID != "" {
					documents = append(documents, tgbotapi.NewInputMediaDocument(tgbotapi.FileID(song.TgFileID)))
				} else {
					fileReader, _ := updateHandler.songService.DownloadPDF(song)
					documents = append(documents, tgbotapi.NewInputMediaDocument(fileReader))
				}
			}(song)
		}
		waitGroup.Wait()

		const chunkSize = 10
		chunks := chunkBy(documents, chunkSize)

		for i, chunk := range chunks {
			responses, err := updateHandler.bot.SendMediaGroup(tgbotapi.NewMediaGroup(update.Message.Chat.ID, chunk))
			if err != nil {
				fromIndex := 0
				toIndex := 0 + len(chunk)

				if i-1 < len(chunks) {
					fromIndex = i * len(chunks[i-1])
					toIndex = fromIndex + len(chunks[i])
				}

				songs := user.State.Context.FoundSongs[fromIndex:toIndex]

				var newChunk []interface{}

				var waitGroup sync.WaitGroup
				waitGroup.Add(len(songs))
				for _, song := range songs {
					go func(song entities.Song) {
						defer waitGroup.Done()
						fileReader, _ := updateHandler.songService.DownloadPDF(song)
						newChunk = append(newChunk, tgbotapi.NewInputMediaDocument(fileReader))
					}(song)
				}
				waitGroup.Wait()
				responses, err = updateHandler.bot.SendMediaGroup(tgbotapi.NewMediaGroup(update.Message.Chat.ID, newChunk))
				if err != nil {
					continue
				}
			}

			for j, response := range responses {
				user.State.Context.FoundSongs[j+(i*len(chunk))].TgFileID = response.Document.FileID
				_, _ = updateHandler.songService.UpdateOne(user.State.Context.FoundSongs[j+(i*len(chunk))])
			}
		}

		user.State = user.State.Prev
		user.State.Index = 0

		return updateHandler.enterStateHandler(update, user)
	})

	return helpers.SetlistState, handleFuncs
}

func chunkBy(items []interface{}, chunkSize int) (chunks [][]interface{}) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}
