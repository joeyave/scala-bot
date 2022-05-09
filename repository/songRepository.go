package repository

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-bot-v2/entity"
	"github.com/joeyave/scala-bot-v2/helpers"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *SongRepository) FindAll() ([]*entity.Song, error) {
	return r.find(bson.M{})
}

func (r *SongRepository) FindManyLiked(userID int64) ([]*entity.Song, error) {
	return r.find(bson.M{
		"likes": bson.M{"$in": bson.A{userID}},
	})
}

func (r *SongRepository) FindManyByDriveFileIDs(IDs []string) ([]*entity.Song, error) {

	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"driveFileId": bson.M{
					"$in": IDs,
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"__order": bson.M{
					"$indexOfArray": bson.A{IDs, "$driveFileId"},
				},
			},
		},
		bson.M{
			"$sort": bson.M{"__order": 1},
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

	var songs []*entity.Song
	for cur.Next(context.TODO()) {
		var song *entity.Song
		err := cur.Decode(&song)
		if err != nil {
			continue
		}

		songs = append(songs, song)
	}

	if len(songs) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return songs, nil
}

func (r *SongRepository) FindOneByID(ID primitive.ObjectID) (*entity.Song, error) {
	songs, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}
	return songs[0], nil
}

func (r *SongRepository) FindOneByDriveFileID(driveFileID string) (*entity.Song, error) {
	songs, err := r.find(bson.M{"driveFileId": driveFileID})
	if err != nil {
		return nil, err
	}
	return songs[0], nil
}

func (r *SongRepository) FindOneByName(name string) (*entity.Song, error) {
	songs, err := r.find(bson.M{"pdf.name": name})
	if err != nil {
		return nil, err
	}
	return songs[0], nil
}

func (r *SongRepository) find(m bson.M) ([]*entity.Song, error) {
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

	var songs []*entity.Song
	for cur.Next(context.TODO()) {
		var song *entity.Song
		err := cur.Decode(&song)
		if err != nil {
			continue
		}

		songs = append(songs, song)
	}

	if len(songs) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return songs, nil
}

func (r *SongRepository) UpdateOne(song entity.Song) (*entity.Song, error) {
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

	var newSong *entity.Song
	err := result.Decode(&newSong)
	if err != nil {
		return nil, err
	}

	// channel, err := r.driveClient.Files.Watch(song.DriveFileID, &drive.Channel{
	//	Address: fmt.Sprintf("%s/driveFileChangeCallback", os.Getenv("HOST")),
	//	Id:      uuid.New().String(),
	//	Kind:    "api#channel",
	//	Type:    "web_hook",
	// }).Do()
	//
	// fmt.Println(channel, err)

	return r.FindOneByID(newSong.ID)
}

func (r *SongRepository) DeleteOneByDriveFileID(driveFileID string) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	_, err := collection.DeleteOne(context.TODO(), bson.M{"driveFileId": driveFileID})
	return err
}

func (r *SongRepository) Like(songID primitive.ObjectID, userID int64) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	filter := bson.M{
		"_id":   songID,
		"likes": bson.M{"$nin": bson.A{userID}},
	}

	update := bson.M{
		"$push": bson.M{
			"likes": userID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *SongRepository) Dislike(songID primitive.ObjectID, userID int64) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	filter := bson.M{"_id": songID}

	update := bson.M{
		"$pull": bson.M{
			"likes": userID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
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

func (r *SongRepository) FindAllExtraByPageNumberSortedByEventsNumber(bandID primitive.ObjectID, pageNumber int) ([]*entity.SongWithEvents, error) {

	return r.findWithExtra(
		bson.M{
			"bandId": bandID,
		},
		bson.M{
			"$addFields": bson.M{
				"eventsSize": bson.M{"$size": "$events"},
			},
		},
		bson.M{
			"$sort": bson.D{
				{"eventsSize", -1},
				{"_id", 1},
			},
		},
		bson.M{
			"$skip": pageNumber * helpers.SongsPageSize,
		},
		bson.M{
			"$limit": helpers.SongsPageSize,
		},
	)
}

func (r *SongRepository) FindAllExtraByPageNumberSortedByLatestEventDate(bandID primitive.ObjectID, pageNumber int) ([]*entity.SongWithEvents, error) {

	return r.findWithExtra(
		bson.M{
			"bandId": bandID,
		},
		bson.M{
			"$sort": bson.D{
				{"events.0.time", -1},
				{"_id", 1},
			},
		},
		bson.M{
			"$skip": pageNumber * helpers.SongsPageSize,
		},
		bson.M{
			"$limit": helpers.SongsPageSize,
		},
	)
}

func (r *SongRepository) FindManyExtraByTag(tag string, bandID primitive.ObjectID, pageNumber int) ([]*entity.SongWithEvents, error) {

	return r.findWithExtra(
		bson.M{
			"bandId": bandID,
			"tags":   tag,
		},
		bson.M{
			"$skip": pageNumber * helpers.SongsPageSize,
		},
		bson.M{
			"$limit": helpers.SongsPageSize,
		},
	)
}

func (r *SongRepository) FindManyExtraByDriveFileIDs(driveFileIDs []string) ([]*entity.SongWithEvents, error) {
	return r.findWithExtra(
		bson.M{
			"driveFileId": bson.M{
				"$in": driveFileIDs,
			},
		},
	)
}

func (r *SongRepository) FindManyExtraByPageNumberLiked(userID int64, pageNumber int) ([]*entity.SongWithEvents, error) {
	return r.findWithExtra(
		bson.M{
			"likes": bson.M{"$in": bson.A{userID}},
		},
		bson.M{
			"$skip": pageNumber * helpers.SongsPageSize,
		},
		bson.M{
			"$limit": helpers.SongsPageSize,
		},
	)
}

func (r *SongRepository) findWithExtra(m bson.M, opts ...bson.M) ([]*entity.SongWithEvents, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

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
			"$lookup": bson.M{
				"from": "events",
				"let":  bson.M{"songId": "$_id"},
				"pipeline": bson.A{
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
						"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$$songId", "$songIds"}}},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "memberships",
							"let":  bson.M{"eventId": "$_id"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{
										"$expr": bson.M{"$eq": bson.A{"$eventId", "$$eventId"}},
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
									"$sort": bson.M{
										"role.priority": 1,
									},
								},
							},
							"as": "memberships",
						},
					},
					bson.M{
						"$sort": bson.M{
							"time": -1,
						},
					},
				},
				"as": "events",
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

	var songs []*entity.SongWithEvents
	err = cur.All(context.TODO(), &songs)
	return songs, err
}

func (r *SongRepository) GetTags(bandID primitive.ObjectID) ([]string, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	pipeline := bson.A{
		bson.M{"$match": bson.M{"bandId": bandID}},
		bson.M{"$unwind": "$tags"},
		bson.M{"$sortByCount": "$tags"},
	}
	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, nil
	}

	var frequencies []*entity.SongTagFrequencies
	err = cur.All(context.TODO(), &frequencies)
	if err != nil {
		return nil, err
	}

	// tagsRaw, err := collection.Distinct(context.TODO(), "tags", bson.D{}, nil)
	// if err != nil {
	// 	return nil, err
	// }

	tags := make([]string, len(frequencies))
	for i, v := range frequencies {
		tags[i] = v.Tag
	}

	sort.Strings(tags)

	return tags, nil
}

func (r *SongRepository) TagOrUntag(tag string, songID primitive.ObjectID) (*entity.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("songs")

	filter := bson.M{
		"_id": songID,
	}

	update := bson.A{
		bson.M{
			"$addFields": bson.M{
				"tags": bson.M{
					"$cond": bson.M{
						"if": bson.M{
							"$ne": bson.A{bson.M{"$type": "$tags"}, "array"},
						},
						"then": bson.A{},
						"else": "$tags",
					},
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"tags": bson.M{
					"$cond": bson.A{
						bson.M{
							"$in": bson.A{tag, "$tags"},
						},
						bson.M{
							"$setDifference": bson.A{"$tags", bson.A{tag}},
						},
						bson.M{
							"$concatArrays": bson.A{"$tags", bson.A{tag}},
						},
					},
				},
			},
		},
	}
	// update := bson.M{
	// 	"$addToSet": bson.M{
	// 		"tags": tag,
	// 	},
	// }

	after := options.After
	opts := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, &opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var song *entity.Song
	err := result.Decode(&song)
	if err != nil {
		return nil, err
	}

	return song, err
}
