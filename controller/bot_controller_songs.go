package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"
	"google.golang.org/api/drive/v3"
	"net/url"
	"os"
	"strings"
	"time"
)

type CreateSongData struct {
	Name   string   `json:"name"`
	Key    string   `json:"key"`
	BPM    string   `json:"bpm"`
	Time   string   `json:"time"`
	Lyrics string   `json:"lyrics"`
	Tags   []string `json:"tags"`
}

func (c *BotController) CreateSong(bot *gotgbot.Bot, ctx *ext.Context) error {

	ctx.EffectiveChat.SendAction(bot, "upload_document", nil)

	var data *CreateSongData
	err := json.Unmarshal([]byte(ctx.EffectiveMessage.WebAppData.Data), &data)
	if err != nil {
		return err
	}

	user := ctx.Data["user"].(*entity.User)

	file := &drive.File{
		Name:     data.Name,
		Parents:  []string{user.Band.DriveFolderID},
		MimeType: "application/vnd.google-apps.document",
	}
	driveFile, err := c.DriveFileService.CreateOne(file, data.Lyrics, data.Key, data.BPM, data.Time)
	if err != nil {
		return err
	}

	driveFile, err = c.DriveFileService.StyleOne(driveFile.Id)
	if err != nil {
		return err
	}

	c.SongService.UpdateOne(entity.Song{
		DriveFileID: driveFile.Id,
		BandID:      user.BandID,
		PDF: entity.PDF{
			Name:        data.Name,
			Key:         data.Key,
			BPM:         data.BPM,
			Time:        data.Time,
			WebViewLink: driveFile.WebViewLink,
		},
		Tags: data.Tags,
	})

	err = c.song(bot, ctx, driveFile.Id)
	if err != nil {
		return err
	}

	return c.GetSongs(0)(bot, ctx)
}

func (c *BotController) song(bot *gotgbot.Bot, ctx *ext.Context, driveFileID string) error {

	user := ctx.Data["user"].(*entity.User)

	ctx.EffectiveChat.SendAction(bot, "upload_document", nil)

	song, driveFile, err := c.SongService.FindOrCreateOneByDriveFileID(driveFileID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.SongInit(song, user, 0, 0, ctx.EffectiveUser.LanguageCode),
	}

	sendDocumentByReader := func() (*gotgbot.Message, error) {
		reader, err := c.DriveFileService.DownloadOneByID(driveFile.Id)
		if err != nil {
			return nil, err
		}

		message, err := bot.SendDocument(ctx.EffectiveChat.Id, gotgbot.NamedFile{
			File:     *reader,
			FileName: fmt.Sprintf("%s.pdf", driveFile.Name),
		}, &gotgbot.SendDocumentOpts{
			Caption:     song.Caption(),
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		})
		return message, err
	}

	sendDocumentByFileID := func() (*gotgbot.Message, error) {
		message, err := bot.SendDocument(ctx.EffectiveChat.Id, song.PDF.TgFileID, &gotgbot.SendDocumentOpts{
			Caption:     song.Caption(),
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		})
		return message, err
	}

	var msg *gotgbot.Message
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

	song.PDF.TgFileID = msg.Document.FileId

	// todo
	//err = SendSongToChannel(h, c, user, song)
	//if err != nil {
	//	return err
	//}

	song, err = c.SongService.UpdateOne(*song)
	if err != nil {
		return err
	}

	user.CallbackCache.ChatID = msg.Chat.Id
	user.CallbackCache.MessageID = msg.MessageId
	user.CallbackCache.UserID = user.ID
	caption := user.CallbackCache.AddToText(song.Caption())

	markup = gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.SongInit(song, user, msg.Chat.Id, msg.MessageId, ctx.EffectiveUser.LanguageCode),
	}

	_, _, err = msg.EditCaption(bot, &gotgbot.EditMessageCaptionOpts{
		ParseMode:   "HTML",
		Caption:     caption,
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *BotController) GetSongs(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.GetSongs {
			user.State = entity.State{
				Index: index,
				Name:  state.GetSongs,
			}
			user.Cache = entity.Cache{}
		}

		switch index {
		case 0:
			{
				ctx.EffectiveChat.SendAction(bot, "typing", nil)

				if ctx.EffectiveMessage.Text == txt.Get("button.prev", ctx.EffectiveUser.LanguageCode) && user.Cache.NextPageToken.GetPrevValue() != "" {
					user.Cache.NextPageToken = user.Cache.NextPageToken.Prev.Prev
				}

				driveFiles, nextPageToken, err := c.DriveFileService.FindAllByFolderID(user.Band.DriveFolderID, user.Cache.NextPageToken.GetValue())
				if err != nil {
					return err
				}

				user.Cache.NextPageToken = &entity.NextPageToken{
					Value: nextPageToken,
					Prev:  user.Cache.NextPageToken,
				}

				if len(driveFiles) == 0 {
					markup := &gotgbot.ReplyKeyboardMarkup{
						Keyboard: [][]gotgbot.KeyboardButton{
							{{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode)}},
						},
						ResizeKeyboard: true,
					}
					_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.noDocs", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
					return err
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
				}

				markup.Keyboard = append(markup.Keyboard, keyboard.GetSongsStateFilterButtons(ctx.EffectiveUser.LanguageCode))
				markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.createDoc", ctx.EffectiveUser.LanguageCode), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/web-app/songs/create?userId=%d&lang=%s", os.Getenv("BOT_DOMAIN"), user.ID, ctx.EffectiveUser.LanguageCode)}}})
				//markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.createDoc", ctx.EffectiveUser.LanguageCode), WebApp: &gotgbot.WebAppInfo{Url: fmt.Sprintf("%s/web-app/songs/create?userId=%d", os.Getenv("BOT_DOMAIN"), user.ID)}}})

				likedSongs, likedSongErr := c.SongService.FindManyLiked(user.BandID, user.ID)

				for _, driveFile := range driveFiles {
					opts := &keyboard.DriveFileButtonOpts{
						ShowLike: true,
					}
					if likedSongErr != nil {
						opts.ShowLike = false
					}
					markup.Keyboard = append(markup.Keyboard, keyboard.DriveFileButton(driveFile, likedSongs, opts))
				}

				markup.Keyboard = append(markup.Keyboard, keyboard.NavigationByToken(user.Cache.NextPageToken, ctx.EffectiveUser.LanguageCode)...)

				_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseSong", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
				if err != nil {
					return err
				}

				user.Cache.DriveFiles = driveFiles

				user.State.Index = 1

				return nil
			}
		case 1:
			{
				switch ctx.EffectiveMessage.Text {

				case txt.Get("button.next", ctx.EffectiveUser.LanguageCode), txt.Get("button.prev", ctx.EffectiveUser.LanguageCode):
					return c.GetSongs(0)(bot, ctx)

				case txt.Get("button.like", ctx.EffectiveUser.LanguageCode), txt.Get("button.calendar", ctx.EffectiveUser.LanguageCode), txt.Get("button.numbers", ctx.EffectiveUser.LanguageCode), txt.Get("button.tag", ctx.EffectiveUser.LanguageCode):
					return c.filterSongs(0)(bot, ctx)
				}

				ctx.EffectiveChat.SendAction(bot, "upload_document", nil)

				driveFileName := keyboard.ParseDriveFileButton(ctx.EffectiveMessage.Text)

				driveFiles := user.Cache.DriveFiles
				var foundDriveFile *drive.File
				for _, driveFile := range driveFiles {
					if driveFile.Name == driveFileName {
						foundDriveFile = driveFile
						break
					}
				}

				if foundDriveFile != nil {
					return c.song(bot, ctx, foundDriveFile.Id)
				} else {
					return c.search(0)(bot, ctx)
				}
			}
		}
		return c.Menu(bot, ctx)
	}
}

func (c *BotController) filterSongs(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.FilterSongs {
			user.State = entity.State{
				Index: index,
				Name:  state.FilterSongs,
			}
			user.Cache = entity.Cache{}
		}

		switch index {
		case 0:
			{

				statsPeriodStartDate := entity.GetStatsPeriodStartDate(user.Cache.StatsPeriod, time.Now())
				ctx.EffectiveChat.SendAction(bot, "typing", nil)

				switch ctx.EffectiveMessage.Text {
				case txt.Get("button.like", ctx.EffectiveUser.LanguageCode), txt.Get("button.numbers", ctx.EffectiveUser.LanguageCode), txt.Get("button.calendar", ctx.EffectiveUser.LanguageCode):
					user.Cache.Filter = ctx.EffectiveMessage.Text

				case txt.Get("button.tag", ctx.EffectiveUser.LanguageCode):
					user.Cache.Filter = ctx.EffectiveMessage.Text
					return c.filterSongs(2)(bot, ctx)
				}

				var (
					songs []*entity.SongWithEvents
					err   error
				)

				switch user.Cache.Filter {
				case txt.Get("button.like", ctx.EffectiveUser.LanguageCode):
					songs, err = c.SongService.FindManyExtraLiked(user.BandID, user.ID, statsPeriodStartDate, user.Cache.PageIndex)
				case txt.Get("button.calendar", ctx.EffectiveUser.LanguageCode):
					songs, err = c.SongService.FindAllExtraByPageNumberSortedByLatestEventDate(user.BandID, statsPeriodStartDate, user.Cache.PageIndex)
				case txt.Get("button.numbers", ctx.EffectiveUser.LanguageCode):
					songs, err = c.SongService.FindAllExtraByPageNumberSortedByEventsNumber(user.BandID, statsPeriodStartDate, user.Cache.PageIndex)
				case txt.Get("button.tag", ctx.EffectiveUser.LanguageCode):
					if keyboard.IsSelectedButton(ctx.EffectiveMessage.Text) {
						return c.GetSongs(0)(bot, ctx)
					}
					if user.Cache.Query == "" {
						user.Cache.Query = ctx.EffectiveMessage.Text
					}
					songs, err = c.SongService.FindManyExtraByTag(user.Cache.Query, user.BandID, statsPeriodStartDate, user.Cache.PageIndex)
				}
				if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
					return err
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
				}

				filterButtons := keyboard.GetSongsStateFilterButtons(ctx.EffectiveUser.LanguageCode)
				for i := range filterButtons {
					if filterButtons[i].Text == user.Cache.Filter {
						filterButtons[i] = keyboard.SelectedButton(filterButtons[i].Text)
						break
					}
				}
				markup.Keyboard = append(markup.Keyboard, filterButtons)
				markup.Keyboard = append(markup.Keyboard, keyboard.GetStatsPeriodButton(user.Cache.StatsPeriod, ctx.EffectiveUser.LanguageCode))

				for _, song := range songs {

					songButtonOpts := &keyboard.SongButtonOpts{
						ShowLike:  false,
						ShowStats: true,
					}

					if user.Cache.Filter != txt.Get("button.like", ctx.EffectiveUser.LanguageCode) {
						songButtonOpts.ShowLike = true
					}

					markup.Keyboard = append(markup.Keyboard, keyboard.SongButton(song, user, ctx.EffectiveUser.LanguageCode, songButtonOpts))
				}

				if user.Cache.PageIndex != 0 {
					markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.prev", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.next", ctx.EffectiveUser.LanguageCode)}})
				} else {
					markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.next", ctx.EffectiveUser.LanguageCode)}})
				}

				_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseSong", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
				if err != nil {
					return err
				}

				user.State.Index = 1

				return nil
			}
		case 1:
			{
				switch ctx.EffectiveMessage.Text {
				case txt.Get("button.like", ctx.EffectiveUser.LanguageCode), txt.Get("button.calendar", ctx.EffectiveUser.LanguageCode),
					txt.Get("button.numbers", ctx.EffectiveUser.LanguageCode), txt.Get("button.tag", ctx.EffectiveUser.LanguageCode):
					user.Cache.PageIndex = 0
					return c.filterSongs(0)(bot, ctx)
				case txt.Get("button.next", ctx.EffectiveUser.LanguageCode):
					user.Cache.PageIndex++
					return c.filterSongs(0)(bot, ctx)
				case txt.Get("button.prev", ctx.EffectiveUser.LanguageCode):
					user.Cache.PageIndex--
					return c.filterSongs(0)(bot, ctx)
				case keyboard.GetStatsPeriodButtonText(user.Cache.StatsPeriod, ctx.EffectiveUser.LanguageCode, false):
					user.Cache.PageIndex = 0
					return c.filterSongs(3)(bot, ctx)
				}

				if keyboard.IsSelectedButton(ctx.EffectiveMessage.Text) {
					return c.GetSongs(0)(bot, ctx)
				}

				ctx.EffectiveChat.SendAction(bot, "upload_document", nil)

				songName := keyboard.ParseSongButton(ctx.EffectiveMessage.Text)

				song, err := c.SongService.FindOneByName(strings.TrimSpace(songName))
				if err != nil {
					return c.search(0)(bot, ctx)
				}

				return c.song(bot, ctx, song.DriveFileID)
			}
		case 2:
			{
				ctx.EffectiveChat.SendAction(bot, "typing", nil)

				tags, err := c.SongService.GetTags(user.BandID)
				if err != nil {
					return err
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
				}

				filterButtons := keyboard.GetSongsStateFilterButtons(ctx.EffectiveUser.LanguageCode)
				for i := range filterButtons {
					if filterButtons[i].Text == user.Cache.Filter {
						filterButtons[i] = keyboard.SelectedButton(filterButtons[i].Text)
						break
					}
				}
				markup.Keyboard = append(markup.Keyboard, filterButtons)

				var kb [][]gotgbot.KeyboardButton
				for _, tag := range tags {
					kb = append(kb, []gotgbot.KeyboardButton{{Text: tag}})
				}

				markup.Keyboard = append(markup.Keyboard, util.SplitKeyboardToColumns(kb, 2)...)

				_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseTag", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
				if err != nil {
					return err
				}

				user.State.Index = 0
				return nil
			}
		case 3:
			{
				ctx.EffectiveChat.SendAction(bot, "typing", nil)

				periods := []entity.StatsPeriod{
					entity.StatsPeriodLastHalfYear,
					entity.StatsPeriodLastYear,
					entity.StatsPeriodAllTime,
					entity.StatsPeriodLastThreeMonths,
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
				}

				filterButtons := keyboard.GetSongsStateFilterButtons(ctx.EffectiveUser.LanguageCode)
				for i := range filterButtons {
					if filterButtons[i].Text == user.Cache.Filter {
						filterButtons[i] = keyboard.SelectedButton(filterButtons[i].Text)
						break
					}
				}
				markup.Keyboard = append(markup.Keyboard, filterButtons)

				var kb [][]gotgbot.KeyboardButton
				for _, period := range periods {
					kb = append(kb, []gotgbot.KeyboardButton{
						{Text: util.ToUpperFirstLetter(keyboard.GetStatsPeriodButtonText(period, ctx.EffectiveUser.LanguageCode, true))},
					})
				}

				markup.Keyboard = append(markup.Keyboard, kb...)

				_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseStatsPeriod", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
				if err != nil {
					return err
				}

				user.State.Index = 4
				return nil
			}
		case 4:
			ctx.EffectiveChat.SendAction(bot, "typing", nil)
			statsPeriod := keyboard.GetStatsPeriodByButtonText(ctx.EffectiveMessage.Text, ctx.EffectiveUser.LanguageCode)
			user.Cache.StatsPeriod = statsPeriod
			return c.filterSongs(0)(bot, ctx)
		}

		return c.Menu(bot, ctx)
	}
}

func (c *BotController) SongCB(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	hex := split[0]
	songID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return err
	}

	song, err := c.SongService.FindOneByID(songID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	if len(split) > 1 {
		switch split[1] {
		case "edit":
			markup.InlineKeyboard = keyboard.SongEdit(song, user, ctx.EffectiveUser.LanguageCode)
		default:
			if ctx.EffectiveMessage != nil {
				markup.InlineKeyboard = keyboard.SongInit(song, user, user.CallbackCache.ChatID, user.CallbackCache.MessageID, ctx.EffectiveUser.LanguageCode)
			} else {
				markup.InlineKeyboard = keyboard.SongInitIQ(song, user, ctx.EffectiveUser.LanguageCode)
			}
		}
	}

	opts := &gotgbot.EditMessageCaptionOpts{
		Caption:     user.CallbackCache.AddToText(song.Caption()),
		ParseMode:   "HTML",
		ReplyMarkup: markup,
	}

	if ctx.EffectiveMessage != nil {
		opts.ChatId = ctx.EffectiveMessage.Chat.Id
		opts.MessageId = ctx.EffectiveMessage.MessageId
	} else {
		opts.InlineMessageId = ctx.CallbackQuery.InlineMessageId
	}

	_, _, err = bot.EditMessageCaption(opts)
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Answer(bot, nil)

	return err
}

func (c *BotController) SongLike(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	songID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	switch split[1] {
	case "like":
		err := c.SongService.Like(songID, user.ID)
		if err != nil {
			return err
		}
	case "dislike":
		err := c.SongService.Dislike(songID, user.ID)
		if err != nil {
			return err
		}
	}

	song, err := c.SongService.FindOneByID(songID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}
	markup.InlineKeyboard = keyboard.SongInit(song, user, user.CallbackCache.ChatID, user.CallbackCache.MessageID, ctx.EffectiveUser.LanguageCode)

	_, _, err = ctx.EffectiveMessage.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Answer(bot, nil)
	return nil
}

func (c *BotController) SongVoices(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	songID, err := primitive.ObjectIDFromHex(payload)
	if err != nil {
		return err
	}

	err2 := c.songVoices(bot, ctx, songID)
	if err2 != nil {
		return err2
	}
	return nil
}

func (c *BotController) songVoices(bot *gotgbot.Bot, ctx *ext.Context, songID primitive.ObjectID) error {

	user := ctx.Data["user"].(*entity.User)

	song, err := c.SongService.FindOneByID(songID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	slices.SortStableFunc(song.Voices, func(v1, v2 *entity.Voice) int {
		if v1.Name < v2.Name {
			return -1
		} else if v1.Name > v2.Name {
			return 1
		}
		return 0
	})
	for _, voice := range song.Voices {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: voice.Name, CallbackData: util.CallbackData(state.SongVoice, song.ID.Hex()+":"+voice.ID.Hex())}})
	}

	if ctx.EffectiveMessage != nil {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.addVoice", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongVoicesCreateVoiceAskForAudio, song.ID.Hex())}})
	} else {
		deeplink, err := url.Parse(fmt.Sprintf("https://t.me/%s?start=addVoice%s", bot.Username, song.ID.Hex()))
		if err != nil {
			return err
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.addVoice", ctx.EffectiveUser.LanguageCode), Url: deeplink.String()}})
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongCB, song.ID.Hex()+":init")}})

	caption := user.CallbackCache.AddToText(txt.Get("text.chooseVoice", ctx.EffectiveUser.LanguageCode))

	opts := &gotgbot.EditMessageMediaOpts{
		ReplyMarkup: markup,
	}

	if ctx.EffectiveMessage != nil {
		opts.ChatId = ctx.EffectiveMessage.Chat.Id
		opts.MessageId = ctx.EffectiveMessage.MessageId
	} else {
		opts.InlineMessageId = ctx.CallbackQuery.InlineMessageId
	}

	_, _, err = bot.EditMessageMedia(&gotgbot.InputMediaDocument{
		Media:     song.PDF.TgFileID,
		Caption:   caption,
		ParseMode: "HTML",
	}, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *BotController) SongVoicesAddVoiceAskForAudio(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	songID, err := primitive.ObjectIDFromHex(payload)
	if err != nil {
		return err
	}

	markup := &gotgbot.ReplyKeyboardMarkup{
		Keyboard:       [][]gotgbot.KeyboardButton{{{Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}}},
		ResizeKeyboard: true,
	}
	_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.sendAudioOrVoice", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	user.State = entity.State{
		Name: state.SongVoices_CreateVoice,
	}
	user.Cache = entity.Cache{
		Voice: &entity.Voice{SongID: songID},
	}

	_, err = c.UserService.UpdateOne(*user)
	if err != nil {
		return err
	}

	return nil
}

func (c *BotController) SongVoices_CreateVoice(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.SongVoices_CreateVoice {
			user.State = entity.State{
				Index: index,
				Name:  state.SongVoices_CreateVoice,
			}
			user.Cache = entity.Cache{
				Voice: user.Cache.Voice,
			}
		}

		switch index {
		case 0:
			{
				ctx.EffectiveChat.SendAction(bot, "typing", nil)

				fileID := ""
				if ctx.EffectiveMessage.Voice != nil {
					fileID = ctx.EffectiveMessage.Voice.FileId
				} else {
					fileID = ctx.EffectiveMessage.Audio.FileId
				}
				user.Cache.Voice.FileID = fileID

				markup := &gotgbot.ReplyKeyboardMarkup{
					Keyboard:       [][]gotgbot.KeyboardButton{{{Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}}},
					ResizeKeyboard: true,
				}
				_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.sendVoiceName", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
					ReplyMarkup: markup,
				})
				if err != nil {
					return err
				}

				user.State.Index = 1
				return nil
			}
		case 1:
			{
				user.Cache.Voice.Name = ctx.EffectiveMessage.Text

				_, err := c.VoiceService.UpdateOne(*user.Cache.Voice)
				if err != nil {
					return err
				}

				_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.added", ctx.EffectiveUser.LanguageCode), nil)
				if err != nil {
					return err
				}

				song, err := c.SongService.FindOneByID(user.Cache.Voice.SongID)
				if err != nil {
					return err
				}
				err = c.song(bot, ctx, song.DriveFileID)
				if err != nil {
					return err
				}
				return c.Menu(bot, ctx)
			}
		}
		return c.Menu(bot, ctx)
	}
}

func (c *BotController) SongVoice(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	songID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	voiceID, err := primitive.ObjectIDFromHex(split[1])
	if err != nil {
		return err
	}

	song, err := c.SongService.FindOneByID(songID)
	if err != nil {
		return err
	}

	voice, err := c.VoiceService.FindOneByID(voiceID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: txt.Get("button.back", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongVoices, song.ID.Hex())},
			},
		},
	}

	if ctx.EffectiveMessage != nil {
		markup.InlineKeyboard[0] = append(markup.InlineKeyboard[0], gotgbot.InlineKeyboardButton{Text: txt.Get("button.delete", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongVoiceDeleteConfirm, song.ID.Hex()+":"+voice.ID.Hex())})
	}

	caption := user.CallbackCache.AddToText(song.Caption())

	if voice.AudioFileID == "" {
		f, err := bot.GetFile(voice.FileID, nil)
		if err != nil {
			return err
		}

		reader, err := util.File(bot, f)
		if err != nil {
			return err
		}

		opts := &gotgbot.EditMessageMediaOpts{
			ReplyMarkup: markup,
		}

		if ctx.EffectiveMessage != nil {
			opts.ChatId = ctx.EffectiveMessage.Chat.Id
			opts.MessageId = ctx.EffectiveMessage.MessageId
		} else {
			opts.InlineMessageId = ctx.CallbackQuery.InlineMessageId
		}

		msg, _, err := bot.EditMessageMedia(&gotgbot.InputMediaAudio{
			Media: gotgbot.NamedFile{
				File:     reader,
				FileName: voice.Name,
			},
			ParseMode: "HTML",
			Caption:   caption,
			Performer: song.PDF.Name,
			Title:     voice.Name,
		}, opts)
		if err != nil {
			return err
		}

		voice.AudioFileID = msg.Audio.FileId
		_, err = c.VoiceService.UpdateOne(*voice)
		if err != nil {
			return err
		}
	} else {

		opts := &gotgbot.EditMessageMediaOpts{
			ReplyMarkup: markup,
		}

		if ctx.EffectiveMessage != nil {
			opts.ChatId = ctx.EffectiveMessage.Chat.Id
			opts.MessageId = ctx.EffectiveMessage.MessageId
		} else {
			opts.InlineMessageId = ctx.CallbackQuery.InlineMessageId
		}

		_, _, err := bot.EditMessageMedia(&gotgbot.InputMediaAudio{
			Media:     voice.AudioFileID, // todo
			ParseMode: "HTML",
			Caption:   caption,
			Performer: song.PDF.Name,
			Title:     voice.Name,
		}, opts)
		if err != nil {
			return err
		}
	}

	ctx.CallbackQuery.Answer(bot, nil)
	return nil
}

func (c *BotController) SongVoiceDeleteConfirm(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	songID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	voiceID, err := primitive.ObjectIDFromHex(split[1])
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	markup.InlineKeyboard = [][]gotgbot.InlineKeyboardButton{
		{
			{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongVoice, songID.Hex()+":"+voiceID.Hex())},
			{Text: txt.Get("button.yes", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongVoiceDelete, songID.Hex()+":"+voiceID.Hex())},
		},
	}

	caption := user.CallbackCache.AddToText(txt.Get("text.voiceDeleteConfirm", ctx.EffectiveUser.LanguageCode))

	_, _, err = ctx.EffectiveMessage.EditCaption(bot,
		&gotgbot.EditMessageCaptionOpts{
			Caption:     caption,
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		})
	if err != nil {
		return err
	}
	return nil
}

func (c *BotController) SongVoiceDelete(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	songID, err := primitive.ObjectIDFromHex(split[0])
	if err != nil {
		return err
	}

	voiceID, err := primitive.ObjectIDFromHex(split[1])
	if err != nil {
		return err
	}

	err = c.VoiceService.DeleteOne(voiceID)
	if err != nil {
		return err
	}

	return c.songVoices(bot, ctx, songID)
}

func (c *BotController) SongDeleteConfirm(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	songID, err := primitive.ObjectIDFromHex(payload)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{}

	markup.InlineKeyboard = [][]gotgbot.InlineKeyboardButton{
		{
			{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongCB, songID.Hex()+":edit")},
			{Text: txt.Get("button.yes", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.SongDelete, songID.Hex())},
		},
	}

	text := user.CallbackCache.AddToText(txt.Get("text.songDeleteConfirm", ctx.EffectiveUser.LanguageCode))

	_, _, err = ctx.EffectiveMessage.EditCaption(bot, &gotgbot.EditMessageCaptionOpts{
		Caption:     text,
		ParseMode:   "HTML",
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *BotController) SongDelete(bot *gotgbot.Bot, ctx *ext.Context) error {

	//user := ctx.Data["user"].(*entity.User)

	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	songID, err := primitive.ObjectIDFromHex(payload)
	if err != nil {
		return err
	}

	song, err := c.SongService.FindOneByID(songID)
	if err != nil {
		return err
	}

	err = c.SongService.DeleteOneByDriveFileID(song.DriveFileID)
	if err != nil {
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditCaption(bot, &gotgbot.EditMessageCaptionOpts{
		Caption:   txt.Get("text.songDeleted", ctx.EffectiveUser.LanguageCode),
		ParseMode: "HTML",
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *BotController) SongCopyToMyBand(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	//ctx.EffectiveChat.SendAction(bot, "typing", nil)

	driveFileID := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	file, err := c.DriveFileService.FindOneByID(driveFileID)
	if err != nil {
		return err
	}

	file = &drive.File{
		Name:    file.Name,
		Parents: []string{user.Band.DriveFolderID},
	}

	copiedSong, err := c.DriveFileService.CloneOne(driveFileID, file)
	if err != nil {
		return err
	}

	song, _, err := c.SongService.FindOrCreateOneByDriveFileID(copiedSong.Id)
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Data = util.CallbackData(state.SongCB, song.ID.Hex()+":init")
	return c.SongCB(bot, ctx)
}

func (c *BotController) SongStyle(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	driveFileID := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	driveFile, err := c.DriveFileService.StyleOne(driveFileID)
	if err != nil {
		return err
	}

	song, err := c.SongService.FindOneByDriveFileID(driveFile.Id)
	if err != nil {
		return err
	}

	fakeTime, _ := time.Parse("2006", "2006")
	song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)

	_, err = c.SongService.UpdateOne(*song)
	if err != nil {
		return err
	}

	reader, err := c.DriveFileService.DownloadOneByID(song.DriveFileID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.SongEdit(song, user, ctx.EffectiveUser.LanguageCode),
	}

	caption := user.CallbackCache.AddToText(song.Caption())

	_, _, err = ctx.EffectiveMessage.EditMedia(bot, gotgbot.InputMediaDocument{
		Media: gotgbot.NamedFile{
			File:     *reader,
			FileName: fmt.Sprintf("%s.pdf", song.PDF.Name),
		},
		Caption:   caption,
		ParseMode: "HTML",
	}, &gotgbot.EditMessageMediaOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
		Text: txt.Get("text.styled", ctx.EffectiveUser.LanguageCode),
	})

	return nil
}

func (c *BotController) SongAddLyricsPage(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	driveFileID := util.ParseCallbackPayload(ctx.CallbackQuery.Data)

	driveFile, err := c.DriveFileService.AddLyricsPage(driveFileID)
	if err != nil {
		return err
	}

	song, err := c.SongService.FindOneByDriveFileID(driveFile.Id)
	if err != nil {
		return err
	}

	fakeTime, _ := time.Parse("2006", "2006")
	song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)

	_, err = c.SongService.UpdateOne(*song)
	if err != nil {
		return err
	}

	reader, err := c.DriveFileService.DownloadOneByID(song.DriveFileID)
	if err != nil {
		return err
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.SongEdit(song, user, ctx.EffectiveUser.LanguageCode),
	}

	caption := user.CallbackCache.AddToText(song.Caption())

	_, _, err = ctx.EffectiveMessage.EditMedia(bot, gotgbot.InputMediaDocument{
		Media: gotgbot.NamedFile{
			File:     *reader,
			FileName: fmt.Sprintf("%s.pdf", song.PDF.Name),
		},
		Caption:   caption,
		ParseMode: "HTML",
	}, &gotgbot.EditMessageMediaOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
		Text: txt.Get("text.addedLyricsPage", ctx.EffectiveUser.LanguageCode),
	})

	return nil
}
