package repositories

import (
	"context"
	"github.com/joeyave/scala-chords-bot/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *BandRepository) FindAll() ([]*entities.Band, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("bands")
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var bands []*entities.Band
	err = cursor.All(context.TODO(), &bands)
	return bands, err
}

func (r *BandRepository) FindOneByID(ID primitive.ObjectID) (*entities.Band, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("bands")
	result := collection.FindOne(context.TODO(), bson.M{"_id": ID})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var band *entities.Band
	err := result.Decode(&band)
	return band, err
}

func (r *BandRepository) UpdateOne(band entities.Band) (*entities.Band, error) {
	if band.ID.IsZero() {
		band.ID = r.generateUniqueID()
	}

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
		return nil, result.Err()
	}

	var newBand *entities.Band
	err := result.Decode(&newBand)
	return newBand, err
}

func (r *BandRepository) generateUniqueID() primitive.ObjectID {
	ID := primitive.NilObjectID

	for ID.IsZero() {
		ID = primitive.NewObjectID()
		_, err := r.FindOneByID(ID)
		if err == nil {
			ID = primitive.NilObjectID
		}
	}

	return ID
}
