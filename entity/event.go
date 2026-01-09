package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/klauspost/lctime"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SongOverride represents a song in an event's setlist with optional key override and caching
type SongOverride struct {
	SongID   primitive.ObjectID `bson:"songId" json:"songId"`
	EventKey Key                `bson:"eventKey,omitempty" json:"eventKey,omitempty"` // Key override for this event
}

type Event struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TimeUTC time.Time          `bson:"time,omitempty" json:"time"`
	Name    string             `bson:"name,omitempty" json:"name"`

	Memberships []*Membership `bson:"memberships,omitempty" json:"memberships"`

	BandID primitive.ObjectID `bson:"bandId,omitempty" json:"bandId"`
	Band   *Band              `bson:"band,omitempty" json:"band"`

	SongIDs       []primitive.ObjectID `bson:"songIds" json:"songIds"`
	SongOverrides []SongOverride       `bson:"songOverrides,omitempty" json:"songOverrides,omitempty"`

	Songs []*Song `bson:"songs,omitempty" json:"songs"`

	Notes *string `bson:"notes" json:"notes"`
}

func (e *Event) GetSongOverride(songID primitive.ObjectID) *SongOverride {
	for _, song := range e.SongOverrides {
		if song.SongID == songID {
			return &song
		}
	}
	return nil
}

func (e *Event) GetLocalTime() time.Time {
	if e.Band == nil {
		return e.TimeUTC
	}
	return e.TimeUTC.In(e.Band.GetLocation())
}

func (e *Event) Alias(lang string) string {
	timeLoc := e.GetLocalTime()
	format := "%A, %d.%m.%Y %H:%M"
	if timeLoc.Hour() == 0 && timeLoc.Minute() == 0 {
		format = "%A, %d.%m.%Y"
	}
	t, _ := lctime.StrftimeLoc(util.IetfToIsoLangCode(lang), format, timeLoc)
	return fmt.Sprintf("%s (%s)", e.Name, t)
}

func (e *Event) RolesString() string {

	var b strings.Builder

	var currRoleID primitive.ObjectID
	for _, membership := range e.Memberships {
		if membership.User == nil {
			continue
		}

		if currRoleID != membership.RoleID {
			currRoleID = membership.RoleID
			fmt.Fprintf(&b, "\n\n<b>%s:</b>", membership.Role.Name)
		}

		fmt.Fprintf(&b, "\n - <a href=\"tg://user?id=%d\">%s</a>", membership.User.ID, membership.User.Name)
	}

	return strings.TrimSpace(b.String())
}

func (e *Event) NotesString(lang string) string {
	return fmt.Sprintf("<b>%s:</b>\n%s", txt.Get("button.notes", lang), *e.Notes)
}

type EventNameFrequencies struct {
	Name  string `bson:"_id"`
	Count int    `bson:"count"`
}
