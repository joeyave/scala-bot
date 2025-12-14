package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/gorilla/schema"
	"github.com/hbollon/go-edlib"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/service"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/drive/v3"
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
	case state.SongVoices_CreateVoice:
		return c.SongVoices_CreateVoice(user.State.Index)(bot, ctx)
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
	user.LanguageCode = ctx.EffectiveUser.LanguageCode

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
		message, ok := ctx.CallbackQuery.Message.(gotgbot.Message)
		if ok {
			for _, e := range message.Entities {
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
			for _, e := range message.CaptionEntities {
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

		bandIndex := slices.IndexFunc(user.Cache.Bands, func(band *entity.Band) bool {
			return band.Name == ctx.EffectiveMessage.Text
		})
		if bandIndex != -1 {
			chosenBand := user.Cache.Bands[bandIndex]
			user.BandID = chosenBand.ID
			return c.Menu(bot, ctx)
		}

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
				_, _ = ctx.EffectiveChat.SendAction(bot, "typing", nil)

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
					_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.cleanQueryInfo", ctx.EffectiveUser.LanguageCode), nil)
					return err
				}

				var (
					driveFiles    []*drive.File
					nextPageToken string
					err           error
				)
				switch user.Cache.Filter {
				case txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode):
					driveFiles, nextPageToken, err = c.DriveFileService.FindSomeByFullTextAndFolderID(query, []string{}, user.Cache.NextPageToken.GetValue())
				default:
					driveFiles, nextPageToken, err = c.DriveFileService.FindSomeByFullTextAndFolderID(query, []string{user.Band.DriveFolderID, user.Band.ArchiveFolderID}, user.Cache.NextPageToken.GetValue())
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
				} else if len(driveFiles) == 1 { // todo: test this case.
					go c.songByDriveFile(bot, ctx, driveFiles[0]) //nolint:errcheck
				}

				markup := &gotgbot.ReplyKeyboardMarkup{
					ResizeKeyboard:        true,
					InputFieldPlaceholder: query,
				}
				if markup.InputFieldPlaceholder == "" {
					markup.InputFieldPlaceholder = txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode)
				}

				likedSongs, likedSongErr := c.SongService.FindManyLiked(user.BandID, user.ID)

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

				if len(driveFiles) == 1 {
					_, _ = ctx.EffectiveChat.SendAction(bot, "upload_document", nil)
				}
				return nil
			}
		case 1:
			{
				switch ctx.EffectiveMessage.Text {
				case txt.Get("button.globalSearch", ctx.EffectiveUser.LanguageCode), txt.Get("button.next", ctx.EffectiveUser.LanguageCode):
					return c.search(0)(bot, ctx)
				}

				_, _ = ctx.EffectiveChat.SendAction(bot, "upload_document", nil)

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
					return c.songByDriveFileID(bot, ctx, foundDriveFile.Id)
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
					_, _ = ctx.EffectiveChat.SendAction(bot, "upload_document", nil)

					songs, _, err := c.SongService.FindOrCreateManyByDriveFileIDs(user.Cache.DriveFileIDs)
					if err != nil {
						return err
					}

					err = c.songsAlbum(bot, ctx, songs)
					if err != nil {
						return err
					}
					return c.Menu(bot, ctx)
				}

				_, _ = ctx.EffectiveChat.SendAction(bot, "typing", nil)

				songNames := user.Cache.SongNames

				currentSongName := songNames[0]
				user.Cache.SongNames = songNames[1:]

				driveFiles, _, err := c.DriveFileService.FindSomeByFullTextAndFolderID(currentSongName, []string{user.Band.DriveFolderID, user.Band.ArchiveFolderID}, "")
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

				foundDriveFile, err := c.DriveFileService.FindOneByNameAndFolderID(ctx.EffectiveMessage.Text, []string{user.Band.DriveFolderID, user.Band.ArchiveFolderID})
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

	bands, err := c.BandService.FindManyByIDs(user.BandIDs)
	if err != nil {
		return err
	}

	user.Cache.Bands = bands

	replyMarkup := &gotgbot.ReplyKeyboardMarkup{
		Keyboard:              keyboard.Menu(user, bands, ctx.EffectiveUser.LanguageCode),
		ResizeKeyboard:        true,
		InputFieldPlaceholder: txt.Get("text.defaultPlaceholder", ctx.EffectiveUser.LanguageCode),
	}

	_, err = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.menu", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
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
		_, _ = ctx.CallbackQuery.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: txt.Get("text.serverError", ctx.EffectiveUser.LanguageCode),
		})
		//if err != nil {
		//	log.Error().Err(err).Msg("Error!")
		//	return ext.DispatcherActionEndGroups
		//}
	} else if ctx.InlineQuery != nil {
		_, _ = ctx.InlineQuery.Answer(bot, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime: 1,
		})

	} else if ctx.EffectiveChat != nil {
		_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.serverError", ctx.EffectiveUser.LanguageCode), nil)
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
	logsChannelID, err := strconv.ParseInt(os.Getenv("BOT_ALERTS_CHANNEL_ID"), 10, 64)
	if err == nil {
		userJsonBytes, err := json.Marshal(user)
		if err != nil {
			log.Error().Err(err).Msg("Error!")
			return ext.DispatcherActionEndGroups
		}

		_, err = bot.SendMessage(logsChannelID, fmt.Sprintf("Error handling update!\n<pre>error=%v</pre>\n<pre>user=%s</pre>", botErr, string(userJsonBytes)), &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: "HTML",
		})
		if err != nil {
			log.Error().Err(err).Msg("Error!")
			return ext.DispatcherActionEndGroups
		}
	}

	return ext.DispatcherActionEndGroups
}

func (c *BotController) buildSongsMediaGroup(songs []*entity.Song, downloadAll bool) ([]gotgbot.InputMedia, []io.Closer, error) {
	g := new(errgroup.Group)

	album := make([]gotgbot.InputMedia, len(songs))
	closers := make([]io.Closer, len(songs))

	for i, song := range songs {
		i, song := i, song // Important! See https://golang.org/doc/faq#closures_and_goroutines.
		g.Go(func() error {
			if song.PDF.TgFileID == "" || downloadAll {
				reader, err := c.DriveFileService.DownloadOneByID(song.DriveFileID)
				if err != nil {
					return err
				}

				album[i] = gotgbot.InputMediaDocument{
					Media:   gotgbot.InputFileByReader(fmt.Sprintf("%s.pdf", song.PDF.Name), reader),
					Caption: song.Meta(),
				}
				closers[i] = reader
			} else {
				album[i] = gotgbot.InputMediaDocument{
					Media:   gotgbot.InputFileByID(song.PDF.TgFileID),
					Caption: song.Meta(),
				}
			}

			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		for _, closer := range closers {
			if closer != nil {
				_ = closer.Close()
			}
		}
		return nil, nil, err
	}

	return album, closers, err
}

func (c *BotController) updateSongTgFileIDs(msgs []gotgbot.Message, songs []*entity.Song) {
	if len(songs) != len(msgs) {
		return
	}

	for j, song := range songs {
		doc := msgs[j].Document
		if doc == nil {
			continue
		}

		// На всякий случай сравниваем названия. Сравниваем с помощью алгоритма Levenshtein.
		//Получаем процент схожести двух строк. Пропускаем больше 90%.
		str1 := song.PDF.Name
		str2 := strings.ReplaceAll(strings.TrimSuffix(doc.FileName, ".pdf"), "_", " ")
		similarity, err := edlib.StringsSimilarity(str1, str2, edlib.Levenshtein)
		fmt.Printf("similarity: %g, str1: %s, str2: %s\n", similarity, str1, str2)
		if err == nil && similarity > 0.9 {
			song.PDF.TgFileID = doc.FileId
		}
	}
	_, err := c.SongService.UpdateMany(songs)
	if err != nil {
		return
	}
}

func (c *BotController) songsAlbum(bot *gotgbot.Bot, ctx *ext.Context, songs []*entity.Song) error {

	// Подготавливаем файлы: берем из кеша (если файл уже есть на серверах телеграм) или загружаем.
	mediaGroup, closers, err := c.buildSongsMediaGroup(songs, false)
	if err != nil {
		return err
	}

	// Если файлов больше чем 10, разделяем их на чанки и отправляем по очереди (ограничения ТГ).
	mediaGroupChunks := helpers.Chunk(mediaGroup, 10)
	closersChunks := helpers.Chunk(closers, 10)
	songsChunks := helpers.Chunk(songs, 10)

	for i, mediaGroupChunk := range mediaGroupChunks {
		closersChunk := closersChunks[i]
		_, err := bot.SendMediaGroup(ctx.EffectiveChat.Id, mediaGroupChunk, nil)
		for _, closer := range closersChunk {
			if closer != nil {
				_ = closer.Close()
			}
		}
		// Если не смогли отправить чанк, возможно проблема с файлами из кеша - TgFileID невалидный.
		// Попробуем скачать все файлы из чанка и отправить.
		if err != nil {
			songsChunk := songsChunks[i]
			mediaGroup, closers, err := c.buildSongsMediaGroup(songsChunk, true)
			if err != nil {
				return err
			}
			msgs, err := bot.SendMediaGroup(ctx.EffectiveChat.Id, mediaGroup, nil)
			for _, closer := range closers {
				if closer != nil {
					_ = closer.Close()
				}
			}
			if err != nil {
				return err
			}

			// Попробуем обновить TgFileID у файлов.
			go c.updateSongTgFileIDs(msgs, songsChunk)
		}
	}

	return nil
}

func (c *BotController) NotifyUsers(bot *gotgbot.Bot) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C

		now := time.Now()
		from := time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0,
			now.Location())
		to := from.Add(48 * time.Hour)

		// Load upcoming events (48h window)
		events, err := c.EventService.FindBetweenDates(from, to)
		if err != nil {
			log.Error().Err(err).Msg("Failed to load upcoming events")
			continue
		}

		for _, event := range events {

			// NEW: Do not send notifications at night (local event time)
			if !canSendNow(event) {
				continue
			}

			// Loop through memberships
			for _, membership := range event.Memberships {
				if membership.Notified {
					continue
				}

				markup := gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
						{
							Text:         txt.Get("button.moreInfo", membership.User.LanguageCode),
							CallbackData: util.CallbackData(state.EventCB, event.ID.Hex()+":init"),
						},
					}},
				}

				text := txt.Get("text.upcomingEventNotification", membership.User.LanguageCode, event.Alias(membership.User.LanguageCode))

				_, err = bot.SendMessage(membership.UserID, text, &gotgbot.SendMessageOpts{
					ParseMode:   "HTML",
					ReplyMarkup: markup,
				})
				if err != nil {
					log.Error().Err(err).Msgf("Failed to send notification to user %d", membership.UserID)
					continue
				}

				membership.Notified = true
				if _, err := c.MembershipService.UpdateOne(*membership); err != nil {
					log.Error().Err(err).Msg("Failed to update membership after sending notification")
				}
			}
		}
	}
}

// Do not send notifications between 21:00–08:00 (event local time)
func canSendNow(event *entity.Event) bool {
	now := event.Band.GetNowTime()
	hour := now.Hour()
	return hour >= 7 && hour < 21
}
