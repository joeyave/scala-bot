package repository

import (
	"context"
	"os"

	"github.com/joeyave/scala-bot/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BandRepository struct {
	mongoClient *mongo.Client
}

func NewBandRepository(mongoClient *mongo.Client) *BandRepository {
	return &BandRepository{
		mongoClient: mongoClient,
	}
}

func (r *BandRepository) FindAll() ([]*entity.Band, error) {
	bands, err := r.find(bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		return nil, err
	}

	return bands, nil
}

func (r *BandRepository) FindManyByIDs(ids []bson.ObjectID) ([]*entity.Band, error) {
	if len(ids) == 0 {
		return []*entity.Band{}, nil
	}
	bands, err := r.find(bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}

	return bands, nil
}

func (r *BandRepository) FindOneByID(ID bson.ObjectID) (*entity.Band, error) {
	bands, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return bands[0], nil
}

func (r *BandRepository) FindOneByDriveFolderID(driveFolderID string) (*entity.Band, error) {
	bands, err := r.find(bson.M{
		"$or": []bson.M{
			{"driveFolderId": driveFolderID},
			{"archiveFolderId": driveFolderID},
		},
	})
	if err != nil {
		return nil, err
	}

	return bands[0], nil
}

func (r *BandRepository) find(m bson.M) ([]*entity.Band, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("bands")

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

	var bands []*entity.Band
	err = cur.All(context.TODO(), &bands)
	if err != nil {
		return nil, err
	}

	if len(bands) == 0 {
		return nil, ErrNotFound
	}

	return bands, nil
}

func (r *BandRepository) UpdateOne(band entity.Band) (*entity.Band, error) {
	if band.ID.IsZero() {
		band.ID = bson.NewObjectID()
	}

	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("bands")

	filter := bson.M{"_id": band.ID}

	band.Roles = nil
	update := bson.M{
		"$set": band,
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var newBand *entity.Band
	err := result.Decode(&newBand)
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newBand.ID)
}
