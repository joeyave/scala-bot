package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-chords-bot/handlers"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/scala-chords-bot/repositories"
	"github.com/joeyave/scala-chords-bot/services"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"github.com/kjk/notionapi"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
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

	songRepository := repositories.NewSongRepository(mongoClient, driveClient)
	driveFileService := services.NewDriveFileService(driveClient, docsClient)
	songService := services.NewSongService(songRepository, voiceRepository, bandRepository, driveClient, notionClient)

	userRepository := repositories.NewUserRepository(mongoClient)
	userService := services.NewUserService(userRepository)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	handler := handlers.NewHandler(bot, userService, driveFileService, songService, voiceService, bandService)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/googlef424063251e2d68b", func(c *gin.Context) {
		c.HTML(http.StatusOK, "googlef424063251e2d68b.html", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {
		defer c.Request.Body.Close()

		bytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			helpers.LogError(nil, bot, err)
			return
		}

		var update tgbotapi.Update
		err = json.Unmarshal(bytes, &update)
		if err != nil {
			helpers.LogError(nil, bot, err)
			return
		}

		err = handler.HandleUpdate(&update)
		if err != nil {
			helpers.LogError(&update, bot, err)
			return
		}
	})

	r.Run(":" + os.Getenv("PORT"))

	//log.Printf("Authorized on account %s", bot.Self.UserName)
	//
	//lastOffset := 0
	//u := tgbotapi.NewUpdate(lastOffset + 1)
	//u.Timeout = 60
	//
	//updates := bot.GetUpdatesChan(u)
	//
	//go func() {
	//	for update := range updates {
	//		lastOffset = update.UpdateID
	//		if update.Message == nil {
	//			continue
	//		}
	//
	//		go func(update tgbotapi.Update) {
	//			err := handler.HandleUpdate(&update)
	//			if err != nil {
	//				helpers.LogError(&update, bot, err)
	//			}
	//		}(update)
	//	}
	//	wg.Done()
	//}()
}
