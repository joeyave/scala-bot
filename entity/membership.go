package entity

import "go.mongodb.org/mongo-driver/v2/bson"

type Membership struct {
	ID      bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	EventID bson.ObjectID `bson:"eventId,omitempty" json:"event_id,omitempty"`

	UserID int64 `bson:"userId,omitempty" json:"user_id,omitempty"`
	User   *User `bson:"user,omitempty" json:"user,omitempty"`

	RoleID bson.ObjectID `bson:"roleId,omitempty" json:"role_id,omitempty"`
	Role   *Role              `bson:"role,omitempty" json:"role,omitempty"`

	Notified bool `bson:"notified,omitempty" json:"-"`
}
