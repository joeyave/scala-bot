package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/service"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type settingsStubUserService struct {
	users map[int64]*entity.User
}

func (s *settingsStubUserService) FindOneByID(id int64) (*entity.User, error) {
	return s.users[id], nil
}

func (s *settingsStubUserService) FindOneOrCreateByID(id int64) (*entity.User, error) {
	if user, ok := s.users[id]; ok {
		return user, nil
	}
	user := &entity.User{ID: id}
	s.users[id] = user
	return user, nil
}

func (s *settingsStubUserService) FindMultipleByBandID(bandID bson.ObjectID) ([]*entity.User, error) {
	users := make([]*entity.User, 0)
	for _, user := range s.users {
		if user.BelongsToBand(bandID) {
			users = append(users, user)
		}
	}
	return users, nil
}

func (s *settingsStubUserService) FindMultipleByIDs(ids []int64) ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		if user, ok := s.users[id]; ok {
			users = append(users, user)
		}
	}
	return users, nil
}

func (s *settingsStubUserService) UpdateOne(user entity.User) (*entity.User, error) {
	userCopy := user
	s.users[user.ID] = &userCopy
	return &userCopy, nil
}

func (s *settingsStubUserService) AddToBand(user *entity.User, bandID bson.ObjectID) (*entity.User, error) {
	if !user.BelongsToBand(bandID) {
		user.BandIDs = append(user.BandIDs, bandID)
	}
	if user.BandID.IsZero() {
		user.BandID = bandID
	}
	return s.UpdateOne(*user)
}

func (s *settingsStubUserService) RemoveFromBand(user *entity.User, bandID bson.ObjectID) (*entity.User, error) {
	nextBandIDs := make([]bson.ObjectID, 0, len(user.BandIDs))
	for _, userBandID := range user.BandIDs {
		if userBandID != bandID {
			nextBandIDs = append(nextBandIDs, userBandID)
		}
	}
	user.BandIDs = nextBandIDs
	if user.BandID == bandID {
		if len(user.BandIDs) > 0 {
			user.BandID = user.BandIDs[0]
		} else {
			user.BandID = bson.NilObjectID
		}
	}
	return s.UpdateOne(*user)
}

func (s *settingsStubUserService) SetActiveBand(user *entity.User, bandID bson.ObjectID) (*entity.User, error) {
	user.BandID = bandID
	return s.UpdateOne(*user)
}

func (s *settingsStubUserService) FindManyExtraByBandID(bson.ObjectID, time.Time, time.Time) ([]*entity.UserWithEvents, error) {
	return nil, nil
}

type settingsStubBandService struct {
	bands map[bson.ObjectID]*entity.Band
}

func (s *settingsStubBandService) FindAll() ([]*entity.Band, error) {
	bands := make([]*entity.Band, 0, len(s.bands))
	for _, band := range s.bands {
		bands = append(bands, band)
	}
	return bands, nil
}

func (s *settingsStubBandService) FindManyByIDs(ids []bson.ObjectID) ([]*entity.Band, error) {
	bands := make([]*entity.Band, 0, len(ids))
	for _, id := range ids {
		if band, ok := s.bands[id]; ok {
			bands = append(bands, band)
		}
	}
	return bands, nil
}

func (s *settingsStubBandService) FindOneByID(id bson.ObjectID) (*entity.Band, error) {
	return s.bands[id], nil
}

func (s *settingsStubBandService) UpdateOne(band entity.Band) (*entity.Band, error) {
	bandCopy := band
	s.bands[band.ID] = &bandCopy
	return &bandCopy, nil
}

func (s *settingsStubBandService) IsUserAdmin(user *entity.User, band *entity.Band) bool {
	return service.NewBandService(nil).IsUserAdmin(user, band)
}

type settingsStubJoinRequestService struct {
	createFunc  func(service.CreateJoinRequestInput) (*entity.JoinRequest, bool, error)
	approveFunc func(bson.ObjectID, int64) (*entity.JoinRequest, *entity.User, error)
	cancelFunc  func(int64, bson.ObjectID) (*entity.JoinRequest, error)
}

func (s settingsStubJoinRequestService) FindOneByID(bson.ObjectID) (*entity.JoinRequest, error) {
	return nil, nil
}

func (s settingsStubJoinRequestService) FindPendingByUserID(int64) ([]*entity.JoinRequest, error) {
	return []*entity.JoinRequest{}, nil
}

func (s settingsStubJoinRequestService) Create(input service.CreateJoinRequestInput) (*entity.JoinRequest, bool, error) {
	if s.createFunc != nil {
		return s.createFunc(input)
	}
	return nil, false, nil
}

func (s settingsStubJoinRequestService) Approve(requestID bson.ObjectID, userID int64) (*entity.JoinRequest, *entity.User, error) {
	if s.approveFunc != nil {
		return s.approveFunc(requestID, userID)
	}
	return nil, nil, nil
}

func (s settingsStubJoinRequestService) Decline(bson.ObjectID, int64) (*entity.JoinRequest, error) {
	return nil, nil
}

func (s settingsStubJoinRequestService) Cancel(userID int64, bandID bson.ObjectID) (*entity.JoinRequest, error) {
	if s.cancelFunc != nil {
		return s.cancelFunc(userID, bandID)
	}
	return nil, nil
}

func TestSettingsMeReturnsUserBands(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	bandID := bson.NewObjectID()
	user := &entity.User{
		ID:           42,
		Name:         "Old Name",
		LanguageCode: "uk",
		BandID:       bandID,
		BandIDs:      []bson.ObjectID{bandID},
	}
	band := &entity.Band{
		ID:            bandID,
		Name:          "Scala Band",
		DriveFolderID: "drive-folder",
		Timezone:      "Europe/Kiev",
		AdminUserIDs:  []int64{42},
	}

	controller := WebAppController{
		UserService: &settingsStubUserService{users: map[int64]*entity.User{
			user.ID: user,
		}},
		BandService: &settingsStubBandService{bands: map[bson.ObjectID]*entity.Band{
			band.ID: band,
		}},
		JoinRequestService: settingsStubJoinRequestService{},
	}

	router := gin.New()
	router.GET("/api/settings/me", controller.SettingsMe)

	request := httptest.NewRequest(http.MethodGet, "/api/settings/me?userId=42", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response struct {
		Data SettingsMeResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data.User.Name != "Old Name" {
		t.Fatalf("expected stored user name, got %q", response.Data.User.Name)
	}
	if len(response.Data.Bands) != 1 {
		t.Fatalf("expected one band, got %d", len(response.Data.Bands))
	}
	if !response.Data.Bands[0].IsAdmin {
		t.Fatal("expected band admin flag")
	}
	if !response.Data.Bands[0].IsActive {
		t.Fatal("expected active band flag")
	}
}

func TestSettingsMeRejectsMissingUserID(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	controller := WebAppController{
		UserService:        &settingsStubUserService{users: map[int64]*entity.User{}},
		BandService:        &settingsStubBandService{bands: map[bson.ObjectID]*entity.Band{}},
		JoinRequestService: settingsStubJoinRequestService{},
	}

	router := gin.New()
	router.GET("/api/settings/me", controller.SettingsMe)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/settings/me", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestSettingsLeaveBandRejectsLastAdmin(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	bandID := bson.NewObjectID()
	otherBandID := bson.NewObjectID()
	user := &entity.User{
		ID:      42,
		Name:    "Alice",
		BandID:  bandID,
		BandIDs: []bson.ObjectID{bandID, otherBandID},
	}
	band := &entity.Band{
		ID:           bandID,
		Name:         "Scala Band",
		AdminUserIDs: []int64{42},
	}

	controller := WebAppController{
		UserService: &settingsStubUserService{users: map[int64]*entity.User{
			user.ID: user,
		}},
		BandService: &settingsStubBandService{bands: map[bson.ObjectID]*entity.Band{
			band.ID:     band,
			otherBandID: {ID: otherBandID, Name: "Other Band"},
		}},
		JoinRequestService: settingsStubJoinRequestService{},
	}

	router := gin.New()
	router.POST("/api/settings/bands/:id/leave", controller.SettingsLeaveBand)

	request := httptest.NewRequest(http.MethodPost, "/api/settings/bands/"+bandID.Hex()+"/leave", nil)
	request.URL.RawQuery = "userId=42"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !user.BelongsToBand(bandID) {
		t.Fatal("expected user to remain in band")
	}
}

func TestSettingsLeaveBandRejectsOnlyBand(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	bandID := bson.NewObjectID()
	user := &entity.User{
		ID:      42,
		Name:    "Alice",
		BandID:  bandID,
		BandIDs: []bson.ObjectID{bandID},
	}
	band := &entity.Band{ID: bandID, Name: "Scala Band"}

	controller := WebAppController{
		UserService: &settingsStubUserService{users: map[int64]*entity.User{
			user.ID: user,
		}},
		BandService: &settingsStubBandService{bands: map[bson.ObjectID]*entity.Band{
			band.ID: band,
		}},
		JoinRequestService: settingsStubJoinRequestService{},
	}

	router := gin.New()
	router.POST("/api/settings/bands/:id/leave", controller.SettingsLeaveBand)

	request := httptest.NewRequest(http.MethodPost, "/api/settings/bands/"+bandID.Hex()+"/leave", nil)
	request.URL.RawQuery = "userId=42"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !user.BelongsToBand(bandID) {
		t.Fatal("expected user to remain in band")
	}
}

func TestSettingsLeaveActiveBandSwitchesToRemainingBand(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	activeBandID := bson.NewObjectID()
	nextBandID := bson.NewObjectID()
	user := &entity.User{
		ID:      42,
		Name:    "Alice",
		BandID:  activeBandID,
		BandIDs: []bson.ObjectID{activeBandID, nextBandID},
	}

	controller := WebAppController{
		UserService: &settingsStubUserService{users: map[int64]*entity.User{
			user.ID: user,
		}},
		BandService: &settingsStubBandService{bands: map[bson.ObjectID]*entity.Band{
			activeBandID: {ID: activeBandID, Name: "Active Band"},
			nextBandID:   {ID: nextBandID, Name: "Next Band"},
		}},
		JoinRequestService: settingsStubJoinRequestService{},
	}

	router := gin.New()
	router.POST("/api/settings/bands/:id/leave", controller.SettingsLeaveBand)

	request := httptest.NewRequest(http.MethodPost, "/api/settings/bands/"+activeBandID.Hex()+"/leave", nil)
	request.URL.RawQuery = "userId=42"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if user.BelongsToBand(activeBandID) {
		t.Fatal("expected user to leave active band")
	}
	if user.BandID != nextBandID {
		t.Fatalf("expected active band %s, got %s", nextBandID.Hex(), user.BandID.Hex())
	}
}

func TestSettingsCancelJoinRequest(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	bandID := bson.NewObjectID()
	requestID := bson.NewObjectID()
	user := &entity.User{ID: 42, Name: "Alice"}
	band := &entity.Band{ID: bandID, Name: "Scala Band"}

	cancelCalled := false
	controller := WebAppController{
		UserService: &settingsStubUserService{users: map[int64]*entity.User{
			user.ID: user,
		}},
		BandService: &settingsStubBandService{bands: map[bson.ObjectID]*entity.Band{
			band.ID: band,
		}},
		JoinRequestService: settingsStubJoinRequestService{
			cancelFunc: func(userID int64, gotBandID bson.ObjectID) (*entity.JoinRequest, error) {
				cancelCalled = true
				if userID != user.ID {
					t.Fatalf("expected cancel user id %d, got %d", user.ID, userID)
				}
				if gotBandID != bandID {
					t.Fatalf("expected cancel band id %s, got %s", bandID.Hex(), gotBandID.Hex())
				}
				return &entity.JoinRequest{
					ID:        requestID,
					UserID:    userID,
					BandID:    gotBandID,
					Status:    entity.JoinRequestCanceled,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		},
	}

	router := gin.New()
	router.DELETE("/api/settings/bands/:id/join-requests", controller.SettingsCancelJoinRequest)

	request := httptest.NewRequest(http.MethodDelete, "/api/settings/bands/"+bandID.Hex()+"/join-requests?userId=42", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !cancelCalled {
		t.Fatal("expected cancel service to be called")
	}
}

func TestSettingsCreateJoinRequestAutoApprovesInTestBotAPI(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	bandID := bson.NewObjectID()
	requestID := bson.NewObjectID()
	user := &entity.User{ID: 42, Name: "Alice"}
	band := &entity.Band{ID: bandID, Name: "Scala Band"}

	approveCalled := false
	controller := WebAppController{
		IsTestBotAPI: true,
		UserService: &settingsStubUserService{users: map[int64]*entity.User{
			user.ID: user,
		}},
		BandService: &settingsStubBandService{bands: map[bson.ObjectID]*entity.Band{
			band.ID: band,
		}},
		JoinRequestService: settingsStubJoinRequestService{
			createFunc: func(input service.CreateJoinRequestInput) (*entity.JoinRequest, bool, error) {
				if input.UserID != user.ID {
					t.Fatalf("expected create user id %d, got %d", user.ID, input.UserID)
				}
				if input.Band.ID != bandID {
					t.Fatalf("expected create band id %s, got %s", bandID.Hex(), input.Band.ID.Hex())
				}
				return &entity.JoinRequest{
					ID:        requestID,
					UserID:    input.UserID,
					UserName:  input.UserName,
					BandID:    input.Band.ID,
					BandName:  input.Band.Name,
					Status:    entity.JoinRequestPending,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, true, nil
			},
			approveFunc: func(gotRequestID bson.ObjectID, decidedByUserID int64) (*entity.JoinRequest, *entity.User, error) {
				approveCalled = true
				if gotRequestID != requestID {
					t.Fatalf("expected approve request id %s, got %s", requestID.Hex(), gotRequestID.Hex())
				}
				if decidedByUserID != user.ID {
					t.Fatalf("expected decided by user id %d, got %d", user.ID, decidedByUserID)
				}
				return &entity.JoinRequest{
					ID:        gotRequestID,
					UserID:    user.ID,
					BandID:    bandID,
					Status:    entity.JoinRequestApproved,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, user, nil
			},
		},
	}

	router := gin.New()
	router.POST("/api/settings/bands/:id/join-requests", controller.SettingsCreateJoinRequest)

	request := httptest.NewRequest(http.MethodPost, "/api/settings/bands/"+bandID.Hex()+"/join-requests?userId=42", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !approveCalled {
		t.Fatal("expected approve service to be called")
	}
}
