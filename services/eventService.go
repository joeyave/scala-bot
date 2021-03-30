package services

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventService struct {
	eventRepository *repositories.EventRepository
	userRepository  *repositories.UserRepository
}

func NewEventService(eventRepository *repositories.EventRepository, userRepository *repositories.UserRepository) *EventService {
	return &EventService{
		eventRepository: eventRepository,
		userRepository:  userRepository,
	}
}

func (s *EventService) FindAll() ([]*entities.Event, error) {
	return s.eventRepository.FindAll()
}

func (s *EventService) FindMultipleByBandID(bandID primitive.ObjectID) ([]*entities.Event, error) {
	return s.eventRepository.FindMultipleByBandID(bandID)
}

func (s *EventService) FindMultipleByBandIDFromToday(bandID primitive.ObjectID) ([]*entities.Event, error) {
	return s.eventRepository.FindMultipleByBandIDFromToday(bandID)
}

func (s *EventService) FindOneByID(ID primitive.ObjectID) (*entities.Event, error) {
	return s.eventRepository.FindOneByID(ID)
}

func (s *EventService) UpdateOne(event entities.Event) (*entities.Event, error) {
	return s.eventRepository.UpdateOne(event)
}

func (s *EventService) ToHtmlStringByID(ID primitive.ObjectID) (string, error) {

	event, err := s.eventRepository.FindOneByID(ID)
	if err != nil {
		return "", err
	}

	eventString := event.Alias()
	membershipGroups := map[string][]*entities.Membership{}
	for _, membership := range event.Memberships {
		if membership.Role == nil {
			continue
		}
		membershipGroups[membership.Role.Name] = append(membershipGroups[membership.Role.Name], membership)
	}

	for membershipName, memberships := range membershipGroups {
		eventString = fmt.Sprintf("%s\n\n%s:", eventString, membershipName)

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

	return eventString, nil
}
