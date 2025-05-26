package controller

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/keyboard"
	"github.com/joeyave/scala-bot/service"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"regexp"
	"strconv"
	"time"
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

// case helpers.Members:
//
//	users, err := h.userService.FindManyExtraByBandID(user.BandID)
//	if err != nil {
//		return err
//	}
//
//	usersStr := ""
//	event, err := h.eventService.FindOneOldestByBandID(user.BandID)
//	if err == nil {
//		usersStr = fmt.Sprintf("Статистика ведется с %s", lctime.Strftime("%d %B, %Y", event.Time))
//	}
//
//	for _, user := range users {
//		if user.User == nil || user.User.Name == "" {
//			continue
//		}
//
//		usersStr = fmt.Sprintf("%s\n\n<b><a href=\"tg://user?id=%d\">%s</a></b>\nВсего участий: %d", usersStr, user.User.ID, user.User.Name, len(user.Events))
//
//		if len(user.Events) > 0 {
//			usersStr = fmt.Sprintf("%s\nИз них:", usersStr)
//		}
//
//		mp := map[entities.Role]int{}
//
//		for _, event := range user.Events {
//			for _, membership := range event.Memberships {
//				if membership.UserID == user.User.ID {
//					mp[*membership.Role]++
//					break
//				}
//			}
//		}
//
//		for role, num := range mp {
//			usersStr = fmt.Sprintf("%s\n - %s: %d", usersStr, role.Name, num)
//		}
//	}
//
//	return c.Send(usersStr, telebot.ModeHTML)

func (h *WebAppController) Statistics(ctx *gin.Context) {

	fmt.Println(ctx.Request.URL.String())

	hex := ctx.Query("bandId")
	bandID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return
	}

	band, err := h.BandService.FindOneByID(bandID)
	if err != nil {
		return
	}

	ctx.HTML(http.StatusOK, "statistics.go.html", gin.H{
		"Lang": ctx.Query("lang"),

		//"Users": viewUsers,
		"FromDate": time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.Local),
		"BandID":   bandID.Hex(),
		"Roles":    band.Roles,
	})
}

func (h *WebAppController) CreateEvent(ctx *gin.Context) {

	fmt.Println(ctx.Request.URL.String())
	hex := ctx.Query("bandId")
	bandID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return
	}

	band, err := h.BandService.FindOneByID(bandID)
	if err != nil {
		return
	}

	event := &entity.Event{
		Time:   time.Now(),
		BandID: bandID,
		Band:   band,
	}
	eventNames, err := h.EventService.GetMostFrequentEventNames(bandID, 4)
	if err != nil {
		return
	}

	ctx.HTML(http.StatusOK, "event.go.html", gin.H{
		"EventNames": eventNames,
		"Event":      event,
		"Action":     "create",
		"Lang":       ctx.Query("lang"),
	})
}

func (h *WebAppController) EditEvent(ctx *gin.Context) {

	fmt.Println(ctx.Request.RequestURI)

	hex := ctx.Param("id")
	eventID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return
	}

	messageID := ctx.Query("messageId")
	chatID := ctx.Query("chatId")
	userID := ctx.Query("userId")

	event, err := h.EventService.FindOneByID(eventID)
	if err != nil {
		return
	}

	eventNames, err := h.EventService.GetMostFrequentEventNames(event.BandID, 4)
	if err != nil {
		return
	}

	ctx.HTML(http.StatusOK, "event.go.html", gin.H{
		"Action": "edit",

		"MessageID": messageID,
		"ChatID":    chatID,
		"UserID":    userID,

		"Event": event,

		"EventNames": eventNames,

		"Lang": ctx.Query("lang"),
	})
}

func (h *WebAppController) EditEventConfirm(ctx *gin.Context) {

	hex := ctx.Param("id")
	eventID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return
	}

	messageIDStr := ctx.Query("messageId")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		return
	}
	chatIDStr := ctx.Query("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return
	}
	userIDStr := ctx.Query("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return
	}

	var event *entity.Event
	err = ctx.ShouldBindJSON(&event)
	if err != nil {
		return
	}
	event.ID = eventID

	updatedEvent, err := h.EventService.UpdateOne(*event)
	if err != nil {
		return
	}

	html := h.EventService.ToHtmlStringByEvent(*updatedEvent, "ru")

	user, err := h.UserService.FindOneByID(userID)
	if err != nil {
		return
	}

	markup := gotgbot.InlineKeyboardMarkup{}
	markup.InlineKeyboard = keyboard.EventEdit(event, user, chatID, messageID, "ru")

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
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *WebAppController) CreateSong(ctx *gin.Context) {

	fmt.Println(ctx.Request.URL.String())
	hex := ctx.Query("userId")
	userID, err := strconv.ParseInt(hex, 10, 64)
	if err != nil {
		return
	}

	user, err := h.UserService.FindOneByID(userID)
	if err != nil {
		return
	}

	allTags, err := h.SongService.GetTags(user.BandID)
	if err != nil {
		return
	}

	var songTags []*SelectEntity
	for _, tag := range allTags {
		songTags = append(songTags, &SelectEntity{Name: tag, IsSelected: false})
	}

	ctx.HTML(http.StatusOK, "song.go.html", gin.H{
		"Action": "create",
		"BPMs":   valuesForSelect("?", bpms, "BPM"),
		"Times":  valuesForSelect("?", times, "Time"),
		"Tags":   songTags,

		"Lang": ctx.Query("lang"),
	})
}

var times = []string{"4/4", "3/4", "6/8", "2/2"}
var bpms []string

func init() {
	for i := 60; i < 180; i++ {
		bpms = append(bpms, strconv.Itoa(i))
	}
}

type SelectEntity struct {
	Name       string
	Value      string
	IsSelected bool
}

func valuesForSelect(songVal string, values []string, name string) []*SelectEntity {
	keysForSelect := []*SelectEntity{
		{
			Name:       name,
			Value:      "?",
			IsSelected: false,
		},
	}

	somethingWasSelected := false
	for _, key := range values {
		if key == songVal {
			somethingWasSelected = true
		}
		keysForSelect = append(keysForSelect, &SelectEntity{Name: key, Value: key, IsSelected: key == songVal})
	}

	if !somethingWasSelected && songVal == "" || songVal == "?" {
		keysForSelect[0].IsSelected = true
	} else if !somethingWasSelected {
		keysForSelect = append(keysForSelect, &SelectEntity{Name: songVal, Value: songVal, IsSelected: true})
	}

	return keysForSelect
}

// API Users.

type User struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Events []*Event `json:"events"`
}

type Event struct {
	ID      primitive.ObjectID `json:"id"`
	Date    string             `json:"date"`
	Weekday time.Weekday       `json:"weekday"`
	Name    string             `json:"name"`
	Roles   []*Role            `json:"roles"`
}

type Role struct {
	ID   primitive.ObjectID `json:"id"`
	Name string             `json:"name"`
}

func (h *WebAppController) UsersWithEvents(ctx *gin.Context) {

	fmt.Println(ctx.Request.URL)

	hex := ctx.Query("bandId")
	bandID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	from := ctx.Query("from")
	fromDate, err := time.Parse("02.01.2006", from)
	if err != nil {
		fromDate = time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.Local)
	}

	users, err := h.UserService.FindManyExtraByBandID(bandID, fromDate, time.Now())
	if err != nil {
		return
	}

	var viewUsers []*User
	for _, user := range users {
		viewUser := &User{
			ID:   user.ID,
			Name: user.Name,
		}

		for _, event := range user.Events {
			viewEvent := &Event{
				ID:      event.ID,
				Date:    event.Time.Format("2006-01-02"),
				Weekday: event.Time.Weekday(),
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

type Song struct {
	LyricsHTML     string `json:"lyricsHtml"`
	SectionsNumber int    `json:"sectionsNumber"`
}

func (h *WebAppController) SongData(ctx *gin.Context) {

	songIDFromQ := ctx.Param("id")
	songID, err := primitive.ObjectIDFromHex(songIDFromQ)
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
	songID, err := primitive.ObjectIDFromHex(songIDFromQ)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	song, err := h.SongService.FindOneByID(songID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	songLyricsHTML, sectionsNumber := h.DriveFileService.GetHTMLTextWithSectionsNumber(song.DriveFileID)

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"lyricsHtml":     songLyricsHTML,
			"sectionsNumber": sectionsNumber,
		},
	})
}

type EditSongData struct {
	Name             string   `json:"name"`
	Key              string   `json:"key"`
	BPM              string   `json:"bpm"`
	Time             string   `json:"time"`
	Tags             []string `json:"tags"`
	TransposeSection string   `json:"transposeSection"`
}

var keyRegex = regexp.MustCompile(`(?i)key:(.*?);`)
var bpmRegex = regexp.MustCompile(`(?i)bpm:(.*?);`)
var timeRegex = regexp.MustCompile(`(?i)time:(.*?);`)

func (h *WebAppController) SongEdit(ctx *gin.Context) {

	songIDStr := ctx.Param("id")
	songID, err := primitive.ObjectIDFromHex(songIDStr)
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

	if song.PDF.Name != data.Name {
		err := h.DriveFileService.Rename(song.DriveFileID, data.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Error().Err(err).Msgf("Error:")
			return
		}
		song.PDF.Name = data.Name
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
			song.PDF.Key = data.Key

			_, err = h.DriveFileService.TransposeHeader(song.DriveFileID, data.Key, 0)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Error().Err(err).Msgf("Error:")
				return
			}
		}
	}

	if song.PDF.BPM != data.BPM {
		h.DriveFileService.ReplaceAllTextByRegex(song.DriveFileID, bpmRegex, fmt.Sprintf("BPM: %s;", data.BPM))
		song.PDF.BPM = data.BPM
	}

	if song.PDF.Time != data.Time {
		h.DriveFileService.ReplaceAllTextByRegex(song.DriveFileID, timeRegex, fmt.Sprintf("TIME: %s;", data.Time))
		song.PDF.Time = data.Time
	}

	fakeTime, _ := time.Parse("2006", "2006")
	song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)

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
		InlineKeyboard: keyboard.SongInit(song, user, chatID, messageID, "ru"),
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
	songID, err := primitive.ObjectIDFromHex(songIDStr)
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
