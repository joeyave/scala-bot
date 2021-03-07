package entities

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Band struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name,omitempty"`
	DriveFolderID string             `bson:"driveFolderId,omitempty"`
	AdminUserIDs  []int64            `bson:"adminUserIds,omitempty"`

	Events           []*Event          `bson:"-"`
	NotionCollection *NotionCollection `bson:"notionCollection"`
}

type Event struct {
	ID      string    `bson:"id"`
	Name    string    `bson:"name"`
	Time    time.Time `bson:"date"`
	Setlist []string  `bson:"setlist"`
}

func (e *Event) GetAlias() string {
	// TODO: move this to the constants
	var weekday string
	switch e.Time.Weekday() {
	case 0:
		weekday = "Воскресенье"
	case 1:
		weekday = "Понедельник"
	case 2:
		weekday = "Вторник"
	case 3:
		weekday = "Среда"
	case 4:
		weekday = "Черверг"
	case 5:
		weekday = "Пятница"
	case 6:
		weekday = "Суббота"
	case 7:
		weekday = "Воскресенье"
	}

	return fmt.Sprintf("%s / %s / %s", e.Time.Format("2006-01-02 15:04"), weekday, e.Name)
}

type NotionCollection struct {
	NotionCollectionID     string `bson:"notionCollectionId"`
	NotionCollectionViewID string `bson:"notionCollectionViewId"`
}
