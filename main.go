package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/controller"
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
	"os"
	"time"
)

func main() {
	out := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	log.Logger = zerolog.New(out).Level(zerolog.GlobalLevel()).With().Timestamp().Logger()

	// Create bot from environment value.
	bot, err := gotgbot.NewBot(os.Getenv("BOT_TOKEN"), &gotgbot.BotOpts{
		Client:      http.Client{},
		GetTimeout:  gotgbot.DefaultGetTimeout,
		PostTimeout: gotgbot.DefaultPostTimeout,
	})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal().Err(err).Msg("error creating mongo client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer mongoClient.Disconnect(ctx)
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal().Err(err).Msg("error pinging mongo")
	}

	driveRepository, err := drive.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatal().Msgf("Unable to retrieve Drive client: %v", err)
	}

	docsRepository, err := docs.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
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
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: nil,
		DispatcherOpts: ext.DispatcherOpts{
			Error:       botController.Error,
			Panic:       nil, // todo
			ErrorLog:    nil,
			MaxRoutines: 0,
		},
	})
	dispatcher := updater.Dispatcher

	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.All, botController.RegisterUser), 0)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(callbackquery.All, botController.RegisterUser), 0)

	// Plain keyboard.
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
		ctx.EffectiveChat.SendMessage(bot, "Статистика временно не доступна.", nil)
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

	// Inline keyboard.
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.BandCreate_AskForName), botController.BandCreate_AskForName), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.RoleCreate_AskForName), botController.RoleCreate_AskForName), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.RoleCreate), botController.RoleCreate), 1)

	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsCB), botController.SettingsCB), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsBands), botController.SettingsBands), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsChooseBand), botController.SettingsChooseBand), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SettingsBandMembers), botController.SettingsBandMembers), 1)
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
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoices), botController.SongVoices), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoicesCreateVoiceAskForAudio), botController.SongVoicesAddVoiceAskForAudio), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoice), botController.SongVoice), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoiceDeleteConfirm), botController.SongVoiceDeleteConfirm), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongVoiceDelete), botController.SongVoiceDelete), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongDeleteConfirm), botController.SongDeleteConfirm), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongDelete), botController.SongDelete), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongCopyToMyBand), botController.SongCopyToMyBand), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongStyle), botController.SongStyle), 1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(util.CallbackState(state.SongAddLyricsPage), botController.SongAddLyricsPage), 1)

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
	})

	router.LoadHTMLGlob("webapp/templates/*.go.html")
	router.Static("/webapp/assets", "./webapp/assets")

	router.Use()
	router.GET("/web-app/events/create", webAppController.CreateEvent)
	router.GET("/web-app/songs/create", webAppController.CreateSong)

	router.GET("/web-app/events/:id/edit", webAppController.EditEvent)
	router.POST("/web-app/events/:id/edit/confirm", webAppController.EditEventConfirm)

	router.GET("/web-app/songs/:id/edit", webAppController.EditSong)
	router.POST("/web-app/songs/:id/edit/confirm", webAppController.EditSongConfirm)

	router.GET("/api/drive-files/search", driveFileController.Search)
	router.GET("/api/songs/find-by-drive-file-id", driveFileController.FindByDriveFileID)

	go func() {
		// Start receiving updates.
		err = updater.StartPolling(bot, &ext.PollingOpts{DropPendingUpdates: true})
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
