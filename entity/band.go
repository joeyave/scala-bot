package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Band struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string        `bson:"name,omitempty" json:"name,omitempty"`
	DriveFolderID   string        `bson:"driveFolderId,omitempty" json:"driveFolderId,omitempty"`
	ArchiveFolderID string        `bson:"archiveFolderId,omitempty" json:"archiveFolderId,omitempty"`
	TempFolderID    string        `bson:"tempFolderID,omitempty" json:"tempFolderID,omitempty"`
	Roles           []*Role       `bson:"roles,omitempty" json:"roles,omitempty"`
	Timezone        string        `bson:"timezone,omitempty" json:"timezone,omitempty"`
	AdminUserIDs    []int64       `bson:"adminUserIds,omitempty" json:"adminUserIds,omitempty"`
}

func (b *Band) GetLocation() *time.Location {
	loc, err := time.LoadLocation(b.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (b *Band) GetNowTime() time.Time {
	return time.Now().In(b.GetLocation())
}

func (b *Band) IsBandAdmin(userID int64) bool {
	for _, adminUserID := range b.AdminUserIDs {
		if adminUserID == userID {
			return true
		}
	}
	return false
}

func (b *Band) AddAdminUserID(userID int64) {
	if b.IsBandAdmin(userID) {
		return
	}
	b.AdminUserIDs = append(b.AdminUserIDs, userID)
}

func (b *Band) RemoveAdminUserID(userID int64) {
	adminUserIDs := make([]int64, 0, len(b.AdminUserIDs))
	for _, adminUserID := range b.AdminUserIDs {
		if adminUserID != userID {
			adminUserIDs = append(adminUserIDs, adminUserID)
		}
	}
	b.AdminUserIDs = adminUserIDs
}
