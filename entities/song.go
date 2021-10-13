package entities

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Song struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	DriveFileID string `bson:"driveFileId,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`

	PDF PDF `bson:"pdf,omitempty"`

	Voices []*Voice `bson:"voices,omitempty"`

	Likes []int64 `bson:"likes,omitempty"`
}

type SongExtra struct {
	Song *Song `bson:",inline"`

	Events []*Event `bson:"events,omitempty"`
}

type PDF struct {
	ModifiedTime string `bson:"modifiedTime,omitempty"`

	TgFileID           string `bson:"tgFileId,omitempty"`
	TgChannelMessageID int    `bson:"tgChannelMessageId,omitempty"`

	Name string `bson:"name,omitempty"`
	Key  string `bson:"key,omitempty"`
	BPM  string `bson:"bpm,omitempty"`
	Time string `bson:"time,omitempty"`

	WebViewLink string `bson:"webViewLink,omitempty"`
}

func (s *Song) Caption() string {
	return fmt.Sprintf("%s, %s, %s", s.PDF.Key, s.PDF.BPM, s.PDF.Time)
}
