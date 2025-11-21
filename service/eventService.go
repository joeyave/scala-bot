package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/joeyave/scala-bot/repository"
	"github.com/joeyave/scala-bot/txt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventService struct {
	eventRepository      *repository.EventRepository
	membershipRepository *repository.MembershipRepository
	driveFileService     *DriveFileService
}

func NewEventService(eventRepository *repository.EventRepository, membershipRepository *repository.MembershipRepository, driveFileService *DriveFileService) *EventService {
	return &EventService{
		eventRepository:      eventRepository,
		membershipRepository: membershipRepository,
		driveFileService:     driveFileService,
	}
}

func (s *EventService) FindManyFromTodayByBandID(bandID primitive.ObjectID, loc *time.Location) ([]*entity.Event, error) {
	startOfDayUTC := helpers.GetStartOfDayInLocUTC(loc)
	return s.eventRepository.FindManyFromDateByBandID(bandID, startOfDayUTC)
}

func (s *EventService) FindManyFromTodayByBandIDAndWeekday(bandID primitive.ObjectID, weekday time.Weekday, loc *time.Location) ([]*entity.Event, error) {
	startOfDayUTC := helpers.GetStartOfDayInLocUTC(loc)

	events, err := s.eventRepository.FindManyFromDateByBandID(bandID, startOfDayUTC)
	if err != nil {
		return nil, err
	}

	var events2 []*entity.Event
	for _, event := range events {
		if event.GetLocalTime().Weekday() == weekday {
			events2 = append(events2, event)
		}
	}
	return events2, nil
}

func (s *EventService) FindBetweenDates(fromLoc, toLoc time.Time) ([]*entity.Event, error) {
	fromUTC := fromLoc.UTC()
	toUTC := toLoc.UTC()

	return s.eventRepository.FindBetweenDates(fromUTC, toUTC)
}

func (s *EventService) FindManyBetweenDatesByBandID(fromLoc, toLoc time.Time, bandID primitive.ObjectID) ([]*entity.Event, error) {
	fromUTC := fromLoc.UTC()
	toUTC := toLoc.UTC()
	return s.eventRepository.FindManyBetweenDatesByBandID(fromUTC, toUTC, bandID)
}

func (s *EventService) FindManyByBandIDAndPageNumber(bandID primitive.ObjectID, pageNumber int) ([]*entity.Event, error) {
	return s.eventRepository.FindManyByBandIDAndPageNumber(bandID, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndPageNumber(bandID primitive.ObjectID, loc *time.Location, pageNumber int) ([]*entity.Event, error) {
	todayStartUTC := helpers.GetStartOfDayInLocUTC(loc)
	return s.eventRepository.FindManyUntilByBandIDAndPageNumber(bandID, todayStartUTC, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndWeekdayAndPageNumber(bandID primitive.ObjectID, loc *time.Location, weekday time.Weekday, pageNumber int) ([]*entity.Event, error) {
	todayStartUTC := helpers.GetStartOfDayInLocUTC(loc)
	return s.eventRepository.FindManyUntilByBandIDAndWeekdayAndPageNumber(bandID, todayStartUTC, weekday, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndUserIDAndPageNumber(bandID primitive.ObjectID, loc *time.Location, userID int64, pageNumber int) ([]*entity.Event, error) {
	todayStartUTC := helpers.GetStartOfDayInLocUTC(loc)
	return s.eventRepository.FindManyUntilByBandIDAndUserIDAndPageNumber(bandID, userID, todayStartUTC, pageNumber)
}

func (s *EventService) FindManyFromTodayByBandIDAndUserID(bandID primitive.ObjectID, loc *time.Location, userID int64, pageNumber int) ([]*entity.Event, error) {
	todayStartUTC := helpers.GetStartOfDayInLocUTC(loc)
	return s.eventRepository.FindManyFromTodayByBandIDAndUserID(bandID, userID, todayStartUTC, pageNumber)
}

func (s *EventService) FindOneOldestByBandID(bandID primitive.ObjectID) (*entity.Event, error) {
	return s.eventRepository.FindOneOldestByBandID(bandID)
}

func (s *EventService) FindOneByID(ID primitive.ObjectID) (*entity.Event, error) {
	return s.eventRepository.FindOneByID(ID)
}

func (s *EventService) FindOneByNameAndTimeAndBandID(name string, timeLoc time.Time, bandID primitive.ObjectID) (*entity.Event, error) {
	fromLocal := time.Date(
		timeLoc.Year(), timeLoc.Month(), timeLoc.Day(),
		0, 0, 0, 0,
		timeLoc.Location())
	toLocal := fromLocal.AddDate(0, 0, 1)

	// Convert to UTC for DB
	fromUTC := fromLocal.UTC()
	toUTC := toLocal.UTC()

	return s.eventRepository.FindOneByNameAndTimeAndBandID(name, fromUTC, toUTC, bandID)
}

func (s *EventService) GetAlias(ctx context.Context, eventID primitive.ObjectID, lang string) (string, error) {
	return s.eventRepository.GetAlias(ctx, eventID, lang)
}

func (s *EventService) UpdateOne(event entity.Event) (*entity.Event, error) {
	return s.eventRepository.UpdateOne(event)
}

func (s *EventService) PushSongID(eventID primitive.ObjectID, songID primitive.ObjectID) error {
	return s.eventRepository.PushSongID(eventID, songID)
}

func (s *EventService) PullSongID(eventID primitive.ObjectID, songID primitive.ObjectID) error {
	return s.eventRepository.PullSongID(eventID, songID)
}

func (s *EventService) ChangeSongIDPosition(eventID primitive.ObjectID, songID primitive.ObjectID, newPosition int) error {
	return s.eventRepository.ChangeSongIDPosition(eventID, songID, newPosition)
}

func (s *EventService) DeleteOneByID(ID primitive.ObjectID) error {
	err := s.eventRepository.DeleteOneByID(ID)
	if err != nil {
		return err
	}

	err = s.membershipRepository.DeleteManyByEventID(ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *EventService) ToHtmlStringByID(ID primitive.ObjectID, lang string) (string, *entity.Event, error) {

	event, err := s.eventRepository.FindOneByID(ID)
	if err != nil {
		return "", nil, err
	}

	return s.ToHtmlStringByEvent(*event, lang), event, nil
}

func (s *EventService) ToHtmlStringByEvent(event entity.Event, lang string) string {

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>", event.Alias(lang))
	rolesString := event.RolesString()
	if rolesString != "" {
		fmt.Fprintf(&b, "\n\n%s", rolesString)
	}

	if len(event.Songs) > 0 {
		fmt.Fprintf(&b, "\n\n<b>%s:</b>", txt.Get("button.setlist", lang))

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(event.Songs))
		songNames := make([]string, len(event.Songs))
		for i := range event.Songs {
			go func(i int) {
				defer waitGroup.Done()
				driveFile, err := s.driveFileService.FindOneByID(event.Songs[i].DriveFileID)
				if err != nil {
					// todo: output some message.
					return
				}
				songName := fmt.Sprintf("%d. <a href=\"%s\">%s</a>  (%s)", i+1, driveFile.WebViewLink, driveFile.Name, event.Songs[i].Meta())
				songNames[i] = songName
			}(i)
		}
		waitGroup.Wait()

		fmt.Fprintf(&b, "\n%s", strings.Join(songNames, "\n"))
	}

	if event.Notes != nil && *event.Notes != "" {
		fmt.Fprintf(&b, "\n\n%s", event.NotesString(lang))
	}

	return b.String()
}

func (s *EventService) GetMostFrequentEventNames(bandID primitive.ObjectID, limit int) ([]*entity.EventNameFrequencies, error) {
	fromUTC := time.Now().AddDate(0, -3, 0).UTC()
	return s.eventRepository.GetMostFrequentEventNames(bandID, limit, fromUTC)
}

func (s *EventService) GetEventWithSongs(eventID primitive.ObjectID) (*entity.Event, error) {
	return s.eventRepository.GetEventWithSongs(eventID)
}
