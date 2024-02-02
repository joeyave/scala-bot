package repository

import (
	"context"
	"github.com/joeyave/scala-bot/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type VoiceRepository struct {
	mongoClient *mongo.Client
}

func NewVoiceRepository(mongoClient *mongo.Client) *VoiceRepository {
	return &VoiceRepository{
		mongoClient: mongoClient,
	}
}

func (r *VoiceRepository) FindOneByID(ID primitive.ObjectID) (*entity.Voice, error) {
	return r.findOne(bson.M{"_id": ID})
}

func (r *VoiceRepository) FindOneByFileID(fileID string) (*entity.Voice, error) {
	return r.findOne(bson.M{"fileId": fileID})
}

func (r *VoiceRepository) UpdateOne(voice entity.Voice) (*entity.Voice, error) {
	if voice.ID.IsZero() {
		voice.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("voices")

	filter := bson.M{
		"_id": voice.ID,
	}

	update := bson.M{
		"$set": voice,
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

	var newVoice *entity.Voice
	err := result.Decode(&newVoice)
	return newVoice, err
}

func (r *VoiceRepository) DeleteOneByID(ID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("voices")

	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": ID})
	return err
}

func (r *VoiceRepository) DeleteManyByIDs(IDs []primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("voices")

	_, err := collection.DeleteMany(context.TODO(), bson.M{"_id": bson.M{"$in": IDs}})
	return err
}

func (r *VoiceRepository) findOne(m bson.M) (*entity.Voice, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("voices")

	result := collection.FindOne(context.TODO(), m)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var voice *entity.Voice
	err := result.Decode(&voice)
	return voice, err
}

func (r *VoiceRepository) generateUniqueID() primitive.ObjectID {
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
