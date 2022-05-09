package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type Voice struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"caption,omitempty"`
	FileID      string             `bson:"fileId,omitempty"`
	AudioFileID string             `bson:"audioFileId,omitempty"`

	SongID primitive.ObjectID `bson:"songId,omitempty"`
}
