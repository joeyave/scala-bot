package services

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
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

func (s *EventService) FindManyFromTodayByBandIDAndUserID(bandID primitive.ObjectID, userID int64) ([]*entities.Event, error) {
	return s.eventRepository.FindManyFromTodayByBandIDAndUserID(bandID, userID)
}

func (s *EventService) FindOneOldestByBandID(bandID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.FindOneOldestByBandID(bandID)
}

func (s *EventService) FindOneByID(ID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.FindOneByID(ID)
}

func (s *EventService) FindOneByNameAndTime(name string, time time.Time) (*entities.Event, error) {
	return s.eventRepository.FindOneByNameAndTime(name, time)
}

func (s *EventService) UpdateOne(event entities.Event) (*entities.Event, error) {
	return s.eventRepository.UpdateOne(event)
}

func (s *EventService) PushSongID(eventID primitive.ObjectID, songID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.PushSongID(eventID, songID)
}

func (s *EventService) PullSongID(eventID primitive.ObjectID, songID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.PullSongID(eventID, songID)
}
func (s *EventService) ChangeSongIDPosition(eventID primitive.ObjectID, songID primitive.ObjectID, newPosition int) (*entities.Event, error) {
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
		eventString = fmt.Sprintf("%s\n\n<b>Список:</b>", eventString)

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

				//key, BPM, _ := s.driveFileService.GetMetadata(event.Songs[i].DriveFileID)

				//songName := fmt.Sprintf("%d. <a href=\"%s\">%s</a> (%s, %s)", i+1, driveFile.WebViewLink, driveFile.Name, key, BPM)
				songName := fmt.Sprintf("%d. <a href=\"%s\">%s</a>  (%s, %s)",
					i+1, driveFile.WebViewLink, driveFile.Name, event.Songs[i].PDF.Key, event.Songs[i].PDF.BPM)
				songNames[i] = songName
			}(i)
		}
		waitGroup.Wait()

		eventString += "\n" + strings.Join(songNames, "\n")
	}

	return eventString, event, nil
}
