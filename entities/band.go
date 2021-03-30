package entities

import (
	"fmt"
	"github.com/klauspost/lctime"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Band struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name,omitempty"`
	DriveFolderID string             `bson:"driveFolderId,omitempty"`

	Roles []*Role `bson:"roles,omitempty"`

	NotionCollection *NotionCollection `bson:"notionCollection"`
}

// TODO: refactor.
type NotionEvent struct {
	ID              string    `bson:"id"`
	Name            string    `bson:"name"`
	Time            time.Time `bson:"date"`
	SetlistPageIDs  []string  `bson:"setlistPageIds"`
	BackVocalistIDs []string  `bson:"vocalistIds"`
	LeadVocalistIDs []string  `bson:"leadVocalistIds"`
}

func (e *NotionEvent) GetAlias() string {
	err := lctime.SetLocale("ru_RU")
	if err != nil {
		fmt.Println(err)
	}

	timeStr := lctime.Strftime("%A / %d %b", e.Time)

	return fmt.Sprintf("%s / %s", timeStr, e.Name)
}

type NotionCollection struct {
	NotionCollectionID     string `bson:"notionCollectionId"`
	NotionCollectionViewID string `bson:"notionCollectionViewId"`
}
