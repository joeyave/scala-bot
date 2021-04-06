package services

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"strings"
	"sync"
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

func (s *EventService) FindOneByID(ID primitive.ObjectID) (*entities.Event, error) {
	event, err := s.eventRepository.FindOneByID(ID)
	if err != nil {
		return nil, err
	}

	return event, err
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

func (s *EventService) ToHtmlStringByID(ID primitive.ObjectID) (string, error) {

	event, err := s.eventRepository.FindOneByID(ID)
	if err != nil {
		return "", err
	}

	eventString := fmt.Sprintf("<b>%s</b>     <a href=\"\">&#8204;</a>", event.Alias())
	membershipGroups := map[string][]*entities.Membership{}
	for _, membership := range event.Memberships {
		if membership.Role == nil {
			continue
		}
		membershipGroups[membership.Role.Name] = append(membershipGroups[membership.Role.Name], membership)
	}

	for membershipName, memberships := range membershipGroups {
		eventString = fmt.Sprintf("%s\n\n<b>%s:</b>", eventString, membershipName)

		var userIDs []int64
		for _, membership := range memberships {
			userIDs = append(userIDs, membership.UserID)
		}
		users, err := s.userRepository.FindMultipleByIDs(userIDs)
		if err != nil {
			continue
		}

		for i, user := range users {
			eventString = fmt.Sprintf("%s\n%d. <a href=\"tg://user?id=%d\">%s</a>", eventString, i+1, user.ID, user.Name)
		}
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
				songName := fmt.Sprintf("%d. <a href=\"%s\">%s</a>", i+1, driveFile.WebViewLink, driveFile.Name)
				songNames[i] = songName
			}(i)
		}
		waitGroup.Wait()

		eventString += "\n" + strings.Join(songNames, "\n")
	}

	return eventString, nil
}
