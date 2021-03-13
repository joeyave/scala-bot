package repositories

import (
	"context"
	"errors"
	"github.com/joeyave/scala-chords-bot/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type UserRepository struct {
	mongoClient *mongo.Client
}

func NewUserRepository(mongoClient *mongo.Client) *UserRepository {
	return &UserRepository{
		mongoClient: mongoClient,
	}
}

func (r *UserRepository) FindAll() ([]*entities.User, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var users []*entities.User
	err = cursor.All(context.TODO(), &users)
	return users, err
}

func (r *UserRepository) FindOneByID(ID int64) (*entities.User, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("users")

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": ID,
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

	if cur.Next(context.TODO()) == false {
		return nil, errors.New("user not found")
	}

	var user *entities.User
	err = cur.Decode(&user)
	if err != nil {
		return nil, err
	}

	var band *entities.Band
	err = cur.Current.Lookup("band").Unmarshal(&band)
	if err == nil {
		user.Band = band
	}

	return user, nil
}

func (r *UserRepository) FindMultipleByIDs(IDs []int64) ([]*entities.User, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("users")

	filter := bson.M{
		"_id": bson.M{
			"$in": IDs,
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var users []*entities.User
	err = cursor.All(context.TODO(), &users)
	return users, err
}

func (r *UserRepository) UpdateOne(user entities.User) (*entities.User, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("users")

	// TODO: check for ID.

	filter := bson.M{"_id": user.ID}

	update := bson.M{
		"$set": user,
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

	var newUser *entities.User
	err := result.Decode(&newUser)
	return newUser, err
}

func (r *UserRepository) UpdateMultiple(users []entities.User) ([]*entities.User, error) {
	var newUsers []*entities.User

	for _, user := range users {
		newUser, err := r.UpdateOne(user)
		if err != nil {
			return nil, err
		}

		newUsers = append(newUsers, newUser)
	}

	return newUsers, nil
}
