package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"log"
	"os"
	"scala-chords-bot/handlers"
	"scala-chords-bot/repositories"
	"scala-chords-bot/services"
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
	songRepository := repositories.NewSongRepository(mongoClient)
	songService := services.NewSongService(songRepository, driveClient, docsClient)

	userRepository := repositories.NewUserRepository(mongoClient)
	userService := services.NewUserService(userRepository)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	handler := handlers.NewHandler(bot, userService, songService)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	lastOffset := 0
	u := tgbotapi.NewUpdate(lastOffset + 1)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// TODO: find out how to recover from panic.
	for update := range updates {
		lastOffset = update.UpdateID
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// TODO: make some handler struct and all that stuff.
		err = handler.HandleUpdate(&update)
	}
}
