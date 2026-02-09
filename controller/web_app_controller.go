package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/service"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type WebAppController struct {
	Bot               *gotgbot.Bot
	EventService      *service.EventService
	UserService       *service.UserService
	BandService       *service.BandService
	DriveFileService  *service.DriveFileService
	SongService       *service.SongService
	VoiceService      *service.VoiceService
	MembershipService *service.MembershipService
	RoleService       *service.RoleService
}

func (h *WebAppController) Statistics(ctx *gin.Context) {
	fmt.Println(ctx.Request.URL.String())

	hex := ctx.Query("bandId")
	bandID, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	band, err := h.BandService.FindOneByID(bandID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	now := band.GetNowTime()

	ctx.HTML(http.StatusOK, "statistics.go.html", gin.H{
		"Lang": ctx.Query("lang"),

		// "Users": viewUsers,
		"FromDate": time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()),
		"BandID":   bandID.Hex(),
		"Roles":    band.Roles,
	})
}

// API Users.

type User struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Events []*Event `json:"events"`
}

type Event struct {
	ID      bson.ObjectID `json:"id"`
	Date    string        `json:"date"`
	Weekday time.Weekday  `json:"weekday"`
	Name    string        `json:"name"`
	Roles   []*Role       `json:"roles"`
}

type Role struct {
	ID   bson.ObjectID `json:"id"`
	Name string        `json:"name"`
}

func (h *WebAppController) UsersWithEvents(ctx *gin.Context) {
	fmt.Println(ctx.Request.URL)

	hex := ctx.Query("bandId")
	bandID, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	band, err := h.BandService.FindOneByID(bandID)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	now := band.GetNowTime()

	from := ctx.Query("from")
	fromDate, err := time.ParseInLocation("02.01.2006", from, band.GetLocation())
	if err != nil {
		fromDate = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.Local)
	}

	users, err := h.UserService.FindManyExtraByBandID(bandID, fromDate, now)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	viewUsers := make([]*User, 0)
	for _, user := range users {
		viewUser := &User{
			ID:   user.ID,
			Name: user.Name,
		}

		for _, event := range user.Events {
			viewEvent := &Event{
				ID:      event.ID,
				Date:    event.TimeUTC.In(band.GetLocation()).Format("2006-01-02"),
				Weekday: event.TimeUTC.In(band.GetLocation()).Weekday(),
				Name:    event.Name,
			}

			for _, membership := range event.Memberships {
				if membership.UserID == user.ID {
					viewRole := &Role{
						ID:   membership.Role.ID,
						Name: membership.Role.Name,
					}
					viewEvent.Roles = append(viewEvent.Roles, viewRole)
					break
				}
			}

			viewUser.Events = append(viewUser.Events, viewEvent)
		}

		viewUsers = append(viewUsers, viewUser)
	}

	ctx.JSON(200, gin.H{
		"users": viewUsers,
	})
}

// API Songs.

func (h *WebAppController) SongData(ctx *gin.Context) {
	songIDFromQ := ctx.Param("id")
	songID, err := bson.ObjectIDFromHex(songIDFromQ)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	songEntity, err := h.SongService.FindOneByID(songID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userIDFromQ := ctx.Query("userId")
	userID, err := strconv.ParseInt(userIDFromQ, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := h.UserService.FindOneByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	allTags, err := h.SongService.GetTags(user.BandID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"song":     songEntity,
			"bandTags": allTags,
		},
	})
}

func (h *WebAppController) SongLyrics(ctx *gin.Context) {
	songIDFromQ := ctx.Param("id")
	songID, err := bson.ObjectIDFromHex(songIDFromQ)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	song, err := h.SongService.FindOneByID(songID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	songLyricsHTML, sectionsNumber, md, err := h.DriveFileService.GetHTMLTextWithSectionsNumberAndMetadata(song.DriveFileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metadataSyncWasUpdated := false
	if song.PDF.Name != md.Title || song.PDF.Key != md.Key || song.PDF.BPM != md.BPM || song.PDF.Time != md.Time {
		song.PDF.Name = md.Title
		song.PDF.Key = md.Key
		song.PDF.BPM = md.BPM
		song.PDF.Time = md.Time
		song, err = h.SongService.UpdateOne(*song)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		metadataSyncWasUpdated = true
	}
	ctx.Header("X-Metadata-Sync-Updated", strconv.FormatBool(metadataSyncWasUpdated))

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"lyricsHtml":             songLyricsHTML,
			"sectionsNumber":         sectionsNumber,
			"metadataSyncWasUpdated": metadataSyncWasUpdated,
			"metadata": gin.H{
				"name": md.Title,
				"key":  md.Key,
				"bpm":  md.BPM,
				"time": md.Time,
			},
		},
	})
}

type EditSongData struct {
	Name             string     `json:"name"`
	Key              entity.Key `json:"key"`
	BPM              string     `json:"bpm"`
	Time             string     `json:"time"`
	Tags             []string   `json:"tags"`
	TransposeSection string     `json:"transposeSection"`
}

func (h *WebAppController) SongEdit(ctx *gin.Context) {
	songIDStr := ctx.Param("id")
	songID, err := bson.ObjectIDFromHex(songIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	userIDStr := ctx.Query("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	messageIDStr := ctx.Query("messageId")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}
	chatIDStr := ctx.Query("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	var data *EditSongData
	err = ctx.ShouldBindJSON(&data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	song, err := h.SongService.FindOneByID(songID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	song.Tags = data.Tags
	nameChanged := song.PDF.Name != data.Name
	bpmChanged := song.PDF.BPM != data.BPM
	timeChanged := song.PDF.Time != data.Time

	if nameChanged {
		err := h.DriveFileService.Rename(song.DriveFileID, data.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Error().Err(err).Msgf("Error:")
			return
		}
		song.PDF.Name = data.Name
	}

	if err := h.DriveFileService.NormalizeMetadataLayout(song.DriveFileID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	if song.PDF.Key != data.Key {
		if data.TransposeSection != "" {
			section, err := strconv.Atoi(data.TransposeSection)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Error().Err(err).Msgf("Error:")
				return
			}
			_, err = h.DriveFileService.TransposeOne(song.DriveFileID, data.Key, section)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Error().Err(err).Msgf("Error:")
				return
			}
			if section == 0 {
				song.PDF.Key = data.Key
			}
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "transposeSection is required when key is changed"})
			return
		}
	}

	metadataPatch := service.MetadataPatch{}
	metadataChanged := false
	if nameChanged {
		metadataPatch.Title = &data.Name
		metadataChanged = true
	}

	if bpmChanged {
		song.PDF.BPM = data.BPM
		metadataPatch.BPM = &data.BPM
		metadataChanged = true
	}

	if timeChanged {
		song.PDF.Time = data.Time
		metadataPatch.Time = &data.Time
		metadataChanged = true
	}

	if metadataChanged {
		if err := h.DriveFileService.UpdateMetadataAcrossSections(song.DriveFileID, metadataPatch); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Error().Err(err).Msgf("Error:")
			return
		}
	}

	// Keep Mongo metadata aligned with the current document metadata after normalize/transpose.
	song.PDF.Key, song.PDF.BPM, song.PDF.Time = h.DriveFileService.GetMetadata(song.DriveFileID)

	// Force refresh of stored Drive metadata token after in-place document edits.
	song.PDF.Version = 0

	song, err = h.SongService.UpdateOne(*song)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}
	user, err := h.UserService.FindOneByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}
	user.CallbackCache = entity.CallbackCache{
		ChatID:    chatID,
		MessageID: messageID,
		UserID:    userID,
	}
	caption := user.CallbackCache.AddToText(song.Caption())

	reader, err := h.DriveFileService.DownloadOneByID(song.DriveFileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	defer reader.Close()

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.SongInit(song, user, chatID, messageID, user.LanguageCode),
	}

	_, _, err = h.Bot.EditMessageMedia(gotgbot.InputMediaDocument{
		Media:     gotgbot.InputFileByReader(fmt.Sprintf("%s.pdf", song.PDF.Name), reader),
		Caption:   caption,
		ParseMode: "HTML",
	}, &gotgbot.EditMessageMediaOpts{
		ChatId:      chatID,
		MessageId:   messageID,
		ReplyMarkup: markup,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func (h *WebAppController) SongFormat(ctx *gin.Context) {
	songIDStr := ctx.Param("id")
	songID, err := bson.ObjectIDFromHex(songIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	userIDStr := ctx.Query("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	messageIDStr := ctx.Query("messageId")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	chatIDStr := ctx.Query("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	song, err := h.SongService.FindOneByID(songID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	user, err := h.UserService.FindOneByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	driveFile, err := h.DriveFileService.StyleOne(song.DriveFileID, user.LanguageCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	song, _, err = h.SongService.SyncPDFMetadataByDriveFileID(driveFile.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	user.CallbackCache = entity.CallbackCache{
		ChatID:    chatID,
		MessageID: messageID,
		UserID:    userID,
	}
	caption := user.CallbackCache.AddToText(song.Caption())

	reader, err := h.DriveFileService.DownloadOneByID(song.DriveFileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}
	defer reader.Close()

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard.SongInit(song, user, chatID, messageID, user.LanguageCode),
	}

	_, _, err = h.Bot.EditMessageMedia(gotgbot.InputMediaDocument{
		Media:     gotgbot.InputFileByReader(fmt.Sprintf("%s.pdf", song.PDF.Name), reader),
		Caption:   caption,
		ParseMode: "HTML",
	}, &gotgbot.EditMessageMediaOpts{
		ChatId:      chatID,
		MessageId:   messageID,
		ReplyMarkup: markup,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func (h *WebAppController) SongDownload(ctx *gin.Context) {
	songIDStr := ctx.Param("id")
	songID, err := bson.ObjectIDFromHex(songIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	song, err := h.SongService.FindOneByID(songID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	resp, err := h.DriveFileService.DownloadOneByIDWithResp(song.DriveFileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/pdf"
	}

	ctx.DataFromReader(
		http.StatusOK,
		resp.ContentLength,
		contentType,
		resp.Body,
		map[string]string{
			"Content-Disposition": `inline"`,
		},
	)
}

// API tags.

func (h *WebAppController) Tags(ctx *gin.Context) {
	bandIdFromQ := ctx.Query("bandId")
	bandID, err := bson.ObjectIDFromHex(bandIdFromQ)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allTags, err := h.SongService.GetTags(bandID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"tags": allTags}})
}

func (h *WebAppController) EventData(ctx *gin.Context) {
	eventIDFromCtx := ctx.Param("id")
	eventID, err := bson.ObjectIDFromHex(eventIDFromCtx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventEntity, err := h.EventService.FindOneByID(eventID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"event": eventEntity,
		},
	})
}

func (h *WebAppController) FrequentEventNames(ctx *gin.Context) {
	bandIdFromQ := ctx.Query("bandId")
	bandID, err := bson.ObjectIDFromHex(bandIdFromQ)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	namesWithFreq, err := h.EventService.GetMostFrequentEventNames(bandID, 5)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	names := make([]string, 0)
	for _, n := range namesWithFreq {
		names = append(names, n.Name)
	}

	ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"names": names}})
}

// SongOverridesData represents a song in the setlist with optional key override from the frontend.
type SongOverridesData struct {
	SongID   string     `json:"songId"`
	EventKey entity.Key `json:"eventKey,omitempty"`
}

type EditEventData struct {
	Name          string              `json:"name"`
	Date          string              `json:"date"`
	Timezone      string              `json:"timezone"`
	SongIDs       []string            `json:"songIds"`
	SongOverrides []SongOverridesData `json:"songOverrides"`
	Notes         string              `json:"notes"`
}

func (d *EditEventData) GetSongOverride(songID string) *SongOverridesData {
	for _, item := range d.SongOverrides {
		if item.SongID == songID {
			return &item
		}
	}
	return nil
}

func (h *WebAppController) EventEdit(ctx *gin.Context) {
	eventIDStr := ctx.Param("id")
	eventID, err := bson.ObjectIDFromHex(eventIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	userIDStr := ctx.Query("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	messageIDStr := ctx.Query("messageId")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	chatIDStr := ctx.Query("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	var data *EditEventData
	err = ctx.ShouldBindJSON(&data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	event, err := h.EventService.FindOneByID(eventID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	event.Name = data.Name

	loc, err := time.LoadLocation(data.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// todo: format.
	eventDate, err := time.ParseInLocation("2006-01-02T15:04", data.Date, loc)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}
	event.TimeUTC = eventDate.UTC()

	songIDs := make([]bson.ObjectID, 0)
	songOverrides := make([]entity.SongOverride, 0)
	for _, songIDHex := range data.SongIDs {
		songID, err := bson.ObjectIDFromHex(songIDHex)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Error().Err(err).Msgf("Error:")
			return
		}
		songIDs = append(songIDs, songID)

		override := data.GetSongOverride(songIDHex)
		if override != nil && override.EventKey != "" {
			songOverrides = append(songOverrides, entity.SongOverride{
				SongID:   songID,
				EventKey: override.EventKey,
			})
		}
	}
	event.SongIDs = songIDs
	event.SongOverrides = songOverrides

	event.Notes = &data.Notes

	event, err = h.EventService.UpdateOne(*event)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	user, err := h.UserService.FindOneByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	markup := gotgbot.InlineKeyboardMarkup{}
	markup.InlineKeyboard = keyboard.EventEdit(event, user, chatID, messageID, user.LanguageCode)

	// call retrieveFreshSongsForEvent here.

	songs, err := h.SongService.RetrieveFreshSongsForEvent(event)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	html := service.HTMLStringForEvent(*event, songs, user.LanguageCode)

	user.CallbackCache = entity.CallbackCache{
		MessageID: messageID,
		ChatID:    chatID,
	}
	text := user.CallbackCache.AddToText(html)

	_, _, err = h.Bot.EditMessageText(text, &gotgbot.EditMessageTextOpts{
		ChatId:    chatID,
		MessageId: messageID,
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Error().Err(err).Msgf("Error:")
		return
	}

	ctx.Status(http.StatusOK)
}
