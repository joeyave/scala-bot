package entity

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/gorilla/schema"
	"github.com/joeyave/scala-bot/util"
	"github.com/klauspost/lctime"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"net/url"
	"sort"
)

type User struct {
	ID   int64  `bson:"_id,omitempty" json:"id,omitempty"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	Role string `bson:"role" json:"role,omitempty"`

	State         State         `bson:"state,omitempty" json:"state"`
	Cache         Cache         `bson:"cache" json:"-"`
	CallbackCache CallbackCache `bson:"-" json:"-"`

	BandID primitive.ObjectID `bson:"bandId,omitempty" json:"band_id,omitempty"`
	Band   *Band              `bson:"band,omitempty" json:"-"`

	BandIDs []primitive.ObjectID `bson:"bandIDs,omitempty" json:"bandIDs,omitempty"`
}

func (u *User) IsAdmin() bool {
	return u.Role == AdminRole
}

func (u *User) IsEventMember(event *Event) bool {
	for _, membership := range event.Memberships {
		if u.ID == membership.UserID {
			return true
		}
	}
	return false
}

type State struct {
	Name  int `bson:"name,omitempty"`
	Index int `bson:"index,omitempty"`
}

type Cache struct {
	Query     string `bson:"query,omitempty"`
	Filter    string `bson:"filter,omitempty"`
	PageIndex int    `bson:"page_index,omitempty"`

	Buttons []gotgbot.KeyboardButton `bson:"buttons,omitempty"`

	DriveFiles []*drive.File `bson:"drive_files,omitempty"`

	NextPageToken *NextPageToken     `bson:"next_page_token,omitempty"`
	SongNames     []string           `bson:"song_names,omitempty"`
	DriveFileIDs  []string           `bson:"drive_file_ids,omitempty"`
	Voice         *Voice             `bson:"voice,omitempty"`
	SongID        primitive.ObjectID `bson:"song_id,omitempty"`
	Band          *Band              `bson:"band,omitempty"`
	Role          *Role              `bson:"role,omitempty"`
	Audio         *gotgbot.Audio     `json:"audio,omitempty"`
}

type CallbackCache struct {
	EventIDHex string `schema:"eventId,omitempty"`
	JsonString string `schema:"jsonString,omitempty"`

	ChatID    int64 `schema:"chatId,omitempty"`
	MessageID int64 `schema:"messageId,omitempty"`
	UserID    int64 `schema:"userId,omitempty"`

	AudioFileId    string `schema:"audioFileId,omitempty"`
	AudioDuration  int64  `schema:"audioDuration,omitempty"`
	AudioPerformer string `schema:"audioPerformer,omitempty"`
	AudioTitle     string `schema:"audioTitle,omitempty"`
	AudioFileName  string `schema:"audioFileName,omitempty"`
	AudioMimeType  string `schema:"audioMimeType,omitempty"`
	AudioFileSize  int64  `schema:"audioFileSize,omitempty"`

	AudioThumbFileId       string `schema:"thumbFileId,omitempty"`
	AudioThumbFileUniqueId string `schema:"thumbFileUniqueId,omitempty"`
	AudioThumbWidth        int64  `schema:"thumbWidth,omitempty"`
	AudioThumbHeight       int64  `schema:"thumbHeight,omitempty"`
	AudioThumbFileSize     int64  `schema:"thumbFileSize,omitempty"`

	IsVoice bool `schema:"isVoice,omitempty"`

	//VoiceFileId   string `schema:"voiceFileId,omitempty"`
	//VoiceDuration int64  `schema:"voiceDuration,omitempty"`
	//VoiceMimeType string `schema:"voiceMimeType,omitempty"`
	//VoiceFileSize int64  `schema:"voiceFileSize,omitempty"`
}

var encoder = schema.NewEncoder()

func (c *CallbackCache) AddToText(text string) string {
	values := url.Values{}
	encoder.Encode(c, values)
	u, _ := url.Parse(util.CallbackCacheURL)
	u.RawQuery = values.Encode()
	//return fmt.Sprintf("%s\n\n<a href=\"%s\">cache</a>", text, u.String())
	return fmt.Sprintf("%s <a href=\"%s\">&#8203;</a>", text, u.String())
}

type NextPageToken struct {
	Value string         `bson:"value"`
	Prev  *NextPageToken `bson:"prev,omitempty"`
}

func (t *NextPageToken) GetValue() string {
	if t != nil {
		return t.Value
	}
	return ""
}

func (t *NextPageToken) GetPrevValue() string {
	if t != nil && t.Prev != nil {
		return t.Prev.Value
	}
	return ""
}

type UserWithEvents struct {
	User `bson:",inline"`

	Events []*Event `bson:"events,omitempty"`
}

func (u *UserWithEvents) NameWithStats() string {
	return fmt.Sprintf("%s (%v, %d)", u.User.Name, lctime.Strftime("%d %b", u.Events[0].Time), len(u.Events))
}

func (u *UserWithEvents) String(lang string) string {
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
			t, _ := lctime.StrftimeLoc(util.IetfToIsoLangCode(lang), "%a", mp2[k][0].Time)
			str = fmt.Sprintf("%s%s %d, ", str, t, len(mp2[k]))
		}
		str = fmt.Sprintf("%s)", str[:len(str)-2])
	}

	return str
}
