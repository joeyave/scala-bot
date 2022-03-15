package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type Membership struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	EventID primitive.ObjectID `bson:"eventId,omitempty" json:"event_id,omitempty"`

	UserID int64 `bson:"userId,omitempty" json:"user_id,omitempty"`
	User   *User `bson:"user,omitempty" json:"-"`

	RoleID primitive.ObjectID `bson:"roleId,omitempty" json:"role_id,omitempty"`
	Role   *Role              `bson:"role,omitempty" json:"role,omitempty"`

	Notified bool `bson:"notified,omitempty" json:"-"`
}
