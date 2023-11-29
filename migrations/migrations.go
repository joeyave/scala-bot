package migrations

import (
	"context"
	"github.com/joeyave/scala-bot/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"
)

func MigrateLikes(client *mongo.Client) error {
	collection := client.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

	// 1. Fetch all documents from the collection.
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.TODO())

	// 2. Iterate through each document and update the "Likes" field.
	for cursor.Next(context.TODO()) {
		var song entity.OldSong
		if err := cursor.Decode(&song); err != nil {
			continue
		}

		// Convert old likes ([]int64) to new likes ([]*Like).
		newLikes := make([]*entity.Like, len(song.Likes))
		for i, userID := range song.Likes {
			newLikes[i] = &entity.Like{
				UserID: userID,
				Time:   time.Now(),
			}
		}

		// Update the document in the collection.
		_, err := collection.UpdateOne(
			context.TODO(),
			bson.M{"_id": song.ID},
			bson.M{"$set": bson.M{"likes": newLikes}},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
