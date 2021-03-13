package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
)

type User struct {
	ID     int64              `bson:"_id,omitempty"`
	State  *State             `bson:"state,omitempty"`
	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"-"`
}

type State struct {
	Index   int     `bson:"index,omitempty"`
	Name    string  `bson:"name,omitempty"`
	Context Context `bson:"context,omitempty"`

	Prev *State `bson:"prev,omitempty"`
	Next *State `bson:"next,omitempty"`
}

type Context struct {
	Setlist          []string `bson:"setlist,omitempty"`
	MessagesToDelete []int    `bson:"messagesToDelete,omitempty"`
	Query            string   `bson:"query,omitempty"`

	CurrentSongID string        `bson:"currentSongId,omitempty"`
	FoundSongIDs  []string      `bson:"foundSongIds,omitempty"`
	DriveFiles    []*drive.File `bson:"driveFiles,omitempty"`

	CurrentVoice *Voice `bson:"currentVoice,omitempty"`

	Key string `bson:"key,omitempty"`

	Bands       []*Band `bson:"bands,omitempty"`
	CurrentBand *Band   `bson:"currentBand,omitempty"`

	Events []*Event `bson:"events,omitempty"`
}
