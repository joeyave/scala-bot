package service

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-bot-v2/entity"
	"github.com/joeyave/scala-bot-v2/repository"
	"github.com/joeyave/scala-bot-v2/txt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"sync"
	"time"
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

func (s *EventService) FindAllFromToday() ([]*entity.Event, error) {
	return s.eventRepository.FindAllFromToday()
}

func (s *EventService) FindManyFromTodayByBandID(bandID primitive.ObjectID) ([]*entity.Event, error) {
	return s.eventRepository.FindManyFromTodayByBandID(bandID)
}

func (s *EventService) FindManyFromTodayByBandIDAndWeekday(bandID primitive.ObjectID, weekday time.Weekday) ([]*entity.Event, error) {
	events, err := s.eventRepository.FindManyFromTodayByBandID(bandID)
	if err != nil {
		return nil, err
	}

	var events2 []*entity.Event
	for _, event := range events {
		if event.Time.Weekday() == weekday {
			events2 = append(events2, event)
		}
	}
	return events2, nil
}

func (s *EventService) FindManyBetweenDatesByBandID(from time.Time, to time.Time, bandID primitive.ObjectID) ([]*entity.Event, error) {
	return s.eventRepository.FindManyBetweenDatesByBandID(from, to, bandID)
}

func (s *EventService) FindManyByBandIDAndPageNumber(bandID primitive.ObjectID, pageNumber int) ([]*entity.Event, error) {
	return s.eventRepository.FindManyByBandIDAndPageNumber(bandID, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndPageNumber(bandID primitive.ObjectID, pageNumber int) ([]*entity.Event, error) {
	return s.eventRepository.FindManyUntilTodayByBandIDAndPageNumber(bandID, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndWeekdayAndPageNumber(bandID primitive.ObjectID, weekday time.Weekday, pageNumber int) ([]*entity.Event, error) {
	return s.eventRepository.FindManyUntilTodayByBandIDAndWeekdayAndPageNumber(bandID, weekday, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndUserIDAndPageNumber(bandID primitive.ObjectID, userID int64, pageNumber int) ([]*entity.Event, error) {
	return s.eventRepository.FindManyUntilTodayByBandIDAndUserIDAndPageNumber(bandID, userID, pageNumber)
}

func (s *EventService) FindManyFromTodayByBandIDAndUserID(bandID primitive.ObjectID, userID int64, pageNumber int) ([]*entity.Event, error) {
	return s.eventRepository.FindManyFromTodayByBandIDAndUserID(bandID, userID, pageNumber)
}

func (s *EventService) FindOneOldestByBandID(bandID primitive.ObjectID) (*entity.Event, error) {
	return s.eventRepository.FindOneOldestByBandID(bandID)
}

func (s *EventService) FindOneByID(ID primitive.ObjectID) (*entity.Event, error) {
	return s.eventRepository.FindOneByID(ID)
}

func (s *EventService) FindOneByNameAndTimeAndBandID(name string, time time.Time, bandID primitive.ObjectID) (*entity.Event, error) {
	return s.eventRepository.FindOneByNameAndTimeAndBandID(name, time, bandID)
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
					return
				}
				songName := fmt.Sprintf("%d. <a href=\"%s\">%s</a>  (%s)", i+1, driveFile.WebViewLink, driveFile.Name, event.Songs[i].Meta())
				songNames[i] = songName
			}(i)
		}
		waitGroup.Wait()

		fmt.Fprintf(&b, "\n%s", strings.Join(songNames, "\n"))
	}

	if event.Notes != "" {
		fmt.Fprintf(&b, "\n\n%s", event.NotesString(lang))
	}

	return b.String()
}

func (s *EventService) GetMostFrequentEventNames(bandID primitive.ObjectID, limit int) ([]*entity.EventNameFrequencies, error) {
	return s.eventRepository.GetMostFrequentEventNames(bandID, limit)
}

func (s *EventService) GetEventWithSongs(eventID primitive.ObjectID) (*entity.Event, error) {
	return s.eventRepository.GetEventWithSongs(eventID)
}
