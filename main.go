package main

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-chords-bot/handlers"
	"github.com/joeyave/scala-chords-bot/repositories"
	"github.com/joeyave/scala-chords-bot/services"
	"github.com/joeyave/telebot/v3"
	"github.com/kjk/notionapi"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

func main() {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(ctx)
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	driveClient, err := drive.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	docsClient, err := docs.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	notionClient := &notionapi.Client{}

	voiceRepository := repositories.NewVoiceRepository(mongoClient)
	voiceService := services.NewVoiceService(voiceRepository)

	bandRepository := repositories.NewBandRepository(mongoClient)
	bandService := services.NewBandService(bandRepository, notionClient)

	driveFileService := services.NewDriveFileService(driveClient, docsClient)

	songRepository := repositories.NewSongRepository(mongoClient, driveClient)
	songService := services.NewSongService(songRepository, voiceRepository, bandRepository, driveClient, notionClient)

	userRepository := repositories.NewUserRepository(mongoClient)
	userService := services.NewUserService(userRepository)

	membershipRepository := repositories.NewMembershipRepository(mongoClient)
	membershipService := services.NewMembershipService(membershipRepository)

	eventRepository := repositories.NewEventRepository(mongoClient)
	eventService := services.NewEventService(eventRepository, userRepository)

	roleRepository := repositories.NewRoleRepository(mongoClient)
	roleService := services.NewRoleService(roleRepository)

	bot, err := telebot.NewBot(telebot.Settings{
		Token:       os.Getenv("BOT_TOKEN"),
		Poller:      &telebot.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
	})
	if err != nil {
		log.Fatal(err)
	}

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

	go func() {
		for range time.Tick(time.Hour * 18) {
			events, err := eventService.FindAllFromToday()
			if err != nil {
				return
			}

			for _, event := range events {
				if event.Time.Add(time.Hour*8).Sub(time.Now()).Hours() < 48 {
					for _, membership := range event.Memberships {
						eventString, _ := eventService.ToHtmlStringByID(event.ID)
						bot.Send(telebot.ChatID(membership.UserID), fmt.Sprintf(
							"Привет. Ты учавствуешь в собрании через несколько дней! Вот план:\n\n%s",
							eventString), telebot.ModeHTML)
					}
				}
			}
		}
	}()

	bot.Start()
}
