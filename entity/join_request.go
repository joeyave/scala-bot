package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type JoinRequestStatus string

const (
	JoinRequestPending  JoinRequestStatus = "pending"
	JoinRequestApproved JoinRequestStatus = "approved"
	JoinRequestDeclined JoinRequestStatus = "declined"
	JoinRequestCanceled JoinRequestStatus = "canceled"
)

type JoinRequest struct {
	ID              bson.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          int64             `bson:"userId" json:"userId"`
	UserName        string            `bson:"userName" json:"userName"`
	BandID          bson.ObjectID     `bson:"bandId" json:"bandId"`
	BandName        string            `bson:"bandName" json:"bandName"`
	Status          JoinRequestStatus `bson:"status" json:"status"`
	CreatedAt       time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time         `bson:"updatedAt" json:"updatedAt"`
	DecidedAt       *time.Time        `bson:"decidedAt,omitempty" json:"decidedAt,omitempty"`
	DecidedByUserID int64             `bson:"decidedByUserId,omitempty" json:"decidedByUserId,omitempty"`
}
