package repository

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-bot/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type RoleRepository struct {
	mongoClient *mongo.Client
}

func NewRoleRepository(mongoClient *mongo.Client) *RoleRepository {
	return &RoleRepository{
		mongoClient: mongoClient,
	}
}

func (r *RoleRepository) FindAll() ([]*entity.Role, error) {
	roles, err := r.find(bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *RoleRepository) FindOneByID(ID primitive.ObjectID) (*entity.Role, error) {
	roles, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return roles[0], nil
}

func (r *RoleRepository) find(m bson.M) ([]*entity.Role, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("roles")

	pipeline := bson.A{
		bson.M{
			"$match": m,
		},
		bson.M{
			"$sort": bson.M{
				"priority": 1,
			},
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var roles []*entity.Role
	err = cur.All(context.TODO(), &roles)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return roles, nil
}

func (r *RoleRepository) UpdateOne(role entity.Role) (*entity.Role, error) {
	if role.ID.IsZero() {
		role.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("roles")

	filter := bson.M{"_id": role.ID}

	update := bson.M{
		"$set": role,
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

	var newRole *entity.Role
	err := result.Decode(&newRole)
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newRole.ID)
}

func (r *RoleRepository) generateUniqueID() primitive.ObjectID {
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
