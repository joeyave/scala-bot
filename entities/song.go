package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Song struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	DriveFileID string `bson:"driveFileId,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`

	PDF PDF `bson:"pdf,omitempty"`

	Voices []*Voice `bson:"voices,omitempty"`
}

type PDF struct {
	ModifiedTime string `bson:"modifiedTime,omitempty"`

	TgFileID           string `bson:"tgFileId,omitempty"`
	TgChannelMessageID int    `bson:"tgChannelMessageId,omitempty"`

	Name string `bson:"name,omitempty"`
	Key  string `bson:"key,omitempty"`
	BPM  int    `bson:"bpm,omitempty"`
	Time string `bson:"time,omitempty"`
}
