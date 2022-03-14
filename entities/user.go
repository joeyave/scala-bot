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
	ID    int64  `bson:"_id,omitempty" json:"id"`
	Name  string `bson:"name,omitempty" json:"name"`
	Role  string `bson:"role,omitempty" json:"role"`
	State *State `bson:"state,omitempty" json:"state,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty" json:"band_id,omitempty"`
	Band   *Band              `bson:"band,omitempty" json:"-"`
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
	Index        int      `bson:"index,omitempty" json:"index"`
	Name         int      `bson:"name,omitempty" json:"name"`
	Context      Context  `bson:"context,omitempty" json:"-"`
	CallbackData *url.URL `bson:"-" json:"-"`

	Prev *State `bson:"prev,omitempty" json:"prev,omitempty"`
	Next *State `bson:"next,omitempty" json:"next,omitempty"`
}

type Context struct {
	SongNames        []string `bson:"songNames,omitempty" json:"song_names,omitempty"`
	MessagesToDelete []int    `bson:"messagesToDelete,omitempty" json:"messages_to_delete,omitempty"`
	Query            string   `bson:"query,omitempty" json:"query,omitempty"`
	QueryType        string   `bson:"queryType,omitempty" json:"query_type,omitempty"`

	DriveFileID       string        `bson:"currentSongId,omitempty" json:"drive_file_id,omitempty"`
	FoundDriveFileIDs []string      `bson:"foundDriveFileIds,omitempty" json:"found_drive_file_i_ds,omitempty"`
	DriveFiles        []*drive.File `bson:"driveFiles,omitempty" json:"drive_files,omitempty"`

	Voice *Voice `bson:"currentVoice,omitempty" json:"voice,omitempty"`

	Band  *Band   `bson:"currentBand,omitempty" json:"band,omitempty"`
	Bands []*Band `bson:"bands,omitempty" json:"bands,omitempty"`

	Role *Role `bson:"role,omitempty" json:"role,omitempty"`

	EventID primitive.ObjectID `bson:"eventId,omitempty" json:"event_id,omitempty"`

	CreateSongPayload struct {
		Name   string `bson:"name,omitempty" json:"name,omitempty"`
		Lyrics string `bson:"lyrics,omitempty" json:"lyrics,omitempty"`
		Key    string `bson:"key,omitempty" json:"key,omitempty"`
		BPM    string `bson:"bpm,omitempty" json:"bpm,omitempty"`
		Time   string `bson:"time,omitempty" json:"time,omitempty"`
	} `bson:"createSongPayload,omitempty" json:"create_song_payload"`

	Map  map[string]string `bson:"map,omitempty" json:"map,omitempty"`
	Time time.Time         `bson:"time,omitempty" json:"time"`

	PageIndex int `bson:"index, omitempty" json:"page_index,omitempty"`

	NextPageToken  *NextPageToken        `bson:"nextPageToken,omitempty" json:"next_page_token,omitempty"`
	WeekdayButtons []telebot.ReplyButton `bson:"weekday_buttons,omitempty" json:"weekday_buttons,omitempty"`
	PrevText       string                `bson:"prev_text,omitempty" json:"prev_text,omitempty"`
}

type NextPageToken struct {
	Token         string         `bson:"token" json:"token,omitempty"`
	PrevPageToken *NextPageToken `bson:"prevPageToken,omitempty" json:"prev_page_token,omitempty"`
}
