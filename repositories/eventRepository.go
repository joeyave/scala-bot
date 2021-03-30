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

type EventRepository struct {
	mongoClient *mongo.Client
}

func NewEventRepository(mongoClient *mongo.Client) *EventRepository {
	return &EventRepository{
		mongoClient: mongoClient,
	}
}

func (r *EventRepository) FindAll() ([]*entities.Event, error) {
	events, err := r.find(bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) FindMultipleByBandID(bandID primitive.ObjectID) ([]*entities.Event, error) {
	return r.find(bson.M{
		"bandId": bandID,
	})
}

func (r *EventRepository) FindOneByID(ID primitive.ObjectID) (*entities.Event, error) {
	events, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return events[0], nil
}

func (r *EventRepository) find(m bson.M) ([]*entities.Event, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	pipeline := bson.A{
		bson.M{
			"$match": m,
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
		bson.M{
			"$lookup": bson.M{
				"from": "memberships",
				"let":  bson.M{"eventId": "$_id"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$eventId", "$$eventId"}}},
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
				},
				"as": "memberships",
			},
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var events []*entities.Event
	err = cur.All(context.TODO(), &events)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return events, nil
}

func (r *EventRepository) UpdateOne(event entities.Event) (*entities.Event, error) {
	if event.ID.IsZero() {
		event.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	filter := bson.M{"_id": event.ID}

	event.Memberships = nil
	event.Band = nil
	update := bson.M{
		"$set": event,
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

	var newEvent *entities.Event
	err := result.Decode(&newEvent)
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newEvent.ID)
}

func (r *EventRepository) generateUniqueID() primitive.ObjectID {
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
