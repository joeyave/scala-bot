package repository

import (
	"context"
	"os"
	"time"

	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/helpers"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type EventRepository struct {
	mongoClient *mongo.Client
}

func NewEventRepository(mongoClient *mongo.Client) *EventRepository {
	return &EventRepository{
		mongoClient: mongoClient,
	}
}

func (r *EventRepository) FindAll() ([]*entity.Event, error) {
	events, err := r.find(
		bson.M{
			"_id": bson.M{
				"$ne": "",
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) FindOneOldestByBandID(bandID bson.ObjectID) (*entity.Event, error) {
	event, err := r.find(
		bson.M{
			"bandId": bandID,
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
		bson.M{
			"$limit": 1,
		},
	)
	if err != nil {
		return nil, err
	}

	return event[0], err
}

func (r *EventRepository) FindManyFromDateByBandID(bandID bson.ObjectID, fromUTC time.Time) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"bandId": bandID,
			"time": bson.M{
				"$gte": fromUTC,
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
	)
}

func (r *EventRepository) FindBetweenDates(fromUTC, toUTC time.Time) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"time": bson.M{
				"$gte": fromUTC,
				"$lte": toUTC,
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
	)
}

func (r *EventRepository) FindManyBetweenDatesByBandID(fromUTC, toUTC time.Time, bandID bson.ObjectID) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"bandId": bandID,
			"time": bson.M{
				"$gte": fromUTC,
				"$lte": toUTC,
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": -1,
			},
		},
	)
}

func (r *EventRepository) FindManyByBandIDAndPageNumber(bandID bson.ObjectID, pageNumber int) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"bandId": bandID,
		},
		bson.M{
			"$sort": bson.M{
				"time": -1,
			},
		},
		bson.M{
			"$skip": pageNumber * helpers.EventsPageSize,
		},
		bson.M{
			"$limit": helpers.EventsPageSize,
		},
	)
}

func (r *EventRepository) FindManyUntilByBandIDAndPageNumber(bandID bson.ObjectID, untilUTC time.Time, pageNumber int) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"bandId": bandID,
			"time": bson.M{
				"$lt": untilUTC,
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": -1,
			},
		},
		bson.M{
			"$skip": pageNumber * helpers.EventsPageSize,
		},
		bson.M{
			"$limit": helpers.EventsPageSize,
		},
	)
}

func (r *EventRepository) FindManyUntilByBandIDAndWeekdayAndPageNumber(bandID bson.ObjectID, untilUTC time.Time, weekday time.Weekday, pageNumber int) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"bandId": bandID,
			"time": bson.M{
				"$lt": untilUTC,
			},
			"dayOfWeek": int(weekday) + 1,
		},
		bson.M{
			"$sort": bson.M{
				"time": -1,
			},
		},
		bson.M{
			"$skip": pageNumber * helpers.EventsPageSize,
		},
		bson.M{
			"$limit": helpers.EventsPageSize,
		},
	)
}

func (r *EventRepository) FindManyUntilByBandIDAndUserIDAndPageNumber(bandID bson.ObjectID, userID int64, untilUTC time.Time, pageNumber int) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"bandId":             bandID,
			"memberships.userId": userID,
			"time": bson.M{
				"$lt": untilUTC,
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": -1,
			},
		},
		bson.M{
			"$skip": pageNumber * helpers.EventsPageSize,
		},
		bson.M{
			"$limit": helpers.EventsPageSize,
		},
	)
}

func (r *EventRepository) FindManyFromTodayByBandIDAndUserID(bandID bson.ObjectID, userID int64, fromUTC time.Time, pageNumber int) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"bandId":             bandID,
			"memberships.userId": userID,
			"time": bson.M{
				"$gte": fromUTC,
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
		bson.M{
			"$skip": pageNumber * helpers.EventsPageSize,
		},
		bson.M{
			"$limit": helpers.EventsPageSize,
		},
	)
}

func (r *EventRepository) FindMultipleByIDs(IDs []bson.ObjectID) ([]*entity.Event, error) {
	return r.find(
		bson.M{
			"_id": bson.M{
				"$in": IDs,
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
	)
}

func (r *EventRepository) FindOneByID(ID bson.ObjectID) (*entity.Event, error) {
	events, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return events[0], nil
}

func (r *EventRepository) FindOneByNameAndTimeAndBandID(name string, fromUTC, toUTC time.Time, bandID bson.ObjectID) (*entity.Event, error) {
	events, err := r.find(
		bson.M{
			"name": name,
			"time": bson.M{
				"$gte": fromUTC,
				"$lt":  toUTC,
			},
			"bandId": bandID,
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return events[0], nil
}

func (r *EventRepository) GetAlias(ctx context.Context, eventID bson.ObjectID, lang string) (string, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	result := collection.FindOne(ctx, bson.M{"_id": eventID})
	if result.Err() != nil {
		return "", result.Err()
	}

	var event *entity.Event
	err := result.Decode(&event)
	if err != nil {
		return "", err
	}

	return event.Alias(lang), nil
}

func (r *EventRepository) find(m bson.M, opts ...bson.M) ([]*entity.Event, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	pipeline := bson.A{
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
					bson.M{
						"$sort": bson.D{
							{Key: "role._id", Value: 1},
							{Key: "role.priority", Value: 1},
						},
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
				},
				"as": "memberships",
			},
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
		bson.M{
			"$addFields": bson.M{
				"bandTimezone": "$band.timezone",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"dayOfWeek": bson.M{
					"$dayOfWeek": bson.M{
						"date":     "$time",
						"timezone": "$bandTimezone",
					},
				},
			},
		},
		bson.M{
			"$match": m,
		},
		bson.M{
			"$addFields": bson.M{
				"songIds": bson.M{
					"$cond": bson.M{
						"if": bson.M{
							"$ne": bson.A{bson.M{"$type": "$songIds"}, "array"},
						},
						"then": bson.A{},
						"else": "$songIds",
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "songs",
				"let":  bson.M{"songIds": "$songIds"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$songIds"}}},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "bands",
							"let":  bson.M{"bandId": "$bandId"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$bandId"}}},
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
							"from": "voices",
							"let":  bson.M{"songId": "$_id"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$songId", "$$songId"}}},
								},
							},
							"as": "voices",
						},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": bson.M{
								"$indexOfArray": bson.A{"$$songIds", "$_id"},
							},
						},
					},
					bson.M{
						"$sort": bson.M{"sort": 1},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": "$$REMOVE",
						},
					},
				},
				"as": "songs",
			},
		},
	}

	for _, o := range opts {
		pipeline = append(pipeline, o)
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var events []*entity.Event
	err = cur.All(context.TODO(), &events)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return events, nil
}

func (r *EventRepository) UpdateOne(event entity.Event) (*entity.Event, error) {
	if event.ID.IsZero() {
		event.ID = bson.NewObjectID()
	}

	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	filter := bson.M{"_id": event.ID}

	event.Memberships = nil
	event.Band = nil
	event.Songs = nil
	update := bson.M{
		"$set": event,
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var newEvent *entity.Event
	err := result.Decode(&newEvent)
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newEvent.ID)
}

func (r *EventRepository) DeleteOneByID(ID bson.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": ID})
	return err
}

func (r *EventRepository) GetEventWithSongs(eventID bson.ObjectID) (*entity.Event, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": eventID,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"songIds": bson.M{
					"$cond": bson.M{
						"if": bson.M{
							"$ne": bson.A{bson.M{"$type": "$songIds"}, "array"},
						},
						"then": bson.A{},
						"else": "$songIds",
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "songs",
				"let":  bson.M{"songIds": "$songIds"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$songIds"}}},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": bson.M{
								"$indexOfArray": bson.A{"$$songIds", "$_id"},
							},
						},
					},
					bson.M{
						"$sort": bson.M{"sort": 1},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": "$$REMOVE",
						},
					},
				},
				"as": "songs",
			},
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var events []*entity.Event
	err = cur.All(context.TODO(), &events)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return events[0], nil
}

func (r *EventRepository) PushSongID(eventID, songID bson.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	filter := bson.M{
		"_id":     eventID,
		"songIds": bson.M{"$nin": bson.A{songID}},
	}

	update := bson.M{
		"$push": bson.M{
			"songIds": songID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *EventRepository) ChangeSongIDPosition(eventID, songID bson.ObjectID, newPosition int) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	filter := bson.M{
		"_id": eventID,
	}

	update := bson.M{
		"$pull": bson.M{
			"songIds": songID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	filter = bson.M{
		"_id":     eventID,
		"songIds": bson.M{"$nin": bson.A{songID}},
	}

	update = bson.M{
		"$push": bson.M{
			"songIds": bson.M{
				"$each":     bson.A{songID},
				"$position": newPosition,
			},
		},
	}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *EventRepository) PullSongID(eventID, songID bson.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	filter := bson.M{"_id": eventID}

	update := bson.M{
		"$pull": bson.M{
			"songIds": songID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *EventRepository) GetMostFrequentEventNames(bandID bson.ObjectID, limit int, fromUTC time.Time) ([]*entity.EventNameFrequencies, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("events")

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"bandId": bandID,
				"_id": bson.M{
					"$gte": bson.NewObjectIDFromTimestamp(fromUTC),
				},
			},
		},
		bson.M{"$sortByCount": "$name"},
		bson.M{"$limit": limit},
	}
	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, nil
	}

	var frequencies []*entity.EventNameFrequencies
	err = cur.All(context.TODO(), &frequencies)
	if err != nil {
		return nil, err
	}

	return frequencies, nil
}
