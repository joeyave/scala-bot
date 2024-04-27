package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/service"
)

type DriveFileController struct {
	DriveFileService *service.DriveFileService
	SongService      *service.SongService
}

func (c *DriveFileController) Search(ctx *gin.Context) {
	query := ctx.Query("q")
	folderID := ctx.Query("driveFolderId")
	archiveFolderID := ctx.Query("archiveFolderId")

	driveFiles, _, err := c.DriveFileService.FindSomeByFullTextAndFolderID(query, []string{folderID, archiveFolderID}, "")
	if err != nil {
		return
	}

	ctx.JSON(200, gin.H{
		"results": driveFiles,
	})
}

func (c *DriveFileController) FindByDriveFileID(ctx *gin.Context) {
	driveFileID := ctx.Query("driveFileId")

	song, _, err := c.SongService.FindOrCreateOneByDriveFileID(driveFileID)
	if err != nil {
		return
	}

	ctx.JSON(200, gin.H{
		"song": song,
	})
}
