package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Band struct {
	ID            primitive.ObjectID   `bson:"_id"`
	Name          string               `bson:"name"`
	DriveFolderID string               `bson:"driveFolderId"`
	AdminUserIDs  []primitive.ObjectID `bson:"adminUserIds"`
}
