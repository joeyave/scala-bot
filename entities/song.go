package entities

import (
	"google.golang.org/api/drive/v3"
	"time"
)

type Song struct {
	ID     string      `bson:"_id,omitempty"`
	File   *drive.File `bson:"file,omitempty"`
	PDF    PDF         `bson:"pdf,omitempty"`
	Voices []*Voice    `bson:"voices,omitempty"`
}

type Voice struct {
	TgFileID string `bson:"tgFileId,omitempty"`
	Caption  string `bson:"caption,omitempty"`
}

type PDF struct {
	TgFileID           string `bson:"tgFileId,omitempty"`
	ModifiedTime       string `bson:"modifiedTime,omitempty"`
	TgChannelMessageID int    `bson:"tgChannelMessageId,omitempty"`
}

func (s *Song) HasOutdatedPDF() bool {
	if s.PDF.ModifiedTime == "" || s.File == nil {
		return true
	}

	pdfModifiedTime, err := time.Parse(time.RFC3339, s.PDF.ModifiedTime)
	driveFileModifiedTime, err := time.Parse(time.RFC3339, s.File.ModifiedTime)

	if err != nil || driveFileModifiedTime.After(pdfModifiedTime) {
		return true
	}
	return false
}

func (s *Song) BelongsToUser(user User) bool {
	for _, userFolderID := range user.GetFolderIDs() {
		for _, songParentID := range s.File.Parents {
			if songParentID == userFolderID {
				return true
			}
		}
	}

	return false
}
