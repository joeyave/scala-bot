package main

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-bot/handlers"
	"github.com/joeyave/scala-bot/repositories"
	"github.com/joeyave/scala-bot/services"
	"github.com/klauspost/lctime"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gopkg.in/telebot.v3"
	"os"
	"time"
)

func main() {

	bot, err := telebot.NewBot(telebot.Settings{
		Token:       os.Getenv("BOT_TOKEN"),
		Poller:      &telebot.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
	})
	if err != nil {
		log.Fatal().Err(err)
	}

	//w := helpers.LogsWriter{
	//	Bot:       bot,
	//	ChannelID: helpers.LogsChannelID,
	//}
	out := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    true,
	}
	log.Logger = zerolog.New(out).Level(zerolog.GlobalLevel()).With().Timestamp().Logger()

	err = lctime.SetLocale("ru_RU")
	if err != nil {
		fmt.Println(err)
	}

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal().Err(err)
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
		log.Fatal().Err(err)
	}

	driveRepository, err := drive.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatal().Msgf("Unable to retrieve Drive client: %v", err)
	}

	docsRepository, err := docs.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatal().Msgf("Unable to retrieve Docs client: %v", err)
	}

	voiceRepository := repositories.NewVoiceRepository(mongoClient)
	voiceService := services.NewVoiceService(voiceRepository)

	bandRepository := repositories.NewBandRepository(mongoClient)
	bandService := services.NewBandService(bandRepository)

	driveFileService := services.NewDriveFileService(driveRepository, docsRepository)

	songRepository := repositories.NewSongRepository(mongoClient)
	songService := services.NewSongService(songRepository, voiceRepository, bandRepository, driveRepository, driveFileService)

	userRepository := repositories.NewUserRepository(mongoClient)
	userService := services.NewUserService(userRepository)

	membershipRepository := repositories.NewMembershipRepository(mongoClient)
	membershipService := services.NewMembershipService(membershipRepository)

	eventRepository := repositories.NewEventRepository(mongoClient)
	eventService := services.NewEventService(eventRepository, userRepository, membershipRepository, driveRepository, driveFileService)

	roleRepository := repositories.NewRoleRepository(mongoClient)
	roleService := services.NewRoleService(roleRepository)

	handler := handlers.NewHandler(
		bot,
		userService,
		driveFileService,
		songService,
		voiceService,
		bandService,
		membershipService,
		eventService,
		roleService,
	)

	bot.OnError = handler.OnError

	bot.Use(handler.RegisterUserMiddleware)

	bot.Handle(telebot.OnText, handler.OnText)
	bot.Handle(telebot.OnVoice, handler.OnVoice)
	bot.Handle(telebot.OnAudio, handler.OnVoice)
	bot.Handle(telebot.OnCallback, handler.OnCallback)

	go handler.NotifyUser()

	bot.Start()
}
