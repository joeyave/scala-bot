package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Band struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name,omitempty"`
	DriveFolderID string             `bson:"driveFolderId,omitempty"`
	AdminUserIDs  []int64            `bson:"adminUserIds,omitempty"`
}
