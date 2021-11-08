package entities

import (
	"fmt"
	"github.com/klauspost/lctime"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

type Event struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Time time.Time          `bson:"time,omitempty"`
	Name string             `bson:"name,omitempty"`

	Memberships []*Membership `bson:"memberships,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`

	SongIDs []primitive.ObjectID `bson:"songIds,omitempty"`
	Songs   []*Song              `bson:"songs,omitempty"`

	Notes string `bson:"notes,omitempty"`
}

func (e *Event) Alias() string {
	timeStr := lctime.Strftime("%A, %d.%m.%Y", e.Time)
	return fmt.Sprintf("%s (%s)", e.Name, timeStr)
}

func (e *Event) Roles() string {
	str := ""

	var currRoleID primitive.ObjectID
	for _, membership := range e.Memberships {
		if membership.User == nil {
			continue
		}

		if currRoleID != membership.RoleID {
			currRoleID = membership.RoleID
			str = fmt.Sprintf("%s\n\n<b>%s:</b>", str, membership.Role.Name)
		}

		str = fmt.Sprintf("%s\n - <a href=\"tg://user?id=%d\">%s</a>", str, membership.User.ID, membership.User.Name)
	}

	return strings.TrimSpace(str)
}

type EventNameFrequencies struct {
	Name  string `bson:"_id"`
	Count int    `bson:"count"`
}
