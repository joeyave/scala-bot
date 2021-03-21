package repositories

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/drive/v3"
	"os"
)

type SongRepository struct {
	mongoClient *mongo.Client
	driveClient *drive.Service
}

func NewSongRepository(mongoClient *mongo.Client, driveClient *drive.Service) *SongRepository {
	return &SongRepository{
		mongoClient: mongoClient,
		driveClient: driveClient,
	}
}

func (r *SongRepository) FindAll() ([]*entities.Song, error) {
	return r.find(bson.M{})
}

func (r *SongRepository) FindOneByID(ID primitive.ObjectID) (*entities.Song, error) {
	songs, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}
	return songs[0], nil
}

func (r *SongRepository) FindOneByDriveFileID(driveFileID string) (*entities.Song, error) {
	songs, err := r.find(bson.M{"driveFileId": driveFileID})
	if err != nil {
		return nil, err
	}
	return songs[0], nil
}

func (r *SongRepository) find(m bson.M) ([]*entities.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	pipeline := bson.A{
		bson.M{
			"$match": m,
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "voices",
				"localField":   "_id",
				"foreignField": "songId",
				"as":           "voices",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "bands",
				"localField":   "bandId",
				"foreignField": "_id",
				"as":           "band",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$band",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var songs []*entities.Song
	for cur.Next(context.TODO()) {
		var song *entities.Song
		err := cur.Decode(&song)
		if err != nil {
			continue
		}

		driveFile, err := r.driveClient.Files.Get(song.DriveFileID).Fields("id, name, modifiedTime, webViewLink, parents").Do()
		if err != nil {
			continue
		}

		song.DriveFile = driveFile

		songs = append(songs, song)
	}

	if len(songs) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return songs, nil
}

func (r *SongRepository) UpdateOne(song entities.Song) (*entities.Song, error) {
	if song.ID.IsZero() {
		song.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	filter := bson.M{
		"_id": song.ID,
	}

	song.Band = nil
	song.Voices = nil
	update := bson.M{
		"$set": song,
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

	var newSong *entities.Song
	err := result.Decode(&newSong)
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newSong.ID)
}

func (r *SongRepository) DeleteOneByID(ID string) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": ID})
	return err
}

func (r *SongRepository) generateUniqueID() primitive.ObjectID {
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
