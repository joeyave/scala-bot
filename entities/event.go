package entities

import (
	"fmt"
	"github.com/klauspost/lctime"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Event struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Time time.Time          `bson:"time,omitempty"`
	Name string             `bson:"name,omitempty"`

	Memberships []*Membership `bson:"memberships,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`
}

func (e *Event) Alias() string {
	err := lctime.SetLocale("ru_RU")
	if err != nil {
		fmt.Println(err)
	}

	timeStr := lctime.Strftime("%A / %d %b", e.Time)

	return fmt.Sprintf("%s / %s", timeStr, e.Name)
}
