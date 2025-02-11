package repository

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/helpers"
	"sort"
	"time"

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

func (r *SongRepository) FindManyLiked(bandID primitive.ObjectID, userID int64) ([]*entity.Song, error) {
	return r.find(
		bson.M{
			"bandId": bandID,
			"likes":  bson.M{"$elemMatch": bson.M{"userId": userID}},
		},
		bson.M{
			"$sort": bson.M{
				"likes.time": -1,
			},
		},
	)
}

func (r *SongRepository) FindManyByDriveFileIDs(IDs []string) ([]*entity.Song, error) {

	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

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

func (r *SongRepository) FindOneByName(name string, bandID primitive.ObjectID) (*entity.Song, error) {
	songs, err := r.find(bson.M{
		"bandId":   bandID,
		"pdf.name": name,
	})
	if err != nil {
		return nil, err
	}
	return songs[0], nil
}

func (r *SongRepository) find(m bson.M, opts ...bson.M) ([]*entity.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

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

	for _, o := range opts {
		pipeline = append(pipeline, o)
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
		song.ID = primitive.NewObjectID()
	}

	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

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
	//	Address: fmt.Sprintf("%s/driveFileChangeCallback", os.Getenv("BOT_DOMAIN")),
	//	Id:      uuid.New().String(),
	//	Kind:    "api#channel",
	//	Type:    "web_hook",
	// }).Do()
	//
	// fmt.Println(channel, err)

	return r.FindOneByID(newSong.ID)
}

func (r *SongRepository) UpdateMany(songs []*entity.Song) (*mongo.BulkWriteResult, error) {

	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

	var models []mongo.WriteModel
	for _, song := range songs {
		if song.ID.IsZero() {
			song.ID = primitive.NewObjectID()
		}

		upsert := true
		song.Band = nil
		song.Voices = nil
		model := &mongo.UpdateOneModel{
			Upsert: &upsert,
			Filter: bson.M{"_id": song.ID},
			Update: bson.M{"$set": song},
		}
		models = append(models, model)
	}

	ordered := false
	res, err := collection.BulkWrite(context.TODO(), models, &options.BulkWriteOptions{
		Ordered: &ordered,
	})
	if err != nil {
		return nil, err
	}

	return res, err
}

func (r *SongRepository) DeleteOneByDriveFileID(driveFileID string) (int64, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

	result, err := collection.DeleteOne(context.TODO(), bson.M{"driveFileId": driveFileID})

	deletedCount := int64(0)
	if result != nil {
		deletedCount = result.DeletedCount
	}

	return deletedCount, err
}

func (r *SongRepository) Archive(songID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

	filter := bson.M{
		"_id": songID,
	}

	update := bson.M{
		"$set": bson.M{
			"isArchived": true,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *SongRepository) Unarchive(songID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

	filter := bson.M{
		"_id": songID,
	}

	update := bson.M{
		"$set": bson.M{
			"isArchived": false,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *SongRepository) Like(songID primitive.ObjectID, userID int64) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

	// Create a new Like struct with the user ID and the current time.
	newLike := &entity.Like{
		UserID: userID,
		Time:   time.Now(),
	}

	filter := bson.M{
		"_id":   songID,
		"likes": bson.M{"$not": bson.M{"$elemMatch": bson.M{"userId": userID}}},
	}

	update := bson.M{
		"$push": bson.M{
			"likes": newLike,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *SongRepository) Dislike(songID primitive.ObjectID, userID int64) error {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

	filter := bson.M{"_id": songID}

	update := bson.M{
		"$pull": bson.M{
			"likes": bson.M{"userId": userID},
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *SongRepository) FindOneWithExtraByID(songID primitive.ObjectID, eventsStartDate time.Time) (*entity.SongWithEvents, error) {

	songs, err := r.findWithExtra(
		bson.M{
			"_id": songID,
		},
		eventsStartDate,
	)
	if err != nil {
		return nil, err
	}
	return songs[0], nil
}

func (r *SongRepository) FindAllExtraByPageNumberSortedByEventsNumber(bandID primitive.ObjectID, eventsStartDate time.Time, isAscending bool, pageNumber int) ([]*entity.SongWithEvents, error) {

	sortingValue := -1
	if isAscending {
		sortingValue = 1
	}

	return r.findWithExtra(
		bson.M{
			"bandId": bandID,
			"$or": []bson.M{
				{"isArchived": bson.M{"$eq": false}},
				{"isArchived": bson.M{"$exists": false}},
			},
		},
		eventsStartDate,
		bson.M{
			"$addFields": bson.M{
				"eventsSize": bson.M{"$size": "$events"},
			},
		},
		bson.M{
			"$sort": bson.D{
				{Key: "eventsSize", Value: sortingValue},
				{Key: "_id", Value: 1},
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

func (r *SongRepository) FindAllExtraByPageNumberSortedByEventDate(bandID primitive.ObjectID, eventsStartDate time.Time, isAscending bool, pageNumber int) ([]*entity.SongWithEvents, error) {

	sortingValue := -1
	if isAscending {
		sortingValue = 1
	}

	return r.findWithExtra(
		bson.M{
			"bandId": bandID,
			"$or": []bson.M{
				{"isArchived": bson.M{"$eq": false}},
				{"isArchived": bson.M{"$exists": false}},
			},
		},
		eventsStartDate,
		bson.M{
			"$sort": bson.D{
				{Key: "events.0.time", Value: sortingValue},
				{Key: "_id", Value: 1},
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

func (r *SongRepository) FindManyExtraByTag(tag string, bandID primitive.ObjectID, eventsStartDate time.Time, pageNumber int) ([]*entity.SongWithEvents, error) {

	return r.findWithExtra(
		bson.M{
			"bandId": bandID,
			"tags":   tag,
			"$or": []bson.M{
				{"isArchived": bson.M{"$eq": false}},
				{"isArchived": bson.M{"$exists": false}},
			},
		},
		eventsStartDate,
		bson.M{
			"$skip": pageNumber * helpers.SongsPageSize,
		},
		bson.M{
			"$limit": helpers.SongsPageSize,
		},
	)
}

func (r *SongRepository) FindManyExtraByDriveFileIDs(driveFileIDs []string, eventsStartDate time.Time) ([]*entity.SongWithEvents, error) {
	return r.findWithExtra(
		bson.M{
			"driveFileId": bson.M{
				"$in": driveFileIDs,
			},
		},
		eventsStartDate,
	)
}

func (r *SongRepository) FindManyExtraByPageNumberLiked(bandID primitive.ObjectID, userID int64, eventsStartDate time.Time, pageNumber int) ([]*entity.SongWithEvents, error) {
	return r.findWithExtra(
		bson.M{
			"bandId": bandID,
			"likes": bson.M{
				"$elemMatch": bson.M{"userId": userID},
			},
		},
		eventsStartDate,
		bson.M{
			"$sort": bson.M{
				"likes.time": -1,
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

func (r *SongRepository) findWithExtra(m bson.M, eventsStartDate time.Time, opts ...bson.M) ([]*entity.SongWithEvents, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

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
						"$match": bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{"$gte": bson.A{"$time", eventsStartDate}},
									bson.M{"$in": bson.A{"$$songId", "$songIds"}},
								},
							},
						},
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
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

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

	tags := make([]string, len(frequencies))
	for i, v := range frequencies {
		tags[i] = v.Tag
	}

	sort.Strings(tags)

	return tags, nil
}

func (r *SongRepository) TagOrUntag(tag string, songID primitive.ObjectID) (*entity.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("BOT_MONGODB_NAME")).Collection("songs")

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
