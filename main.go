package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"log"
	"os"
	"scalaChordsBot/handlers"
	"scalaChordsBot/repositories"
	"scalaChordsBot/services"
)

func main() {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	err = mongoClient.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	defer mongoClient.Disconnect(context.TODO())

	driveClient, err := drive.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLEAPIS_CREDENTIALS"))))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	songRepository := repositories.NewSongRepository(mongoClient)
	songService := services.NewSongService(songRepository, driveClient)

	userRepository := repositories.NewUserRepository(mongoClient)
	userService := services.NewUserService(userRepository)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	handler := handlers.NewHandler(bot, userService, songService)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// TODO: find out how to recover from panic.
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// TODO: make some handler struct and all that stuff.
		err = handler.Handle(&update)
	}
}
