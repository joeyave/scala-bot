package main

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-chords-bot/handlers"
	"github.com/joeyave/scala-chords-bot/repositories"
	"github.com/joeyave/scala-chords-bot/services"
	"github.com/joeyave/telebot/v3"
	"github.com/kjk/notionapi"
	"github.com/klauspost/lctime"
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
	err := lctime.SetLocale("ru_RU")
	if err != nil {
		fmt.Println(err)
	}

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

	driveRepository, err := drive.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	docsRepository, err := docs.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	notionClient := &notionapi.Client{}

	voiceRepository := repositories.NewVoiceRepository(mongoClient)
	voiceService := services.NewVoiceService(voiceRepository)

	bandRepository := repositories.NewBandRepository(mongoClient)
	bandService := services.NewBandService(bandRepository, notionClient)

	driveFileService := services.NewDriveFileService(driveRepository, docsRepository)

	songRepository := repositories.NewSongRepository(mongoClient)
	songService := services.NewSongService(songRepository, voiceRepository, bandRepository, driveRepository, notionClient, driveFileService)

	userRepository := repositories.NewUserRepository(mongoClient)
	userService := services.NewUserService(userRepository)

	membershipRepository := repositories.NewMembershipRepository(mongoClient)
	membershipService := services.NewMembershipService(membershipRepository)

	eventRepository := repositories.NewEventRepository(mongoClient)
	eventService := services.NewEventService(eventRepository, userRepository, membershipRepository, driveRepository, driveFileService)

	roleRepository := repositories.NewRoleRepository(mongoClient)
	roleService := services.NewRoleService(roleRepository)

	//songs, _ := songService.FindAll()
	//for _, song := range songs {
	//	t, _ := time.Parse("2006", "2006")
	//	song.PDF.ModifiedTime = t.Format(time.RFC3339)
	//	songService.UpdateOne(*song)
	//}

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
	bot.Handle(telebot.OnCallback, handler.OnCallback)

	go func() {
		for range time.Tick(time.Hour * 2) {
			events, err := eventService.FindAllFromToday()
			if err != nil {
				return
			}

			for _, event := range events {
				if event.Time.Add(time.Hour*8).Sub(time.Now()).Hours() < 48 {
					for _, membership := range event.Memberships {
						if membership.Notified == true {
							continue
						}

						eventString, _, _ := eventService.ToHtmlStringByID(event.ID)
						_, err := bot.Send(telebot.ChatID(membership.UserID),
							fmt.Sprintf("Привет. Ты учавствуешь в собрании через несколько дней! "+
								"Вот план:\n\n%s", eventString), telebot.ModeHTML, telebot.NoPreview)
						if err != nil {
							continue
						}

						membership.Notified = true
						membershipService.UpdateOne(*membership)
					}
				}
			}
		}
	}()

	bot.Start()
}
