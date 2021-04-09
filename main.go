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
	songService := services.NewSongService(songRepository, voiceRepository, bandRepository, driveRepository, notionClient)

	userRepository := repositories.NewUserRepository(mongoClient)
	userService := services.NewUserService(userRepository)

	membershipRepository := repositories.NewMembershipRepository(mongoClient)
	membershipService := services.NewMembershipService(membershipRepository)

	eventRepository := repositories.NewEventRepository(mongoClient)
	eventService := services.NewEventService(eventRepository, userRepository, membershipRepository, driveRepository, driveFileService)

	roleRepository := repositories.NewRoleRepository(mongoClient)
	roleService := services.NewRoleService(roleRepository)

	//bid, err := primitive.ObjectIDFromHex("6037711ebabcba41d446e401")
	//song1, _ := primitive.ObjectIDFromHex("6061c8328750643343a6d434")
	//song2, _ := primitive.ObjectIDFromHex("6061dd699865cdb0a44092f4")
	//song3, _ := primitive.ObjectIDFromHex("6061e5d26af2e8bfbfb5b1b8")
	//song4, _ := primitive.ObjectIDFromHex("6062fc4bb222aa2a5fb89498")
	//song5, _ := primitive.ObjectIDFromHex("606f79a7fe43159e21b7c483")
	//song6, _ := primitive.ObjectIDFromHex("606f79b5fe43159e21b7c486")
	//
	//roleID, _ := primitive.ObjectIDFromHex("606310e4f216783ac62ce66d")
	//for i := 0; i < 1000; i++ {
	//	event, err := eventService.UpdateOne(entities.Event{
	//		Time:    time.Now(),
	//		Name:    string(rune(i)),
	//		BandID:  bid,
	//		SongIDs: []primitive.ObjectID{song1, song2, song3, song4, song5, song6},
	//	})
	//	if err != nil {
	//		continue
	//	}
	//
	//	membershipService.UpdateOne(entities.Membership{
	//		EventID: event.ID,
	//		UserID:  195295372,
	//		RoleID:  roleID,
	//	})
	//
	//	for i := 0; i < 5; i++ {
	//		membershipService.UpdateOne(entities.Membership{
	//			EventID: event.ID,
	//			UserID:  int64(i),
	//			RoleID:  roleID,
	//		})
	//	}
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
