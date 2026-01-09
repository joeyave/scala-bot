package handlers

// func uploadVoiceHandler() (int, []HandlerFunc) {
//	handlerFunc := make([]HandlerFunc, 0)
//
//	markup := &gotgbot.ReplyKeyboardMarkup{
//		Keyboard:       [][]gotgbot.KeyboardButton{{{Text: helpers.Cancel}}},
//		ResizeKeyboard: true,
//	}
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c *ext.Context, user *entity.User) error {
//
//		_, err := c.EffectiveChat.SendMessage(h.bot, "Введи название песни, к которой ты хочешь прикрепить эту партию:", &gotgbot.SendMessageOpts{ReplyMarkup: markup})
//		if err != nil {
//			return err
//		}
//
//		user.State.Index++
//		return nil
//	})
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c *ext.Context, user *entity.User) error {
//
//		c.EffectiveChat.SendAction(h.bot, "typing")
//
//		driveFiles, _, err := h.driveFileService.FindSomeByFullTextAndFolderID(c.EffectiveMessage.Text, user.Band.DriveFolderID, "")
//		if err != nil {
//			return err
//		}
//
//		if len(driveFiles) == 0 {
//			_, err := c.EffectiveChat.SendMessage(h.bot, "Ничего не найдено. Попробуй другое название.", &gotgbot.SendMessageOpts{ReplyMarkup: markup})
//			return err
//		}
//
//		markup := &gotgbot.ReplyKeyboardMarkup{
//			ResizeKeyboard: true,
//		}
//
//		// TODO: some sort of pagination.
//		for _, driveFile := range driveFiles {
//			markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: driveFile.Name}})
//		}
//		markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: helpers.Cancel}})
//
//		_, err = c.EffectiveChat.SendMessage(h.bot, txt.Get("text.chooseSong", c.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
//		if err != nil {
//			return err
//		}
//
//		user.State.Index++
//		return nil
//	})
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c *ext.Context, user *entity.User) error {
//
//		c.EffectiveChat.SendAction(h.bot, "upload_document")
//
//		foundDriveFile, err := h.driveFileService.FindOneByNameAndFolderID(c.EffectiveMessage.Text, user.Band.DriveFolderID)
//		if err != nil {
//			user.State.Index--
//			return h.Enter(c, user)
//		}
//
//		song, _, err := h.songService.FindOrCreateOneByDriveFileID(foundDriveFile.Id)
//		if err != nil {
//			return err
//		}
//
//		user.State.Context.DriveFileID = song.DriveFileID
//
//		markup := markup
//		_, err = c.EffectiveChat.SendMessage(h.bot, "Отправь мне название этой партии:", &gotgbot.SendMessageOpts{ReplyMarkup: markup})
//		if err != nil {
//			return err
//		}
//
//		user.State.Index++
//		return nil
//	})
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c *ext.Context, user *entity.User) error {
//
//		user.State.Context.Voice.Name = c.EffectiveMessage.Text
//
//		song, err := h.songService.FindOneByDriveFileID(user.State.Context.DriveFileID)
//		if err != nil {
//			return err
//		}
//
//		user.State.Context.Voice.SongID = song.ID
//
//		_, err = h.voiceService.UpdateOne(*user.State.Context.Voice)
//		if err != nil {
//			return err
//		}
//
//		c.EffectiveChat.SendMessage(h.bot, "Добавление завершено.", nil)
//
//		user.State = entity.State{
//			Name: helpers.SongActionsState,
//			Context: entity.Context{
//				DriveFileID: user.State.Context.DriveFileID,
//			},
//		}
//		return h.Enter(c, user)
//	})
//
//	// Upload voice from song menu.
//	handlerFunc = append(handlerFunc, func(h *Handler, c *ext.Context, user *entity.User) error {
//
//		user.State = entity.State{
//			Name:    helpers.UploadVoiceState,
//			Index:   4,
//			Context: entity.Context{DriveFileID: user.State.CallbackData.Query().Get("driveFileId")},
//		}
//
//		_, err := c.EffectiveChat.SendMessage(h.bot, "Отправь мне аудио или голосовое сообщение:", &gotgbot.SendMessageOpts{ReplyMarkup: markup})
//		if err != nil {
//			return err
//		}
//
//		user.State.Index++
//		return nil
//	})
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c *ext.Context, user *entity.User) error {
//
//		c.EffectiveChat.SendAction(h.bot, "typing")
//
//		fileID := c.EffectiveMessage.Voice.FileId
//		if fileID == "" {
//			fileID = c.EffectiveMessage.Audio.FileId
//		}
//		user.State.Context.Voice = &entity.Voice{FileID: fileID}
//
//		_, err := c.EffectiveChat.SendMessage(h.bot, "Отправь мне название этой партии:", &gotgbot.SendMessageOpts{ReplyMarkup: markup})
//		if err != nil {
//			return err
//		}
//
//		user.State.Index++
//		return nil
//	})
//
//	handlerFunc = append(handlerFunc, func(h *Handler, c *ext.Context, user *entity.User) error {
//
//		user.State.Context.Voice.Name = c.EffectiveMessage.Text
//
//		song, err := h.songService.FindOneByDriveFileID(user.State.Context.DriveFileID)
//		if err != nil {
//			return err
//		}
//
//		user.State.Context.Voice.SongID = song.ID
//
//		_, err = h.voiceService.UpdateOne(*user.State.Context.Voice)
//		if err != nil {
//			return err
//		}
//
//		c.EffectiveChat.SendMessage(h.bot, "Добавление завершено.", nil)
//
//		user.State = entity.State{
//			Name: helpers.SongActionsState,
//			Context: entity.Context{
//				DriveFileID: user.State.Context.DriveFileID,
//			},
//			Next: &entity.State{Name: helpers.MainMenuState},
//		}
//		return h.Enter(c, user)
//	})
//
//	return helpers.UploadVoiceState, handlerFunc
//}
