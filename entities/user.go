package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
)

type User struct {
	ID      int64                `bson:"_id"`
	State   *State               `bson:"states"`
	BandIDs []primitive.ObjectID `bson:"bandIds"`
	Bands   []Band               `bson:"-"`
}

type State struct {
	Index   int     `bson:"index"`
	Name    string  `bson:"name"`
	Context Context `bson:"context"`

	Prev *State `bson:"prev"`
	Next *State `bson:"next"`
}

type Context struct {
	CurrentSong *Song `bson:"currentSong, omitempty"`
	//Songs            []Song        `bson:"songs"`
	Setlist          []string `bson:"setlist"`
	FoundSongs       []*Song  `bson:"foundSongs"`
	MessagesToDelete []int    `bson:"messagesToDelete"`
	Query            string   `bson:"query,omitempty"`

	DriveFiles []*drive.File `bson:"driveFiles,omitempty"`

	CurrentVoice *Voice `bson:"currentVoice"`

	Key string `bson:"key"`

	Bands       []Band `bson:"bands"`
	CurrentBand Band   `bson:"band"`
}

func (u *User) GetFolderIDs() []string {
	folderIDs := make([]string, 0)

	for i := range u.Bands {
		folderIDs = append(folderIDs, u.Bands[i].DriveFolderID)
	}

	return folderIDs
}
