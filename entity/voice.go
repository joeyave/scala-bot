package entity

import "go.mongodb.org/mongo-driver/v2/bson"

type Voice struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"caption,omitempty"`
	FileID      string             `bson:"fileId,omitempty"`
	AudioFileID string             `bson:"audioFileId,omitempty"`

	SongID bson.ObjectID `bson:"songId,omitempty"`
}
