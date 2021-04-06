package repositories

import (
	"context"
	"fmt"
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
	bands, err := r.find(bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		return nil, err
	}

	return bands, nil
}

func (r *BandRepository) FindOneByID(ID primitive.ObjectID) (*entities.Band, error) {
	bands, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return bands[0], nil
}

func (r *BandRepository) FindOneByDriveFolderID(driveFolderID string) (*entities.Band, error) {
	bands, err := r.find(bson.M{"driveFolderId": driveFolderID})
	if err != nil {
		return nil, err
	}

	return bands[0], nil
}

func (r *BandRepository) find(m bson.M) ([]*entities.Band, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("bands")

	pipeline := bson.A{
		bson.M{
			"$match": m,
		},
		bson.M{
			"$sort": bson.M{
				"priority": 1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "roles",
				"localField":   "_id",
				"foreignField": "bandId",
				"as":           "roles",
			},
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var bands []*entities.Band
	err = cur.All(context.TODO(), &bands)
	if err != nil {
		return nil, err
	}

	if len(bands) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return bands, nil
}

func (r *BandRepository) UpdateOne(band entities.Band) (*entities.Band, error) {
	if band.ID.IsZero() {
		band.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("bands")

	filter := bson.M{"_id": band.ID}

	band.Roles = nil
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
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newBand.ID)
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
