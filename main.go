package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/inlinequery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/controller"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/repository"
	"github.com/joeyave/scala-bot/service"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
	_ "time/tzdata" // This is mandatory for AWS containers.
)

func main() {

	location, err := time.LoadLocation("Europe/Kiev")
	if err != nil {
		log.Fatal().Msgf("Err loading location: %v", err)
	}
	time.Local = location

	out := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	log.Logger = zerolog.New(out).Level(zerolog.GlobalLevel()).With().Timestamp().Logger()

	// Create bot from environment value.
	bot, err := gotgbot.NewBot(os.Getenv("BOT_TOKEN"), &gotgbot.BotOpts{})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	_, err = bot.SetMyCommands([]gotgbot.BotCommand{
		{Command: "/schedule", Description: txt.Get("button.schedule", "ru")},
		{Command: "/songs", Description: txt.Get("button.songs", "ru")},
		{Command: "/menu", Description: txt.Get("button.menu", "ru")},
	}, &gotgbot.SetMyCommandsOpts{Scope: gotgbot.BotCommandScopeDefault{}, LanguageCode: "ru"})
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting commands")
	}

	_, err = bot.SetMyCommands([]gotgbot.BotCommand{
		{Command: "/schedule", Description: txt.Get("button.schedule", "")},
		{Command: "/songs", Description: txt.Get("button.songs", "")},
		{Command: "/menu", Description: txt.Get("button.menu", "")},
	}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting commands")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("BOT_MONGODB_URI")))
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to MongoDB")
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Fatal().Err(err).Msg("Error disconnecting from MongoDB")
		}
	}()

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal().Err(err).Msg("Error pinging MongoDB")
	}

	driveRepository, err := drive.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("BOT_GOOGLEAPIS_KEY"))))
	if err != nil {
		log.Fatal().Msgf("Unable to retrieve Drive client: %v", err)
	}

	docsRepository, err := docs.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("BOT_GOOGLEAPIS_KEY"))))
	if err != nil {
		log.Fatal().Msgf("Unable to retrieve Docs client: %v", err)
	}

	voiceRepository := repository.NewVoiceRepository(mongoClient)
	voiceService := service.NewVoiceService(voiceRepository)

	bandRepository := repository.NewBandRepository(mongoClient)
	bandService := service.NewBandService(bandRepository)

	driveFileService := service.NewDriveFileService(driveRepository, docsRepository)

	songRepository := repository.NewSongRepository(mongoClient)
	songService := service.NewSongService(songRepository, voiceRepository, bandRepository, driveRepository, driveFileService)

	userRepository := repository.NewUserRepository(mongoClient)
	userService := service.NewUserService(userRepository)

	membershipRepository := repository.NewMembershipRepository(mongoClient)
	membershipService := service.NewMembershipService(membershipRepository)

	eventRepository := repository.NewEventRepository(mongoClient)
	eventService := service.NewEventService(eventRepository, membershipRepository, driveFileService)

	roleRepository := repository.NewRoleRepository(mongoClient)
	roleService := service.NewRoleService(roleRepository)

	//handler := myhandlers.NewHandler(
	//	bot,
	//	userService,
	//	driveFileService,
	//	songService,
	//	voiceService,
	//	bandService,
	//	membershipService,
	//	eventService,
	//	roleService,
	//)

	botController := controller.BotController{
		//OldHandler:        handler,
		UserService:       userService,
		DriveFileService:  driveFileService,
		SongService:       songService,
		VoiceService:      voiceService,
		BandService:       bandService,
		MembershipService: membershipService,
		EventService:      eventService,
		RoleService:       roleService,
	}
	webAppController := controller.WebAppController{
		Bot: bot,

		UserService:       userService,
		DriveFileService:  driveFileService,
		SongService:       songService,
		VoiceService:      voiceService,
		BandService:       bandService,
		MembershipService: membershipService,
		EventService:      eventService,
		RoleService:       roleService,
	}
	driveFileController := controller.DriveFileController{
		DriveFileService: driveFileService,
		SongService:      songService,
	}

	// Create updater and dispatcher.
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: botController.Error,
		Panic: func(b *gotgbot.Bot, ctx *ext.Context, r interface{}) {
			err, ok := r.(error)
			if ok {
				botController.Error(bot, ctx, err)
			} else {
				botController.Error(bot, ctx, fmt.Errorf("panic: %s", fmt.Sprint(r)))
			}
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)

	dispatcher.AddHandlerToGroup(handlers.NewInlineQuery(inlinequery.All, func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user, err := userService.FindOneByID(ctx.EffectiveUser.Id)
		if err == nil && user.BandID != primitive.NilObjectID {
			ctx.Data["user"] = user
			return nil
		}

		_, err = ctx.InlineQuery.Answer(bot, nil, &gotgbot.AnswerInlineQueryOpts{
			Button: &gotgbot.InlineQueryResultsButton{
				Text:           txt.Get("text.selectOrCreateBand", ctx.EffectiveUser.LanguageCode), // todo: put in txt
				StartParameter: "test",
			},
		})
		if err != nil {
			return err
		}

		return ext.EndGroups
	}), 0)

	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, botController.RegisterUser), 0)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(callbackquery.All, botController.RegisterUser), 0)

	// Plain keyboard.
	dispatcher.AddHandlerToGroup(handlers.NewCommand("start", func(bot *gotgbot.Bot, ctx *ext.Context) error {
		fmt.Println(ctx)

		if strings.Contains(ctx.EffectiveMessage.Text, "addVoice") {
			split := strings.Split(ctx.EffectiveMessage.Text, "addVoice")
			ctx.CallbackQuery = &gotgbot.CallbackQuery{
				Data: "0:" + split[1],
			}
			return botController.SongVoicesAddVoiceAskForAudio(bot, ctx)
		}

		return botController.Menu(bot, ctx)
		//return nil
	}), 1)

	dispatcher.AddHandlerToGroup(handlers.NewCommand("schedule", botController.GetEvents(0)), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCommand("songs", botController.GetSongs(0)), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCommand("menu", botController.Menu), 1)

	dispatcher.AddHandlerToGroup(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return msg.Text == txt.Get("button.menu", msg.From.LanguageCode) || msg.Text == txt.Get("button.cancel", msg.From.LanguageCode)
	}, botController.Menu), 1)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return msg.Text == txt.Get("button.schedule", msg.From.LanguageCode)
	}, botController.GetEvents(0)), 1)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return msg.Text == txt.Get("button.songs", msg.From.LanguageCode)
	}, botController.GetSongs(0)), 1)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return msg.Text == txt.Get("button.stats", msg.From.LanguageCode)
	}, func(bot *gotgbot.Bot, ctx *ext.Context) error {
		_, _ = ctx.EffectiveChat.SendMessage(bot, txt.Get("text.noStats", ctx.EffectiveUser.LanguageCode), nil)
		return nil
	}), 1)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return msg.Text == txt.Get("button.settings", msg.From.LanguageCode)
	}, botController.Settings), 1)

	// Web app.
	dispatcher.AddHandlerToGroup(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return msg.WebAppData != nil && msg.WebAppData.ButtonText == txt.Get("button.createEvent", msg.From.LanguageCode)
	}, botController.CreateEvent), 1)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return msg.WebAppData != nil && msg.WebAppData.ButtonText == txt.Get("button.createDoc", msg.From.LanguageCode)
	}, botController.CreateSong), 1)

	// Callback query.
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.BandCreate_AskForName), botController.BandCreate_AskForName), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.RoleCreate_AskForName), botController.RoleCreate_AskForName), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.RoleCreate), botController.RoleCreate), 1)

	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsCB), botController.SettingsCB), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsBands), botController.SettingsBands), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsChooseBand), botController.SettingsChooseBand), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsBandMembers), botController.SettingsBandMembers), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsCleanupDatabase), botController.SettingsCleanupDatabase), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsBandAddAdmin), botController.SettingsBandAddAdmin), 1)

	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventCB), botController.EventCB), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventSetlistDocs), botController.EventSetlistDocs), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventSetlistMetronome), botController.EventSetlistMetronome), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventSetlist), botController.EventSetlist), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventSetlistDeleteOrRecoverSong), botController.EventSetlistDeleteOrRecoverSong), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventMembers), botController.EventMembers), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventMembersDeleteOrRecoverMember), botController.EventMembersDeleteOrRecoverMember), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventMembersAddMemberChooseRole), botController.EventMembersAddMemberChooseRole), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventMembersAddMemberChooseUser), botController.EventMembersAddMemberChooseUser), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventMembersAddMember), botController.EventMembersAddMember), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventMembersDeleteMember), botController.EventMembersDeleteMember), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventDeleteConfirm), botController.EventDeleteConfirm), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.EventDelete), botController.EventDelete), 1)

	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongCB), botController.SongCB), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongLike), botController.SongLike), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongArchive), botController.SongArchive), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoices), botController.SongVoices), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoicesCreateVoiceAskForAudio), botController.SongVoicesAddVoiceAskForAudio), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoice), botController.SongVoice), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoiceDeleteConfirm), botController.SongVoiceDeleteConfirm), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoiceDelete), botController.SongVoiceDelete), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongStats), botController.SongStats), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongDeleteConfirm), botController.SongDeleteConfirm), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongDelete), botController.SongDelete), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongCopyToMyBand), botController.SongCopyToMyBand), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongStyle), botController.SongStyle), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongAddLyricsPage), botController.SongAddLyricsPage), 1)

	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.TransposeAudio_AskForSemitonesNumber), botController.TransposeAudio_AskForSemitonesNumber), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.TransposeAudio), botController.TransposeAudio), 1)

	// Inline query.
	dispatcher.AddHandlerToGroup(handlers.NewInlineQuery(inlinequery.All, func(bot *gotgbot.Bot, ctx *ext.Context) error {

		user := ctx.Data["user"].(*entity.User)

		//if len(ctx.InlineQuery.Query) < 2 {
		//	ctx.InlineQuery.Answer(bot, nil, nil)
		//	return nil
		//}

		driveFiles, _, err := driveFileService.FindSomeByFullTextAndFolderID(ctx.InlineQuery.Query, []string{user.Band.DriveFolderID, user.Band.ArchiveFolderID}, "")
		if err != nil {
			return err
		}

		var driveFileIDs []string
		for _, file := range driveFiles {
			driveFileIDs = append(driveFileIDs, file.Id)
		}

		songs, err := songService.FindManyByDriveFileIDs(driveFileIDs)
		if err != nil {
			return err
		}

		var results []gotgbot.InlineQueryResult
		for _, song := range songs {
			if song.PDF.TgFileID == "" {
				continue
			}
			result := gotgbot.InlineQueryResultCachedDocument{
				Id:             song.ID.Hex(),
				Title:          song.PDF.Name,
				DocumentFileId: song.PDF.TgFileID,
				ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: keyboard.SongInitIQ(song, user, ctx.EffectiveUser.LanguageCode),
				},
			}
			results = append(results, result)
		}

		_, err = ctx.InlineQuery.Answer(bot, results, &gotgbot.AnswerInlineQueryOpts{
			CacheTime: 1,
		})
		if err != nil {
			return err
		}

		return nil
	}), 1)

	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.Audio, botController.TransposeAudio_AskForSemitonesNumber), 1)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.Voice, botController.TransposeAudio_AskForSemitonesNumber), 1)

	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, botController.ChooseHandlerOrSearch), 1)

	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, botController.UpdateUser), 2)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(callbackquery.Prefix(fmt.Sprintf("%d:", state.SettingsChooseBand)), botController.UpdateUser), 2)

	go botController.NotifyUsers(bot)

	router := gin.New()
	router.SetFuncMap(template.FuncMap{
		"hex": func(id primitive.ObjectID) string {
			return id.Hex()
		},
		"json": func(s interface{}) string {
			jsonBytes, err := json.Marshal(s)
			if err != nil {
				return ""
			}
			return string(jsonBytes)
		},
		"translate": txt.Get,
	})

	router.LoadHTMLGlob("./webapp/templates/*.go.html")
	router.Static("/webapp/assets", "./webapp/assets")

	router.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/web-app/statistics", webAppController.Statistics)

	router.GET("/web-app/events/create", webAppController.CreateEvent)

	router.GET("/web-app/events/:id/edit", webAppController.EditEvent)
	router.POST("/web-app/events/:id/edit/confirm", webAppController.EditEventConfirm)

	router.GET("/api/drive-files/search", driveFileController.Search)
	router.GET("/api/songs/find-by-drive-file-id", driveFileController.FindByDriveFileID)

	router.GET("/api/users-with-events", webAppController.UsersWithEvents)

	router.GET("/api/songs/:id", webAppController.SongData)
	router.GET("/api/songs/:id/lyrics", webAppController.SongLyrics)
	router.POST("/api/songs/:id/edit", webAppController.SongEdit)
	router.GET("/api/songs/:id/download", webAppController.SongDownload)

	router.GET("/api/tags", webAppController.Tags)

	// Check if we're in development mode
	if os.Getenv("ENV") == "dev" {
		// Create a reverse proxy to the Vite dev server
		proxy, err := createProxy("http://localhost:5173")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create reverse proxy")
		}

		// Forward requests with the /webapp-react prefix to the Vite dev server
		router.Any("/webapp-react/*path", func(c *gin.Context) {
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	} else {
		// In production, serve the built static files (keep your current path)
		router.Static("/webapp-react", "./webapp-react/dist")
	}

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/webapp-react") {
			// Different paths for dev and prod
			if os.Getenv("ENV") == "dev" {
				// This is handled by the proxy in development
				c.Next()
			} else {
				c.File("./webapp-react/dist/index.html")
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "Not Found"})
		}
	})

	go func() {
		// Start receiving updates.
		err = updater.StartPolling(bot, &ext.PollingOpts{
			GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
				Timeout: 10,
				RequestOpts: &gotgbot.RequestOpts{
					Timeout: 11 * time.Second,
				},
			},
		})
		if err != nil {
			panic("failed to start polling: " + err.Error())
		}
		fmt.Printf("%s has been started...\n", bot.User.Username)

		// Idle, to keep updates coming in, and avoid bot stopping.
		updater.Idle()
	}()

	err = router.Run()
	if err != nil {
		panic("error starting Gin: " + err.Error())
	}
}

func createProxy(target string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(url), nil
}
