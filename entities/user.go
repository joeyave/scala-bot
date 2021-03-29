package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
)

type User struct {
	ID    int64  `bson:"_id,omitempty"`
	Name  string `bson:"name,omitempty"`
	Role  string `bson:"role,omitempty"`
	State *State `bson:"state,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`
}

type State struct {
	Index   int     `bson:"index,omitempty"`
	Name    string  `bson:"name,omitempty"`
	Context Context `bson:"context,omitempty"`

	Prev *State `bson:"prev,omitempty"`
	Next *State `bson:"next,omitempty"`
}

type Context struct {
	SongNames        []string `bson:"songNames,omitempty"`
	MessagesToDelete []int    `bson:"messagesToDelete,omitempty"`
	Query            string   `bson:"query,omitempty"`

	DriveFileID       string        `bson:"currentSongId,omitempty"`
	FoundDriveFileIDs []string      `bson:"foundSongIds,omitempty"`
	DriveFiles        []*drive.File `bson:"driveFiles,omitempty"`

	CurrentBandID primitive.ObjectID `bson:"currentBandId,omitempty"`

	CurrentVoice *Voice `bson:"currentVoice,omitempty"`

	Key string `bson:"key,omitempty"`

	Bands       []*Band `bson:"bands,omitempty"`
	CurrentBand *Band   `bson:"currentBand,omitempty"`

	Events []*Event `bson:"events,omitempty"`

	CreateSongPayload struct {
		Name   string `bson:"name,omitempty"`
		Lyrics string `bson:"lyrics,omitempty"`
		Key    string `bson:"key,omitempty"`
		BPM    string `bson:"bpm,omitempty"`
		Time   string `bson:"time,omitempty"`
	} `bson:"createSongPayload,omitempty"`
}
