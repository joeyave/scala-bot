package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type stubStatisticsBandService struct {
	band *entity.Band
}

func (s stubStatisticsBandService) FindOneByID(bson.ObjectID) (*entity.Band, error) {
	return s.band, nil
}

type stubStatisticsUserService struct {
	users     []*entity.UserWithEvents
	fromCalls []time.Time
}

func (s *stubStatisticsUserService) FindOneByID(int64) (*entity.User, error) {
	return nil, nil
}

func (s *stubStatisticsUserService) FindManyExtraByBandID(_ bson.ObjectID, from, _ time.Time) ([]*entity.UserWithEvents, error) {
	s.fromCalls = append(s.fromCalls, from)
	return s.users, nil
}

func TestStatisticsRedirectsToReactRoute(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/web-app/statistics?bandId=band-123&lang=ru", nil)

	controller := WebAppController{}
	controller.Statistics(ctx)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	expected := "/webapp-react/#/statistics?bandId=band-123&lang=ru"
	if location != expected {
		t.Fatalf("expected redirect to %q, got %q", expected, location)
	}
}

func TestStatisticsDataReturnsWrappedPayloadAndAcceptsDateFormats(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	location, err := time.LoadLocation("Europe/Kiev")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	bandID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	eventID := bson.NewObjectID()
	band := &entity.Band{
		ID:       bandID,
		Name:     "Team Alpha",
		Timezone: "Europe/Kiev",
		Roles: []*entity.Role{
			{ID: roleID, Name: "Vocal"},
		},
	}

	users := []*entity.UserWithEvents{
		{
			User: entity.User{
				ID:   42,
				Name: "Alice",
			},
			Events: []*entity.Event{
				{
					ID:      eventID,
					Name:    "Sunday Set",
					TimeUTC: time.Date(2026, time.February, 15, 8, 0, 0, 0, time.UTC),
					Memberships: []*entity.Membership{
						{
							UserID: 42,
							Role: &entity.Role{
								ID:   roleID,
								Name: "Vocal",
							},
						},
					},
				},
			},
		},
	}

	userService := &stubStatisticsUserService{users: users}
	controller := WebAppController{
		BandService: stubStatisticsBandService{band: band},
		UserService: userService,
	}

	for _, rawDate := range []string{"2026-02-01", "01.02.2026"} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(
			http.MethodGet,
			"/api/statistics?bandId="+bandID.Hex()+"&from="+rawDate,
			nil,
		)

		controller.StatisticsData(ctx)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status %d for %s, got %d", http.StatusOK, rawDate, recorder.Code)
		}
	}

	if len(userService.fromCalls) != 2 {
		t.Fatalf("expected 2 service calls, got %d", len(userService.fromCalls))
	}

	expectedFrom := time.Date(2026, time.February, 1, 0, 0, 0, 0, location)
	for _, fromCall := range userService.fromCalls {
		if !fromCall.Equal(expectedFrom) {
			t.Fatalf("expected parsed from date %s, got %s", expectedFrom, fromCall)
		}
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/statistics?bandId="+bandID.Hex(),
		nil,
	)

	controller.StatisticsData(ctx)

	var response struct {
		Data StatisticsResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data.BandName != "Team Alpha" {
		t.Fatalf("expected band name %q, got %q", "Team Alpha", response.Data.BandName)
	}

	nowInLocation := time.Now().In(location)
	sixMonthsAgo := nowInLocation.AddDate(0, -6, 0)
	expectedDefaultFromDate := time.Date(
		sixMonthsAgo.Year(),
		sixMonthsAgo.Month(),
		sixMonthsAgo.Day(),
		0,
		0,
		0,
		0,
		location,
	).Format("2006-01-02")
	if response.Data.DefaultFromDate != expectedDefaultFromDate {
		t.Fatalf("expected defaultFromDate %q, got %q", expectedDefaultFromDate, response.Data.DefaultFromDate)
	}

	expectedCurrentDate := time.Now().In(location).Format("2006-01-02")
	if response.Data.CurrentDate != expectedCurrentDate {
		t.Fatalf("expected currentDate %q, got %q", expectedCurrentDate, response.Data.CurrentDate)
	}

	if len(response.Data.Roles) != 1 || response.Data.Roles[0].Name != "Vocal" {
		t.Fatalf("expected one Vocal role in response, got %+v", response.Data.Roles)
	}

	if len(response.Data.Users) != 1 {
		t.Fatalf("expected one user in response, got %d", len(response.Data.Users))
	}

	if len(response.Data.Users[0].Events) != 1 {
		t.Fatalf("expected one event in response, got %d", len(response.Data.Users[0].Events))
	}
}
