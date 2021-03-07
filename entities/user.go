package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
)

type User struct {
	ID      int64                `bson:"_id,omitempty"`
	State   *State               `bson:"state,omitempty"`
	BandIDs []primitive.ObjectID `bson:"bandIds,omitempty"`
	Bands   []*Band              `bson:"-"`
}

type State struct {
	Index   int     `bson:"index,omitempty"`
	Name    string  `bson:"name,omitempty"`
	Context Context `bson:"context,omitempty"`

	Prev *State `bson:"prev,omitempty"`
	Next *State `bson:"next,omitempty"`
}

type Context struct {
	CurrentSong      *Song    `bson:"currentSong,omitempty"`
	Setlist          []string `bson:"setlist,omitempty"`
	FoundSongs       []*Song  `bson:"foundSongs,omitempty"`
	MessagesToDelete []int    `bson:"messagesToDelete,omitempty"`
	Query            string   `bson:"query,omitempty"`

	DriveFiles []*drive.File `bson:"driveFiles,omitempty"`

	CurrentVoice *Voice `bson:"currentVoice,omitempty"`

	Key string `bson:"key,omitempty"`

	Bands       []*Band `bson:"bands,omitempty"`
	CurrentBand *Band   `bson:"currentBand,omitempty"`

	Events []*Event `bson:events,omitempty`
}

func (u *User) GetFolderIDs() []string {
	folderIDs := make([]string, 0)

	for i := range u.Bands {
		folderIDs = append(folderIDs, u.Bands[i].DriveFolderID)
	}

	return folderIDs
}
