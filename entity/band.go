package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Band struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string             `bson:"name,omitempty" json:"name,omitempty"`
	DriveFolderID   string             `bson:"driveFolderId,omitempty" json:"driveFolderId,omitempty"`
	ArchiveFolderID string             `bson:"archiveFolderId,omitempty" json:"archiveFolderId,omitempty"`
	TempFolderID    string             `bson:"tempFolderID,omitempty" json:"tempFolderID,omitempty"`
	Roles           []*Role            `bson:"roles,omitempty" json:"roles,omitempty"`
	Timezone        string             `bson:"timezone,omitempty" json:"timezone,omitempty"`
}

func (b *Band) GetLocation() *time.Location {
	loc, err := time.LoadLocation(b.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (b *Band) GetNowTime() time.Time {
	return time.Now().In(b.GetLocation())
}
