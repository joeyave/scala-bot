package repositories

import (
	"context"
	"github.com/joeyave/scala-chords-bot/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type BandRepository struct {
	mongoClient *mongo.Client
}

func NewBandRepository(mongoClient *mongo.Client) *BandRepository {
	return &BandRepository{
		mongoClient: mongoClient,
	}
}

func (r *BandRepository) FindAll() ([]entities.Band, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("bands")
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var bands []entities.Band
	err = cursor.All(context.TODO(), &bands)
	return bands, err
}

func (r *BandRepository) UpdateOne(band entities.Band) (entities.Band, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("bands")

	filter := bson.M{"_id": band.ID}

	update := bson.M{
		"$set": band,
	}

	after := options.After
	upsert := true
	opts := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, &opts)
	if result.Err() != nil {
		return band, result.Err()
	}

	var newBand = entities.Band{}
	err := result.Decode(&newBand)
	return newBand, err
}
