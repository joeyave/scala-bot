package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type Role struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name,omitempty" json:"name,omitempty"`
	Priority int                `bson:"priority" json:"priority,omitempty"`
	BandID   primitive.ObjectID `bson:"bandId,omitempty" json:"band_id,omitempty"`
}

const (
	AdminRole = "Admin"
)
