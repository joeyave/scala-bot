package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"time"
)

type Song struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	DriveFileID string      `bson:"driveFileId,omitempty"`
	DriveFile   *drive.File `bson:"-"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`

	PDF PDF `bson:"pdf,omitempty"`

	Voices []*Voice `bson:"voices,omitempty"`
}

type PDF struct {
	TgFileID           string `bson:"tgFileId,omitempty"`
	ModifiedTime       string `bson:"modifiedTime,omitempty"`
	TgChannelMessageID int    `bson:"tgChannelMessageId,omitempty"`
}

func (s *Song) HasOutdatedPDF() bool {
	if s.PDF.ModifiedTime == "" || s.DriveFile == nil {
		return true
	}

	pdfModifiedTime, err := time.Parse(time.RFC3339, s.PDF.ModifiedTime)
	if err != nil {
		return true
	}

	driveFileModifiedTime, err := time.Parse(time.RFC3339, s.DriveFile.ModifiedTime)
	if err != nil {
		return true
	}

	if driveFileModifiedTime.After(pdfModifiedTime) {
		return true
	}

	return false
}
