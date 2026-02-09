package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joeyave/scala-bot/repository"
	"github.com/joeyave/scala-bot/service"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("BOT_MONGODB_URI")))
	if err != nil {
		panic(fmt.Sprintf("failed to connect mongo: %v", err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		panic(fmt.Sprintf("failed to ping mongo: %v", err))
	}

	driveRepository, err := drive.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("BOT_GOOGLEAPIS_KEY"))))
	if err != nil {
		panic(fmt.Sprintf("failed to init drive: %v", err))
	}
	docsRepository, err := docs.NewService(context.TODO(), option.WithCredentialsJSON([]byte(os.Getenv("BOT_GOOGLEAPIS_KEY"))))
	if err != nil {
		panic(fmt.Sprintf("failed to init docs: %v", err))
	}

	driveFileService := service.NewDriveFileService(driveRepository, docsRepository)
	songRepository := repository.NewSongRepository(mongoClient)
	voiceRepository := repository.NewVoiceRepository(mongoClient)
	bandRepository := repository.NewBandRepository(mongoClient)
	songService := service.NewSongService(songRepository, voiceRepository, bandRepository, driveRepository, driveFileService)

	songs, err := songService.FindAll()
	if err != nil {
		panic(fmt.Sprintf("failed to load songs: %v", err))
	}

	processed := 0
	succeeded := 0
	failed := 0

	for _, song := range songs {
		if song == nil || song.DriveFileID == "" {
			continue
		}
		processed++
		if _, err := driveFileService.EnsureBodyMetadataLayout(song.DriveFileID); err != nil {
			failed++
			fmt.Printf("[FAIL] %s %s: %v\n", song.ID.Hex(), song.DriveFileID, err)
			continue
		}
		succeeded++
		fmt.Printf("[OK] %s %s\n", song.ID.Hex(), song.DriveFileID)
	}

	fmt.Printf("Migration finished. processed=%d succeeded=%d failed=%d\n", processed, succeeded, failed)
}
