package entities

import (
	"fmt"
	"github.com/klauspost/lctime"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"gopkg.in/telebot.v3"
	"net/url"
	"sort"
	"time"
)

type User struct {
	ID    int64  `bson:"_id,omitempty"`
	Name  string `bson:"name,omitempty"`
	Role  string `bson:"role,omitempty"`
	State *State `bson:"state,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`
}

type UserExtra struct {
	User *User `bson:",inline"`

	Events []*Event `bson:"events,omitempty"`
}

func (u *UserExtra) String() string {
	str := fmt.Sprintf("<b><a href=\"tg://user?id=%d\">%s</a></b>\nВсего участий: %d", u.User.ID, u.User.Name, len(u.Events))

	if len(u.Events) > 0 {
		str = fmt.Sprintf("%s\nИз них:", str)
	}

	mp := map[Role][]*Event{}

	for _, event := range u.Events {
		for _, membership := range event.Memberships {
			if membership.UserID == u.User.ID {
				mp[*membership.Role] = append(mp[*membership.Role], event)
				break
			}
		}
	}

	keys := make([]Role, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Name < keys[j].Name
	})
	keys = append(keys[1:], keys[0])

	for _, role := range keys {
		mp2 := map[int][]*Event{}
		for _, event := range mp[role] {
			mp2[int(event.Time.Weekday())] = append(mp2[int(event.Time.Weekday())], event)
		}
		str = fmt.Sprintf("%s\n - %s: %d", str, role.Name, len(mp[role]))

		keys := make([]int, 0, len(mp2))
		for k := range mp2 {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		keys = append(keys[1:], keys[0])

		str = fmt.Sprintf("%s (", str)
		for _, k := range keys {
			str = fmt.Sprintf("%s%s %d, ", str, lctime.Strftime("%a", mp2[k][0].Time), len(mp2[k]))
		}
		str = fmt.Sprintf("%s)", str[:len(str)-2])
	}

	return str
}

type State struct {
	Index        int      `bson:"index,omitempty"`
	Name         int      `bson:"name,omitempty"`
	Context      Context  `bson:"context,omitempty"`
	CallbackData *url.URL `bson:"-"`

	Prev *State `bson:"prev,omitempty"`
	Next *State `bson:"next,omitempty"`
}

type Context struct {
	SongNames        []string `bson:"songNames,omitempty"`
	MessagesToDelete []int    `bson:"messagesToDelete,omitempty"`
	Query            string   `bson:"query,omitempty"`
	QueryType        string   `bson:"queryType,omitempty"`

	DriveFileID       string        `bson:"currentSongId,omitempty"`
	FoundDriveFileIDs []string      `bson:"foundDriveFileIds,omitempty"`
	DriveFiles        []*drive.File `bson:"driveFiles,omitempty"`

	Voice *Voice `bson:"currentVoice,omitempty"`

	Band  *Band   `bson:"currentBand,omitempty"`
	Bands []*Band `bson:"bands,omitempty"`

	Role *Role `bson:"role,omitempty"`

	EventID primitive.ObjectID `bson:"eventId,omitempty"`

	CreateSongPayload struct {
		Name   string `bson:"name,omitempty"`
		Lyrics string `bson:"lyrics,omitempty"`
		Key    string `bson:"key,omitempty"`
		BPM    string `bson:"bpm,omitempty"`
		Time   string `bson:"time,omitempty"`
	} `bson:"createSongPayload,omitempty"`

	Map  map[string]string `bson:"map,omitempty"`
	Time time.Time         `bson:"time,omitempty"`

	PageIndex int `bson:"index, omitempty"`

	NextPageToken  *NextPageToken        `bson:"nextPageToken,omitempty"`
	WeekdayButtons []telebot.ReplyButton `bson:"weekday_buttons,omitempty"`
	PrevText       string                `json:"prev_text,omitempty"`
}

type NextPageToken struct {
	Token         string         `bson:"token"`
	PrevPageToken *NextPageToken `bson:"prevPageToken,omitempty"`
}
