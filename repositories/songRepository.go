package repositories

import (
	"context"
	"github.com/joeyave/scala-chords-bot/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type SongRepository struct {
	mongoClient *mongo.Client
}

func NewSongRepository(mongoClient *mongo.Client) *SongRepository {
	return &SongRepository{
		mongoClient: mongoClient,
	}
}

func (r *SongRepository) FindAll() ([]*entities.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var songs []*entities.Song
	err = cursor.All(context.TODO(), &songs)
	return songs, err
}

func (r *SongRepository) FindOneByID(ID string) (*entities.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")
	result := collection.FindOne(context.TODO(), bson.M{"_id": ID})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var song *entities.Song
	err := result.Decode(&song)
	return song, err
}

func (r *SongRepository) FindMultipleByIDs(IDs []string) ([]*entities.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	filter := bson.M{
		"_id": bson.M{
			"$in": IDs,
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var songs []*entities.Song
	err = cursor.All(context.TODO(), &songs)
	return songs, err
}

func (r *SongRepository) UpdateOne(song entities.Song) (*entities.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	filter := bson.M{"_id": song.ID}

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
	return newSong, err
}

func (r *SongRepository) UpdateMultiple(songs []entities.Song) ([]*entities.Song, error) {
	var newSongs []*entities.Song

	for _, song := range songs {
		newSong, err := r.UpdateOne(song)
		if err != nil {
			return nil, err
		}

		newSongs = append(newSongs, newSong)
	}

	return newSongs, nil
}

func (r *SongRepository) DeleteOneByID(ID string) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	filter := bson.M{"_id": ID}
	_, err := collection.DeleteOne(context.TODO(), filter)
	return err
}
