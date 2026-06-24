package controller

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/repository"
	"github.com/joeyave/scala-bot/service"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type SettingsUserResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	LanguageCode string `json:"languageCode,omitempty"`
	ActiveBandID string `json:"activeBandId,omitempty"`
}

type SettingsBandResponse struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	DriveFolderID         string `json:"driveFolderId"`
	ArchiveFolderID       string `json:"archiveFolderId"`
	TempFolderID          string `json:"tempFolderId"`
	Timezone              string `json:"timezone"`
	IsMember              bool   `json:"isMember"`
	IsActive              bool   `json:"isActive"`
	IsAdmin               bool   `json:"isAdmin"`
	HasPendingJoinRequest bool   `json:"hasPendingJoinRequest"`
}

type SettingsMemberResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsAdmin  bool   `json:"isAdmin"`
	IsSelf   bool   `json:"isSelf"`
	IsActive bool   `json:"isActive"`
}

type SettingsJoinRequestResponse struct {
	ID        string `json:"id"`
	BandID    string `json:"bandId"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type SettingsMeResponse struct {
	User  SettingsUserResponse   `json:"user"`
	Bands []SettingsBandResponse `json:"bands"`
}

type SettingsBandsResponse struct {
	Bands []SettingsBandResponse `json:"bands"`
}

type SettingsMembersResponse struct {
	Members []SettingsMemberResponse `json:"members"`
}

type SettingsJoinRequestCreateResponse struct {
	Request SettingsJoinRequestResponse `json:"request"`
	Created bool                        `json:"created"`
}

type settingsBandRequest struct {
	Name           *string `json:"name"`
	DriveFolderID  *string `json:"driveFolderId"`
	DriveFolderURL *string `json:"driveFolderUrl"`
	Timezone       *string `json:"timezone"`
}

type settingsMemberRoleRequest struct {
	IsAdmin bool `json:"isAdmin"`
}

var (
	driveFolderURLRe = regexp.MustCompile(`(?:/folders/|id=)([a-zA-Z0-9_-]+)`)
	driveFolderIDRe  = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func (h *WebAppController) SettingsMe(ctx *gin.Context) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return
	}

	bands, err := h.findBandsByIDs(settingsUserBandIDs(user))
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	pendingRequests, err := h.JoinRequestService.FindPendingByUserID(user.ID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": SettingsMeResponse{
			User:  h.settingsUserResponse(user),
			Bands: h.settingsBandResponses(user, bands, pendingRequests),
		},
	})
}

func (h *WebAppController) SettingsBands(ctx *gin.Context) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return
	}

	bands, err := h.BandService.FindAll()
	if errors.Is(err, repository.ErrNotFound) {
		bands = []*entity.Band{}
	} else if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	pendingRequests, err := h.JoinRequestService.FindPendingByUserID(user.ID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": SettingsBandsResponse{
			Bands: h.settingsBandResponses(user, bands, pendingRequests),
		},
	})
}

func (h *WebAppController) SettingsCreateBand(ctx *gin.Context) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return
	}

	var request settingsBandRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.badSettingsRequest(ctx, "invalid request body")
		return
	}

	name, err := validateSettingsBandName(request.Name)
	if err != nil {
		h.badSettingsRequest(ctx, err.Error())
		return
	}

	driveFolderID, err := validateSettingsDriveFolder(request.DriveFolderID, request.DriveFolderURL)
	if err != nil {
		h.badSettingsRequest(ctx, err.Error())
		return
	}

	timezone, err := validateSettingsTimezone(request.Timezone)
	if err != nil {
		h.badSettingsRequest(ctx, err.Error())
		return
	}

	band, err := h.BandService.UpdateOne(entity.Band{
		Name:          name,
		DriveFolderID: driveFolderID,
		Timezone:      timezone,
		AdminUserIDs:  []int64{user.ID},
	})
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	user, err = h.UserService.AddToBand(user, band.ID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"band": h.settingsBandResponse(user, band, nil),
		},
	})
}

func (h *WebAppController) SettingsSetActiveBand(ctx *gin.Context) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return
	}

	band, ok := h.settingsBandFromPath(ctx)
	if !ok {
		return
	}

	user, err := h.UserService.SetActiveBand(user, band.ID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user": h.settingsUserResponse(user),
			"band": h.settingsBandResponse(user, band, nil),
		},
	})
}

func (h *WebAppController) SettingsCreateJoinRequest(ctx *gin.Context) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return
	}

	band, ok := h.settingsBandFromPath(ctx)
	if !ok {
		return
	}
	if user.BelongsToBand(band.ID) {
		h.settingsConflict(ctx, "user is already a member of this group")
		return
	}

	var adminUserIDs []int64
	if !h.IsTestBotAPI {
		var err error
		adminUserIDs, err = h.adminUserIDsForBand(band)
		if err != nil {
			h.handleSettingsError(ctx, err)
			return
		}
		if len(adminUserIDs) == 0 {
			h.settingsConflict(ctx, "group has no administrators")
			return
		}
	}

	joinRequest, created, err := h.JoinRequestService.Create(service.CreateJoinRequestInput{
		UserID:   user.ID,
		UserName: settingsDisplayName(user),
		Band:     band,
	})
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	if h.IsTestBotAPI {
		joinRequest, user, err = h.JoinRequestService.Approve(joinRequest.ID, user.ID)
		if err != nil {
			h.handleSettingsError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": SettingsJoinRequestCreateResponse{
				Request: settingsJoinRequestResponse(joinRequest),
				Created: created,
			},
		})
		return
	}

	if created {
		h.notifyBandAdminsAboutJoinRequest(joinRequest, adminUserIDs)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": SettingsJoinRequestCreateResponse{
			Request: settingsJoinRequestResponse(joinRequest),
			Created: created,
		},
	})
}

func (h *WebAppController) SettingsCancelJoinRequest(ctx *gin.Context) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return
	}

	band, ok := h.settingsBandFromPath(ctx)
	if !ok {
		return
	}

	joinRequest, err := h.JoinRequestService.Cancel(user.ID, band.ID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"request": settingsJoinRequestResponse(joinRequest),
		},
	})
}

func (h *WebAppController) SettingsUpdateBand(ctx *gin.Context) {
	user, band, ok := h.settingsAdminUserAndBand(ctx)
	if !ok {
		return
	}

	var request settingsBandRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.badSettingsRequest(ctx, "invalid request body")
		return
	}

	if request.Name != nil {
		name, err := validateSettingsBandName(request.Name)
		if err != nil {
			h.badSettingsRequest(ctx, err.Error())
			return
		}
		band.Name = name
	}

	if request.DriveFolderID != nil || request.DriveFolderURL != nil {
		driveFolderID, err := validateSettingsDriveFolder(request.DriveFolderID, request.DriveFolderURL)
		if err != nil {
			h.badSettingsRequest(ctx, err.Error())
			return
		}
		band.DriveFolderID = driveFolderID
	}

	if request.Timezone != nil {
		timezone, err := validateSettingsTimezone(request.Timezone)
		if err != nil {
			h.badSettingsRequest(ctx, err.Error())
			return
		}
		band.Timezone = timezone
	}

	band, err := h.BandService.UpdateOne(*band)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"band": h.settingsBandResponse(user, band, nil),
		},
	})
}

func (h *WebAppController) SettingsLeaveBand(ctx *gin.Context) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return
	}

	band, ok := h.settingsBandFromPath(ctx)
	if !ok {
		return
	}
	if !user.BelongsToBand(band.ID) {
		h.handleSettingsError(ctx, service.ErrForbidden)
		return
	}
	if len(settingsUserBandIDs(user)) <= 1 {
		h.handleSettingsError(ctx, service.ErrInvalidOperation)
		return
	}

	adminUserIDs, err := h.adminUserIDsForBand(band)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}
	if containsInt64(adminUserIDs, user.ID) {
		if len(adminUserIDs) <= 1 {
			h.handleSettingsError(ctx, service.ErrInvalidOperation)
			return
		}
		band.AdminUserIDs = removeInt64(adminUserIDs, user.ID)
		if _, err := h.BandService.UpdateOne(*band); err != nil {
			h.handleSettingsError(ctx, err)
			return
		}
	}

	user, err = h.UserService.RemoveFromBand(user, band.ID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user": h.settingsUserResponse(user),
		},
	})
}

func (h *WebAppController) SettingsBandMembers(ctx *gin.Context) {
	user, band, ok := h.settingsAdminUserAndBand(ctx)
	if !ok {
		return
	}

	members, err := h.UserService.FindMultipleByBandID(band.ID)
	if errors.Is(err, repository.ErrNotFound) {
		members = []*entity.User{}
	} else if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": SettingsMembersResponse{
			Members: h.settingsMemberResponses(user, band, members),
		},
	})
}

func (h *WebAppController) SettingsUpdateBandMember(ctx *gin.Context) {
	user, band, ok := h.settingsAdminUserAndBand(ctx)
	if !ok {
		return
	}

	member, ok := h.settingsMemberFromPath(ctx, band)
	if !ok {
		return
	}

	var request settingsMemberRoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.badSettingsRequest(ctx, "invalid request body")
		return
	}

	adminUserIDs, err := h.adminUserIDsForBand(band)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	if request.IsAdmin {
		band.AdminUserIDs = appendUniqueInt64(adminUserIDs, member.ID)
	} else {
		if member.ID == user.ID {
			h.handleSettingsError(ctx, service.ErrInvalidOperation)
			return
		}
		if containsInt64(adminUserIDs, member.ID) && len(adminUserIDs) <= 1 {
			h.handleSettingsError(ctx, service.ErrInvalidOperation)
			return
		}
		band.AdminUserIDs = removeInt64(adminUserIDs, member.ID)
	}

	band, err = h.BandService.UpdateOne(*band)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"member": h.settingsMemberResponse(member, band, member.ID),
		},
	})
}

func (h *WebAppController) SettingsRemoveBandMember(ctx *gin.Context) {
	_, band, ok := h.settingsAdminUserAndBand(ctx)
	if !ok {
		return
	}

	member, ok := h.settingsMemberFromPath(ctx, band)
	if !ok {
		return
	}

	adminUserIDs, err := h.adminUserIDsForBand(band)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return
	}
	if containsInt64(adminUserIDs, member.ID) {
		if len(adminUserIDs) <= 1 {
			h.handleSettingsError(ctx, service.ErrInvalidOperation)
			return
		}
		band.AdminUserIDs = removeInt64(adminUserIDs, member.ID)
		if _, err := h.BandService.UpdateOne(*band); err != nil {
			h.handleSettingsError(ctx, err)
			return
		}
	}

	if _, err := h.UserService.RemoveFromBand(member, band.ID); err != nil {
		h.handleSettingsError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": gin.H{}})
}

func (h *WebAppController) settingsCurrentUser(ctx *gin.Context) (*entity.User, bool) {
	userID, err := strconv.ParseInt(strings.TrimSpace(ctx.Query("userId")), 10, 64)
	if err != nil || userID <= 0 {
		h.badSettingsRequest(ctx, "userId is required")
		return nil, false
	}

	user, err := h.UserService.FindOneOrCreateByID(userID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return nil, false
	}

	return user, true
}

func (h *WebAppController) settingsAdminUserAndBand(ctx *gin.Context) (*entity.User, *entity.Band, bool) {
	user, ok := h.settingsCurrentUser(ctx)
	if !ok {
		return nil, nil, false
	}

	band, ok := h.settingsBandFromPath(ctx)
	if !ok {
		return nil, nil, false
	}

	if !h.BandService.IsUserAdmin(user, band) {
		h.handleSettingsError(ctx, service.ErrForbidden)
		return nil, nil, false
	}

	return user, band, true
}

func (h *WebAppController) settingsBandFromPath(ctx *gin.Context) (*entity.Band, bool) {
	bandID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		h.badSettingsRequest(ctx, "invalid band id")
		return nil, false
	}

	band, err := h.BandService.FindOneByID(bandID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return nil, false
	}

	return band, true
}

func (h *WebAppController) settingsMemberFromPath(ctx *gin.Context, band *entity.Band) (*entity.User, bool) {
	memberID, err := strconv.ParseInt(ctx.Param("memberId"), 10, 64)
	if err != nil {
		h.badSettingsRequest(ctx, "invalid member id")
		return nil, false
	}

	member, err := h.UserService.FindOneByID(memberID)
	if err != nil {
		h.handleSettingsError(ctx, err)
		return nil, false
	}
	if !member.BelongsToBand(band.ID) {
		h.handleSettingsError(ctx, service.ErrForbidden)
		return nil, false
	}

	return member, true
}

func (h *WebAppController) findBandsByIDs(ids []bson.ObjectID) ([]*entity.Band, error) {
	if len(ids) == 0 {
		return []*entity.Band{}, nil
	}
	bands, err := h.BandService.FindManyByIDs(ids)
	if errors.Is(err, repository.ErrNotFound) {
		return []*entity.Band{}, nil
	}
	return bands, err
}

func (h *WebAppController) settingsUserResponse(user *entity.User) SettingsUserResponse {
	response := SettingsUserResponse{
		ID:           user.ID,
		Name:         settingsDisplayName(user),
		LanguageCode: user.LanguageCode,
	}
	if !user.BandID.IsZero() {
		response.ActiveBandID = user.BandID.Hex()
	}
	return response
}

func settingsDisplayName(user *entity.User) string {
	name := strings.TrimSpace(user.Name)
	if name != "" {
		return name
	}
	return "Telegram user " + strconv.FormatInt(user.ID, 10)
}

func (h *WebAppController) settingsBandResponses(user *entity.User, bands []*entity.Band, pendingRequests []*entity.JoinRequest) []SettingsBandResponse {
	responses := make([]SettingsBandResponse, 0, len(bands))
	pendingByBandID := pendingJoinRequestsByBandID(pendingRequests)
	for _, band := range bands {
		responses = append(responses, h.settingsBandResponse(user, band, pendingByBandID))
	}
	return responses
}

func (h *WebAppController) settingsBandResponse(user *entity.User, band *entity.Band, pendingByBandID map[bson.ObjectID]*entity.JoinRequest) SettingsBandResponse {
	_, hasPendingJoinRequest := pendingByBandID[band.ID]
	return SettingsBandResponse{
		ID:                    band.ID.Hex(),
		Name:                  band.Name,
		DriveFolderID:         band.DriveFolderID,
		ArchiveFolderID:       band.ArchiveFolderID,
		TempFolderID:          band.TempFolderID,
		Timezone:              band.Timezone,
		IsMember:              user.BelongsToBand(band.ID),
		IsActive:              user.BandID == band.ID,
		IsAdmin:               h.BandService.IsUserAdmin(user, band),
		HasPendingJoinRequest: hasPendingJoinRequest,
	}
}

func (h *WebAppController) settingsMemberResponses(currentUser *entity.User, band *entity.Band, members []*entity.User) []SettingsMemberResponse {
	responses := make([]SettingsMemberResponse, 0, len(members))
	for _, member := range members {
		responses = append(responses, h.settingsMemberResponse(member, band, currentUser.ID))
	}
	return responses
}

func (h *WebAppController) settingsMemberResponse(member *entity.User, band *entity.Band, currentUserID int64) SettingsMemberResponse {
	return SettingsMemberResponse{
		ID:       member.ID,
		Name:     member.Name,
		IsAdmin:  h.BandService.IsUserAdmin(member, band),
		IsSelf:   member.ID == currentUserID,
		IsActive: member.BandID == band.ID,
	}
}

func (h *WebAppController) adminUserIDsForBand(band *entity.Band) ([]int64, error) {
	if len(band.AdminUserIDs) > 0 {
		return uniqueInt64s(band.AdminUserIDs), nil
	}

	members, err := h.UserService.FindMultipleByBandID(band.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return []int64{}, nil
	}
	if err != nil {
		return nil, err
	}

	adminUserIDs := make([]int64, 0)
	for _, member := range members {
		if member.IsAdmin() {
			adminUserIDs = appendUniqueInt64(adminUserIDs, member.ID)
		}
	}
	return adminUserIDs, nil
}

func (h *WebAppController) notifyBandAdminsAboutJoinRequest(joinRequest *entity.JoinRequest, adminUserIDs []int64) {
	if h.Bot == nil {
		return
	}

	for _, adminUserID := range adminUserIDs {
		lang := "ru"
		if adminUser, err := h.UserService.FindOneByID(adminUserID); err == nil && adminUser != nil && adminUser.LanguageCode != "" {
			lang = adminUser.LanguageCode
		}
		text := txt.Get(
			"text.joinRequestNotification",
			lang,
			html.EscapeString(joinRequest.UserName),
			html.EscapeString(joinRequest.BandName),
		)

		replyMarkup := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{Text: txt.Get("button.approve", lang), CallbackData: util.CallbackData(state.JoinRequestApprove, joinRequest.ID.Hex())},
					{Text: txt.Get("button.decline", lang), CallbackData: util.CallbackData(state.JoinRequestDecline, joinRequest.ID.Hex())},
				},
			},
		}

		_, err := h.Bot.SendMessage(adminUserID, text, &gotgbot.SendMessageOpts{
			ParseMode:   "HTML",
			ReplyMarkup: replyMarkup,
		})
		if err != nil {
			log.Warn().Err(err).Int64("adminUserID", adminUserID).Str("joinRequestID", joinRequest.ID.Hex()).Msg("failed to notify admin about join request")
		}
	}
}

func (h *WebAppController) handleSettingsError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, service.ErrForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, service.ErrInvalidOperation):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation"})
	case errors.Is(err, service.ErrAlreadyExists):
		ctx.JSON(http.StatusConflict, gin.H{"error": "already exists"})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *WebAppController) badSettingsRequest(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func (h *WebAppController) settingsConflict(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusConflict, gin.H{"error": message})
}

func validateSettingsBandName(rawName *string) (string, error) {
	if rawName == nil {
		return "", fmt.Errorf("name is required")
	}
	name := strings.TrimSpace(*rawName)
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	return name, nil
}

func validateSettingsDriveFolder(rawDriveFolderID, rawDriveFolderURL *string) (string, error) {
	raw := ""
	if rawDriveFolderID != nil {
		raw = *rawDriveFolderID
	}
	if raw == "" && rawDriveFolderURL != nil {
		raw = *rawDriveFolderURL
	}

	driveFolderID, err := parseDriveFolderID(raw)
	if err != nil {
		return "", err
	}
	return driveFolderID, nil
}

func parseDriveFolderID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("drive folder is required")
	}

	if matches := driveFolderURLRe.FindStringSubmatch(raw); len(matches) == 2 {
		return matches[1], nil
	}

	if driveFolderIDRe.MatchString(raw) {
		return raw, nil
	}

	return "", fmt.Errorf("invalid drive folder")
}

func validateSettingsTimezone(rawTimezone *string) (string, error) {
	if rawTimezone == nil {
		return "", fmt.Errorf("timezone is required")
	}
	timezone := strings.TrimSpace(*rawTimezone)
	if timezone == "" {
		return "", fmt.Errorf("timezone is required")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return "", fmt.Errorf("invalid timezone")
	}
	return timezone, nil
}

func pendingJoinRequestsByBandID(requests []*entity.JoinRequest) map[bson.ObjectID]*entity.JoinRequest {
	result := make(map[bson.ObjectID]*entity.JoinRequest, len(requests))
	for _, request := range requests {
		result[request.BandID] = request
	}
	return result
}

func settingsJoinRequestResponse(request *entity.JoinRequest) SettingsJoinRequestResponse {
	return SettingsJoinRequestResponse{
		ID:        request.ID.Hex(),
		BandID:    request.BandID.Hex(),
		Status:    string(request.Status),
		CreatedAt: request.CreatedAt.Format(time.RFC3339),
	}
}

func settingsUserBandIDs(user *entity.User) []bson.ObjectID {
	ids := make([]bson.ObjectID, 0, len(user.BandIDs)+1)
	if !user.BandID.IsZero() {
		ids = appendUniqueBandID(ids, user.BandID)
	}
	for _, bandID := range user.BandIDs {
		if !bandID.IsZero() {
			ids = appendUniqueBandID(ids, bandID)
		}
	}
	return ids
}

func appendUniqueBandID(ids []bson.ObjectID, id bson.ObjectID) []bson.ObjectID {
	for _, existingID := range ids {
		if existingID == id {
			return ids
		}
	}
	return append(ids, id)
}

func uniqueInt64s(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	for _, value := range values {
		result = appendUniqueInt64(result, value)
	}
	return result
}

func appendUniqueInt64(values []int64, value int64) []int64 {
	for _, existingValue := range values {
		if existingValue == value {
			return values
		}
	}
	return append(values, value)
}

func removeInt64(values []int64, value int64) []int64 {
	result := make([]int64, 0, len(values))
	for _, existingValue := range values {
		if existingValue != value {
			result = append(result, existingValue)
		}
	}
	return result
}

func containsInt64(values []int64, value int64) bool {
	for _, existingValue := range values {
		if existingValue == value {
			return true
		}
	}
	return false
}
