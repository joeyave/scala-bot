package services

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-bot/entities"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/joeyave/scala-bot/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"strings"
	"sync"
	"time"
)

type EventService struct {
	eventRepository      *repositories.EventRepository
	userRepository       *repositories.UserRepository
	membershipRepository *repositories.MembershipRepository
	driveRepository      *drive.Service
	driveFileService     *DriveFileService
}

func NewEventService(eventRepository *repositories.EventRepository, userRepository *repositories.UserRepository, membershipRepository *repositories.MembershipRepository, driveRepository *drive.Service, driveFileService *DriveFileService) *EventService {
	return &EventService{
		eventRepository:      eventRepository,
		userRepository:       userRepository,
		membershipRepository: membershipRepository,
		driveRepository:      driveRepository,
		driveFileService:     driveFileService,
	}
}

func (s *EventService) FindAllFromToday() ([]*entities.Event, error) {
	return s.eventRepository.FindAllFromToday()
}

func (s *EventService) FindManyFromTodayByBandID(bandID primitive.ObjectID) ([]*entities.Event, error) {
	return s.eventRepository.FindManyFromTodayByBandID(bandID)
}

func (s *EventService) FindManyFromTodayByBandIDAndWeekday(bandID primitive.ObjectID, weekday time.Weekday) ([]*entities.Event, error) {
	events, err := s.eventRepository.FindManyFromTodayByBandID(bandID)
	if err != nil {
		return nil, err
	}

	var events2 []*entities.Event
	for _, event := range events {
		if event.Time.Weekday() == weekday {
			events2 = append(events2, event)
		}
	}
	return events2, nil
}

func (s *EventService) FindManyBetweenDatesByBandID(from time.Time, to time.Time, bandID primitive.ObjectID) ([]*entities.Event, error) {
	return s.eventRepository.FindManyBetweenDatesByBandID(from, to, bandID)
}

func (s *EventService) FindManyByBandIDAndPageNumber(bandID primitive.ObjectID, pageNumber int) ([]*entities.Event, error) {
	return s.eventRepository.FindManyByBandIDAndPageNumber(bandID, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndPageNumber(bandID primitive.ObjectID, pageNumber int) ([]*entities.Event, error) {
	return s.eventRepository.FindManyUntilTodayByBandIDAndPageNumber(bandID, pageNumber)
}

func (s *EventService) FindManyUntilTodayByBandIDAndWeekdayAndPageNumber(bandID primitive.ObjectID, weekday time.Weekday, pageNumber int) ([]*entities.Event, error) {
	return s.eventRepository.FindManyUntilTodayByBandIDAndWeekdayAndPageNumber(bandID, weekday, pageNumber)
}

func (s *EventService) FindManyFromTodayByBandIDAndUserID(bandID primitive.ObjectID, userID int64, pageNumber int) ([]*entities.Event, error) {
	return s.eventRepository.FindManyFromTodayByBandIDAndUserID(bandID, userID, pageNumber)
}

func (s *EventService) FindOneOldestByBandID(bandID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.FindOneOldestByBandID(bandID)
}

func (s *EventService) FindOneByID(ID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.FindOneByID(ID)
}

func (s *EventService) FindOneByNameAndTimeAndBandID(name string, time time.Time, bandID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.FindOneByNameAndTimeAndBandID(name, time, bandID)
}

func (s *EventService) GetAlias(ctx context.Context, eventID primitive.ObjectID) (string, error) {
	return s.eventRepository.GetAlias(ctx, eventID)
}

func (s *EventService) UpdateOne(event entities.Event) (*entities.Event, error) {
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

func (s *EventService) ToHtmlStringByID(ID primitive.ObjectID) (string, *entities.Event, error) {

	event, err := s.eventRepository.FindOneByID(ID)
	if err != nil {
		return "", nil, err
	}

	return s.ToHtmlStringByEvent(*event), event, nil
}

func (s *EventService) ToHtmlStringByEvent(event entities.Event) string {
	eventString := fmt.Sprintf("<b>%s</b>", event.Alias())

	var currRoleID primitive.ObjectID
	for _, membership := range event.Memberships {
		if membership.User == nil {
			continue
		}

		if currRoleID != membership.RoleID {
			currRoleID = membership.RoleID
			eventString = fmt.Sprintf("%s\n\n<b>%s:</b>", eventString, membership.Role.Name)
		}

		eventString = fmt.Sprintf("%s\n - <a href=\"tg://user?id=%d\">%s</a>", eventString, membership.User.ID, membership.User.Name)
	}

	if len(event.Songs) > 0 {
		eventString = fmt.Sprintf("%s\n\n<b>%s:</b>", eventString, helpers.Setlist)

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

				songName := fmt.Sprintf("%d. <a href=\"%s\">%s</a>  (%s)",
					i+1, driveFile.WebViewLink, driveFile.Name, event.Songs[i].Caption())
				songNames[i] = songName
			}(i)
		}
		waitGroup.Wait()

		eventString += "\n" + strings.Join(songNames, "\n")
	}

	if event.Notes != "" {
		eventString += "\n\n<b>" + helpers.Notes + ":</b>\n" + event.Notes
	}

	return eventString
}

func (s *EventService) GetMostFrequentEventNames() ([]*entities.EventNameFrequencies, error) {
	return s.eventRepository.GetMostFrequentEventNames()
}

func (s *EventService) GetEventWithSongs(eventID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.GetEventWithSongs(eventID)
}
