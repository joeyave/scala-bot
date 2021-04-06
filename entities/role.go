package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Role struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name,omitempty"`
	Priority int                `bson:"priority,omitempty"`
	BandID   primitive.ObjectID `bson:"bandId,omitempty"`
}
