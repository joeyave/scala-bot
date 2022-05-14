package controller

import (
	"encoding/json"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/gorilla/schema"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/service"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BotController struct {
	UserService       *service.UserService
	DriveFileService  *service.DriveFileService
	SongService       *service.SongService
	VoiceService      *service.VoiceService
	BandService       *service.BandService
	MembershipService *service.MembershipService
	EventService      *service.EventService
	RoleService       *service.RoleService
	//OldHandler        *myhandlers.Handler
}

func (c *BotController) ChooseHandlerOrSearch(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	switch user.State.Name {
	case state.GetEvents:
		return c.GetEvents(user.State.Index)(bot, ctx)
	case state.FilterEvents:
		return c.filterEvents(user.State.Index)(bot, ctx)
	case state.GetSongs:
		return c.GetSongs(user.State.Index)(bot, ctx)
	case state.FilterSongs:
		return c.filterSongs(user.State.Index)(bot, ctx)
	case state.SearchSetlist:
		return c.searchSetlist(user.State.Index)(bot, ctx)
	case state.SongVoicesCreateVoice:
		return c.SongVoicesCreateVoice(user.State.Index)(bot, ctx)
	case state.BandCreate:
		return c.BandCreate(user.State.Index)(bot, ctx)
	case state.RoleCreate_ChoosePosition:
		return c.RoleCreate_ChoosePosition(bot, ctx)
	}

	return c.search(user.State.Index)(bot, ctx)
}

var decoder = schema.NewDecoder()

func (c *BotController) RegisterUser(bot *gotgbot.Bot, ctx *ext.Context) error {

	user, err := c.UserService.FindOneOrCreateByID(ctx.EffectiveUser.Id)
	if err != nil {
		return err
	}
	ctx.Data["user"] = user
	user = ctx.Data["user"].(*entity.User)

	user.Name = strings.TrimSpace(fmt.Sprintf("%s %s", ctx.EffectiveUser.FirstName, ctx.EffectiveUser.LastName))

	// todo
	//if user.BandID == primitive.NilObjectID && user.State.Name != helpers.ChooseBandState && user.State.Name != helpers.CreateBandState {
	//	user.State = entity.State{
	//		Name: helpers.ChooseBandState,
	//	}
	//}
	if user.BandID == primitive.NilObjectID || user.Band == nil {

		if ctx.CallbackQuery != nil {
			parsedData := strings.Split(ctx.CallbackQuery.Data, ":")
			if parsedData[0] == strconv.Itoa(state.SettingsChooseBand) || parsedData[0] == strconv.Itoa(state.BandCreate_AskForName) {
				return nil
			}
		} else if user.State.Name == state.BandCreate {
			return nil
		}

		markup := gotgbot.InlineKeyboardMarkup{}

		bands, err := c.BandService.FindAll()
		if err != nil {
			return err
		}
		for _, band := range bands {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: band.Name, CallbackData: util.CallbackData(state.SettingsChooseBand, band.ID.Hex())}})
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: txt.Get("button.createBand", ctx.EffectiveUser.LanguageCode), CallbackData: util.CallbackData(state.BandCreate_AskForName, "")}})

		_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseBand", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
			ReplyMarkup: markup,
		})
		if err != nil {
			return err
		}

		return ext.EndGroups
	}

	if ctx.CallbackQuery != nil && ctx.CallbackQuery.Message != nil {
		for _, e := range ctx.CallbackQuery.Message.Entities {
			if strings.HasPrefix(e.Url, util.CallbackCacheURL) {
				u, err := url.Parse(e.Url)
				if err != nil {
					return err
				}
				err = decoder.Decode(&user.CallbackCache, u.Query())
				if err != nil {
					return err
				}
				break
			}
		}
		for _, e := range ctx.CallbackQuery.Message.CaptionEntities {
			if strings.HasPrefix(e.Url, util.CallbackCacheURL) {
				u, err := url.Parse(e.Url)
				if err != nil {
					return err
				}
				err = decoder.Decode(&user.CallbackCache, u.Query())
				if err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}

func (c *BotController) UpdateUser(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	_, err := c.UserService.UpdateOne(*user)
	return err
}

func (c *BotController) search(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.Search {
			user.State = entity.State{
				Index: index,
				Name:  state.Search,
			}
			user.Cache = entity.Cache{}
		}

		switch index {
		case 0:
			{
				ctx.EffectiveChat.SendAction(bot, "typing")

				var query string

				switch ctx.EffectiveMessage.Text {
				case txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode):
					user.Cache.Filter = ctx.EffectiveMessage.Text
					query = user.Cache.Query
				case txt.Get("button.prev", ctx.EffectiveUser.LanguageCode):
					query = user.Cache.Query
					user.Cache.NextPageToken = user.Cache.NextPageToken.Prev.Prev
				case txt.Get("button.next", ctx.EffectiveUser.LanguageCode):
					query = user.Cache.Query
				default:
					query = ctx.EffectiveMessage.Text
					// Обнуляем страницы при новом запросе.
					user.Cache.NextPageToken = nil
				}

				query = util.CleanUpText(query)
				songNames := util.SplitTextByNewlines(query)

				if len(songNames) > 1 {
					user.Cache.SongNames = songNames
					return c.searchSetlist(0)(bot, ctx)

				} else if len(songNames) == 1 {
					query = songNames[0]
					user.Cache.Query = query
				} else {
					_, err := ctx.EffectiveChat.SendMessage(bot, "Из запроса удаляются все числа, дефисы и скобки вместе с тем, что в них.", nil)
					return err
				}

				var (
					driveFiles    []*drive.File
					nextPageToken string
					err           error
				)
				switch user.Cache.Filter {
				case txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode):
					driveFiles, nextPageToken, err = c.DriveFileService.FindSomeByFullTextAndFolderID(query, "", user.Cache.NextPageToken.GetValue())
				default:
					driveFiles, nextPageToken, err = c.DriveFileService.FindSomeByFullTextAndFolderID(query, user.Band.DriveFolderID, user.Cache.NextPageToken.GetValue())
				}
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
					_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.nothingFound", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{ReplyMarkup: markup})
					return err
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: query,
				}
				if markup.InputFieldPlaceholder == "" {
					markup.InputFieldPlaceholder = txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode)
				}

				likedSongs, likedSongErr := c.SongService.FindManyLiked(user.ID)

				set := make(map[string]*entity.Band)
				for i, driveFile := range driveFiles {

					if user.Cache.Filter == txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode) {

						for _, parentFolderID := range driveFile.Parents {
							_, exists := set[parentFolderID]
							if !exists {
								band, err := c.BandService.FindOneByDriveFolderID(parentFolderID)
								if err == nil {
									set[parentFolderID] = band
									driveFiles[i].Name += fmt.Sprintf(" (%s)", band.Name)
									break
								}
							} else {
								driveFiles[i].Name += fmt.Sprintf(" (%s)", set[parentFolderID].Name)
							}
						}
					}

					opts := &keyboard.DriveFileButtonOpts{
						ShowLike: true,
					}
					if likedSongErr != nil {
						opts.ShowLike = false
					}
					markup.Keyboard = append(markup.Keyboard, keyboard.DriveFileButton(driveFile, likedSongs, opts))
				}

				if ctx.EffectiveMessage.Text != txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode) {
					markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode)}})
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
				case txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode), txt.Get("button.next", ctx.EffectiveUser.LanguageCode):
					return c.search(0)(bot, ctx)
				}

				ctx.EffectiveChat.SendAction(bot, "upload_document")

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

func (c *BotController) searchSetlist(index int) handlers.Response {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {
		user := ctx.Data["user"].(*entity.User)

		if user.State.Name != state.SearchSetlist {
			user.State = entity.State{
				Index: index,
				Name:  state.SearchSetlist,
			}
			user.Cache = entity.Cache{
				SongNames: user.Cache.SongNames,
			}
		}

		switch index {
		case 0:
			{
				if len(user.Cache.SongNames) < 1 {
					ctx.EffectiveChat.SendAction(bot, "upload_document")

					err := c.songsAlbum(bot, ctx, user.Cache.DriveFileIDs)
					if err != nil {
						return err
					}
					return c.Menu(bot, ctx)
				}

				ctx.EffectiveChat.SendAction(bot, "typing")

				songNames := user.Cache.SongNames

				currentSongName := songNames[0]
				user.Cache.SongNames = songNames[1:]

				driveFiles, _, err := c.DriveFileService.FindSomeByFullTextAndFolderID(currentSongName, user.Band.DriveFolderID, "")
				if err != nil {
					return err
				}

				if len(driveFiles) == 0 {
					markup := &gotgbot.ReplyKeyboardMarkup{
						Keyboard: [][]gotgbot.KeyboardButton{
							{{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.skip", ctx.EffectiveUser.LanguageCode)}},
						},
						ResizeKeyboard: true,
					}

					_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.nothingFoundByQuery", ctx.EffectiveUser.LanguageCode, currentSongName), &gotgbot.SendMessageOpts{
						ReplyMarkup: markup,
					})
					if err != nil {
						return err
					}

					user.State.Index = 1
					return nil
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: currentSongName,
				}

				for _, song := range driveFiles {
					markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{{Text: song.Name}})
				}
				markup.Keyboard = append(markup.Keyboard, []gotgbot.KeyboardButton{
					{Text: txt.Get("button.cancel", ctx.EffectiveUser.LanguageCode)}, {Text: txt.Get("button.skip", ctx.EffectiveUser.LanguageCode)},
				})

				_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.chooseSongOrTypeAnotherQuery", ctx.EffectiveUser.LanguageCode, currentSongName), &gotgbot.SendMessageOpts{
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
				switch ctx.EffectiveMessage.Text {
				case txt.Get("button.skip", ctx.EffectiveUser.LanguageCode):
					return c.searchSetlist(0)(bot, ctx)
				}

				foundDriveFile, err := c.DriveFileService.FindOneByNameAndFolderID(ctx.EffectiveMessage.Text, user.Band.DriveFolderID)
				if err != nil {
					user.Cache.SongNames = append([]string{ctx.EffectiveMessage.Text}, user.Cache.SongNames...)
				} else {
					user.Cache.DriveFileIDs = append(user.Cache.DriveFileIDs, foundDriveFile.Id)
				}

				return c.searchSetlist(0)(bot, ctx)
			}
		}
		return c.Menu(bot, ctx)
	}
}

func (c *BotController) Menu(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	user.State = entity.State{}
	user.Cache = entity.Cache{}

	replyMarkup := &gotgbot.ReplyKeyboardMarkup{
		Keyboard:              keyboard.Menu(ctx.EffectiveUser.LanguageCode),
		ResizeKeyboard:        true,
		InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
	}

	_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.menu", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
		ReplyMarkup: replyMarkup,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *BotController) Error(bot *gotgbot.Bot, ctx *ext.Context, botErr error) ext.DispatcherAction {

	log.Error().Msgf("Error handling update: %v", botErr)

	user, err := c.UserService.FindOneByID(ctx.EffectiveUser.Id)
	if err != nil {
		log.Error().Err(err).Msg("Error!")
		return ext.DispatcherActionEndGroups
	}

	if ctx.CallbackQuery != nil {
		_, err := ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Произошла ошибка. Поправим.",
		})
		if err != nil {
			log.Error().Err(err).Msg("Error!")
			return ext.DispatcherActionEndGroups
		}
	} else if ctx.InlineQuery != nil {
		ctx.InlineQuery.Answer(bot, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime: 1,
		})
	} else if ctx.EffectiveChat != nil {
		_, err := ctx.EffectiveChat.SendMessage(bot, "Произошла ошибка. Поправим.", nil)
		if err != nil {
			log.Error().Err(err).Msg("Error!")
			return ext.DispatcherActionEndGroups
		}
		user.State = entity.State{}
		_, err = c.UserService.UpdateOne(*user)
		if err != nil {
			log.Error().Err(err).Msg("Error!")
			return ext.DispatcherActionEndGroups
		}
	}

	// todo: send message to the logs channel
	logsChannelID, err := strconv.ParseInt(os.Getenv("LOG_CHANNEL"), 10, 64)
	if err == nil {
		userJsonBytes, err := json.Marshal(user)
		if err != nil {
			log.Error().Err(err).Msg("Error!")
			return ext.DispatcherActionEndGroups
		}

		_, err = bot.SendMessage(logsChannelID, fmt.Sprintf("Error handling update!\n<pre>error=%v</pre>\n<pre>user=%s</pre>", botErr, string(userJsonBytes)), &gotgbot.SendMessageOpts{
			DisableWebPagePreview: true,
			ParseMode:             "HTML",
		})
		if err != nil {
			log.Error().Err(err).Msg("Error!")
			return ext.DispatcherActionEndGroups
		}
	}

	return ext.DispatcherActionEndGroups
}

func (c *BotController) songsAlbum(bot *gotgbot.Bot, ctx *ext.Context, driveFileIDs []string) error {

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(driveFileIDs))
	bigAlbum := make([]gotgbot.InputMedia, len(driveFileIDs))

	for i := range driveFileIDs {
		go func(i int) {
			defer waitGroup.Done()

			song, driveFile, err := c.SongService.FindOrCreateOneByDriveFileID(driveFileIDs[i])
			if err != nil {
				return
			}

			if song.PDF.TgFileID == "" {
				reader, err := c.DriveFileService.DownloadOneByID(driveFile.Id)
				if err != nil {
					return
				}

				bigAlbum[i] = gotgbot.InputMediaDocument{
					Media:   gotgbot.NamedFile{File: *reader, FileName: fmt.Sprintf("%s.pdf", song.PDF.Name)},
					Caption: song.Meta(),
				}
			} else {
				bigAlbum[i] = gotgbot.InputMediaDocument{
					Media:   song.PDF.TgFileID,
					Caption: song.Meta(),
				}
			}
		}(i)
	}
	waitGroup.Wait()

	inputMediaChunks := chunkAlbum(bigAlbum, 10)
	driveFileIDsChunks := chunk(driveFileIDs, 10)

	for i := range inputMediaChunks {
		_, err := bot.SendMediaGroup(ctx.EffectiveChat.Id, inputMediaChunks[i], nil)
		if err != nil { // todo: try to download all
			var waitGroup sync.WaitGroup
			waitGroup.Add(len(driveFileIDsChunks[i]))
			inputMedia := make([]gotgbot.InputMedia, len(driveFileIDsChunks[i]))

			for i := range driveFileIDsChunks[i] {
				go func(i int) {
					defer waitGroup.Done()

					song, _, err := c.SongService.FindOrCreateOneByDriveFileID(driveFileIDs[i])
					if err != nil {
						return
					}

					reader, err := c.DriveFileService.DownloadOneByID(driveFileIDs[i])
					if err != nil {
						return
					}

					inputMedia[i] = gotgbot.InputMediaDocument{
						Media:   gotgbot.NamedFile{File: *reader, FileName: fmt.Sprintf("%s.pdf", song.PDF.Name)},
						Caption: song.Meta(),
					}
				}(i)
			}
			waitGroup.Wait()

			_, err = bot.SendMediaGroup(ctx.EffectiveChat.Id, inputMedia, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func chunkAlbum(items []gotgbot.InputMedia, chunkSize int) (chunks [][]gotgbot.InputMedia) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}
func chunk(items []string, chunkSize int) (chunks [][]string) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}

func (c *BotController) NotifyUsers(bot *gotgbot.Bot) {
	for range time.Tick(time.Hour * 2) {
		events, err := c.EventService.FindAllFromToday()
		if err != nil {
			return
		}

		for _, event := range events {
			if event.Time.Add(time.Hour*8).Sub(time.Now()).Hours() < 48 {
				for _, membership := range event.Memberships {
					if membership.Notified == true {
						continue
					}

					markup := gotgbot.InlineKeyboardMarkup{
						InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{{Text: "ℹ️ Подробнее", CallbackData: util.CallbackData(state.EventCB, event.ID.Hex()+":init")}}},
					}

					text := fmt.Sprintf("Привет. Ты учавствуешь в собрании через несколько дней (%s)!", event.Alias("ru"))

					_, err = bot.SendMessage(membership.UserID, text, &gotgbot.SendMessageOpts{
						ParseMode:   "HTML",
						ReplyMarkup: markup,
					})
					if err != nil {
						continue
					}

					membership.Notified = true
					c.MembershipService.UpdateOne(*membership)
				}
			}
		}
	}
}
