package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Band struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name"`
	DriveFolderID string             `bson:"driveFolderId"`
	AdminUserIDs  []int64            `bson:"adminUserIds"`
}
