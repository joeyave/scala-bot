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

type Key string

type Song struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	DriveFileID string `bson:"driveFileId,omitempty" json:"driveFileId"`

	BandID primitive.ObjectID `bson:"bandId,omitempty" json:"bandId"`
	Band   *Band              `bson:"band,omitempty" json:"band"`

	PDF     PDF            `bson:"pdf,omitempty" json:"pdf"`
	AltPDFs map[Key]AltPDF `bson:"altPdfs,omitempty" json:"altPdfs,omitempty"`
	AltPDF  *AltPDF        `bson:"-" json:"-"`

	Voices []*Voice `bson:"voices,omitempty" json:"-"`

	Likes []*Like  `bson:"likes,omitempty" json:"-"`
	Tags  []string `bson:"tags" json:"tags"`

	IsArchived bool `bson:"isArchived" json:"isArchived"`
}

type Like struct {
	UserID int64     `bson:"userId"`
	Time   time.Time `bson:"time"`
}

type AltPDF struct {
	Key          Key    `bson:"key,omitempty" json:"key,omitempty"`
	ModifiedTime string `bson:"modifiedTime,omitempty" json:"modifiedTime,omitempty"`
	DriveFileID  string `bson:"driveFileId,omitempty" json:"driveFileId,omitempty"`
	TgFileID     string `bson:"tgFileId,omitempty" json:"tgFileId,omitempty"`
}
type PDF struct {
	ModifiedTime string `bson:"modifiedTime,omitempty" json:"modifiedTime,omitempty"`

	TgFileID           string `bson:"tgFileId,omitempty" json:"tgFileId,omitempty"`
	TgChannelMessageID int    `bson:"tgChannelMessageId,omitempty" json:"tgChannelMessageId,omitempty"`

	Name string `bson:"name,omitempty" json:"name,omitempty"`
	Key  Key    `bson:"key,omitempty" json:"key,omitempty"`
	BPM  string `bson:"bpm,omitempty" json:"bpm,omitempty"`
	Time string `bson:"time,omitempty" json:"time,omitempty"`

	WebViewLink string `bson:"webViewLink,omitempty" json:"webViewLink,omitempty"`
}

func (s *Song) GetTgFileID() string {
	if s.AltPDF != nil {
		return s.AltPDF.TgFileID
	}
	return s.PDF.TgFileID
}

func (s *Song) SetTgFileID(tgFileID string) {
	if s.AltPDF != nil {
		s.AltPDF.TgFileID = tgFileID
		if s.AltPDFs == nil {
			s.AltPDFs = make(map[Key]AltPDF)
		}
		s.AltPDFs[s.AltPDF.Key] = *s.AltPDF
		return
	}
	s.PDF.TgFileID = tgFileID
}

func (s *Song) GetDriveFileID() string {
	if s.AltPDF != nil {
		return s.AltPDF.DriveFileID
	}
	return s.DriveFileID
}

func (s *Song) Meta() string {
	key := s.PDF.Key
	if s.AltPDF != nil {
		key = s.AltPDF.Key
	}
	return fmt.Sprintf("%s, %s, %s", key, s.PDF.BPM, s.PDF.Time)
}

func (s *Song) Caption() string {
	caption := fmt.Sprintf("%s, %s", s.Meta(), strings.Join(s.Tags, ", "))
	return strings.Trim(caption, ", ")
}

type SongTagFrequencies struct {
	Tag   string `bson:"_id"`
	Count int    `bson:"count"`
}

type SongWithEvents struct {
	Song `bson:",inline"`

	Events []*Event `bson:"events,omitempty"`
}

func (s *SongWithEvents) Stats(lang string) string {
	if len(s.Events) == 0 {
		return "-, 0"
	}
	t, _ := lctime.StrftimeLoc(util.IetfToIsoLangCode(lang), "%d %b", s.Events[0].GetLocalTime())
	return fmt.Sprintf("%v, %d", t, len(s.Events))
}

func (s *SongWithEvents) StatsForCaption(periodText string, lang string) string {
	t := txt.GetTranslator(lang)

	if len(s.Events) == 0 {
		return txt.Get("text.noStatsCaption", lang)
	}

	text := txt.Get("text.statsCaption", lang, len(s.Events), periodText, t.FmtDateLong(s.Events[0].GetLocalTime()), s.Events[0].Name)
	return text
}
