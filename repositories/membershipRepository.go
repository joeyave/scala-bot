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

type MembershipRepository struct {
	mongoClient *mongo.Client
}

func NewMembershipRepository(mongoClient *mongo.Client) *MembershipRepository {
	return &MembershipRepository{
		mongoClient: mongoClient,
	}
}

func (r *MembershipRepository) FindAll() ([]*entities.Membership, error) {
	memberships, err := r.find(bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *MembershipRepository) FindOneByID(ID primitive.ObjectID) (*entities.Membership, error) {
	memberships, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return memberships[0], nil
}

func (r *MembershipRepository) FindMultipleByUserIDAndEventID(userID int64, eventID primitive.ObjectID) ([]*entities.Membership, error) {
	memberships, err := r.find(bson.M{"userId": userID, "eventId": eventID})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *MembershipRepository) find(m bson.M) ([]*entities.Membership, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("memberships")

	pipeline := bson.A{
		bson.M{
			"$match": m,
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var memberships []*entities.Membership
	err = cur.All(context.TODO(), &memberships)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return memberships, nil
}

func (r *MembershipRepository) UpdateOne(membership entities.Membership) (*entities.Membership, error) {
	if membership.ID.IsZero() {
		membership.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("memberships")

	filter := bson.M{"_id": membership.ID}

	membership.Role = nil
	update := bson.M{
		"$set": membership,
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

	var newMembership *entities.Membership
	err := result.Decode(&newMembership)
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newMembership.ID)
}

func (r *MembershipRepository) DeleteOneByID(ID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("memberships")

	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": ID})
	return err
}

func (r *MembershipRepository) DeleteManyByEventID(eventID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("memberships")

	_, err := collection.DeleteMany(context.TODO(), bson.M{"eventId": eventID})
	return err
}

func (r *MembershipRepository) generateUniqueID() primitive.ObjectID {
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
