package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Band struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name,omitempty" json:"name,omitempty"`
	DriveFolderID string             `bson:"driveFolderId,omitempty" json:"drive_folder_id,omitempty"`

	Roles []*Role `bson:"roles,omitempty" json:"roles,omitempty"`
}
