package repository

import (
	"context"
	"os"

	"github.com/joeyave/scala-bot/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type JoinRequestRepository struct {
	mongoClient *mongo.Client
}

func NewJoinRequestRepository(mongoClient *mongo.Client) *JoinRequestRepository {
	return &JoinRequestRepository{
		mongoClient: mongoClient,
	}
}

func (r *JoinRequestRepository) FindOneByID(ID bson.ObjectID) (*entity.JoinRequest, error) {
	requests, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}
	return requests[0], nil
}

func (r *JoinRequestRepository) FindPendingByUserID(userID int64) ([]*entity.JoinRequest, error) {
	return r.find(bson.M{
		"userId": userID,
		"status": entity.JoinRequestPending,
	})
}

func (r *JoinRequestRepository) FindPendingByUserIDAndBandID(userID int64, bandID bson.ObjectID) (*entity.JoinRequest, error) {
	requests, err := r.find(bson.M{
		"userId": userID,
		"bandId": bandID,
		"status": entity.JoinRequestPending,
	})
	if err != nil {
		return nil, err
	}
	return requests[0], nil
}

func (r *JoinRequestRepository) UpdateOne(request entity.JoinRequest) (*entity.JoinRequest, error) {
	if request.ID.IsZero() {
		request.ID = bson.NewObjectID()
	}

	collection := r.collection()
	filter := bson.M{"_id": request.ID}
	update := bson.M{"$set": request}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var newRequest *entity.JoinRequest
	if err := result.Decode(&newRequest); err != nil {
		return nil, err
	}

	return r.FindOneByID(newRequest.ID)
}

func (r *JoinRequestRepository) find(m bson.M) ([]*entity.JoinRequest, error) {
	collection := r.collection()

	cursor, err := collection.Find(context.TODO(), m, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		return nil, err
	}

	var requests []*entity.JoinRequest
	if err := cursor.All(context.TODO(), &requests); err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return nil, ErrNotFound
	}

	return requests, nil
}

func (r *JoinRequestRepository) collection() *mongo.Collection {
	return r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("join_requests")
}
