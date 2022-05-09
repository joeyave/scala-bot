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

type MembershipRepository struct {
	mongoClient *mongo.Client
}

func NewMembershipRepository(mongoClient *mongo.Client) *MembershipRepository {
	return &MembershipRepository{
		mongoClient: mongoClient,
	}
}

func (r *MembershipRepository) FindAll() ([]*entity.Membership, error) {
	memberships, err := r.find(bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *MembershipRepository) FindOneByID(ID primitive.ObjectID) (*entity.Membership, error) {
	memberships, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return memberships[0], nil
}

func (r *MembershipRepository) FindMultipleByUserIDAndEventID(userID int64, eventID primitive.ObjectID) ([]*entity.Membership, error) {
	memberships, err := r.find(bson.M{"userId": userID, "eventId": eventID})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *MembershipRepository) FindMultipleByUserIDAndEventIDAndRoleID(userID int64, eventID, roleID primitive.ObjectID) ([]*entity.Membership, error) {
	memberships, err := r.find(bson.M{"userId": userID, "eventId": eventID, "roleId": roleID})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *MembershipRepository) FindMultipleByEventID(eventID primitive.ObjectID) ([]*entity.Membership, error) {
	memberships, err := r.find(bson.M{"eventId": eventID})
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *MembershipRepository) find(m bson.M) ([]*entity.Membership, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("memberships")

	pipeline := bson.A{
		bson.M{
			"$match": m,
		},
		bson.M{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"userId": "$userId"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$userId"}}},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "bands",
							"let":  bson.M{"bandId": "$bandId"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$bandId"}}},
								},
								bson.M{
									"$lookup": bson.M{
										"from": "roles",
										"let":  bson.M{"bandId": "$_id"},
										"pipeline": bson.A{
											bson.M{
												"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$bandId", "$$bandId"}}},
											},
											bson.M{
												"$sort": bson.M{
													"priority": 1,
												},
											},
										},
										"as": "roles",
									},
								},
							},
							"as": "band",
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "$band",
							"preserveNullAndEmptyArrays": true,
						},
					},
				},
				"as": "user",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$user",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "roles",
				"let":  bson.M{"roleId": "$roleId"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$roleId"}}},
					},
				},
				"as": "role",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$role",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$sort": bson.D{
				{"role._id", 1},
				{"role.priority", 1},
			},
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var memberships []*entity.Membership
	err = cur.All(context.TODO(), &memberships)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return memberships, nil
}

func (r *MembershipRepository) UpdateOne(membership entity.Membership) (*entity.Membership, error) {
	if membership.ID.IsZero() {
		membership.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("memberships")

	filter := bson.M{"_id": membership.ID}

	membership.User = nil
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

	var newMembership *entity.Membership
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
