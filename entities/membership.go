package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Membership struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	EventID primitive.ObjectID `bson:"eventId,omitempty"`
	UserID  int64              `bson:"userId,omitempty"`

	RoleID primitive.ObjectID `bson:"roleId,omitempty"`
	Role   *Role              `bson:"role,omitempty"`

	Notified bool `bson:"notified,omitempty"`
}
