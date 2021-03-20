package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Voice struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	FileID  string             `bson:"fileId,omitempty"`
	Caption string             `bson:"caption,omitempty"`

	SongID primitive.ObjectID `bson:"songId,omitempty"`
}
