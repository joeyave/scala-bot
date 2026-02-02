package entity

import "go.mongodb.org/mongo-driver/v2/bson"

type Role struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name,omitempty" json:"name,omitempty"`
	Priority int                `bson:"priority" json:"priority,omitempty"`
	BandID   bson.ObjectID `bson:"bandId,omitempty" json:"band_id,omitempty"`
}

const (
	AdminRole = "Admin"
)
