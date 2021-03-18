package services

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
)

type UserService struct {
	userRepository *repositories.UserRepository
}

func NewUserService(userRepository *repositories.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) CreateOne(user entities.User) (*entities.User, error) {
	if user.ID == 0 {
		return nil, fmt.Errorf("invalid id for User %v", user)
	}

	_, err := s.userRepository.FindOneByID(user.ID)
	if err == nil {
		return nil, fmt.Errorf("User with id %d already exists", user.ID)
	}

	return s.userRepository.UpdateOne(user)
}

func (s *UserService) FindOneByID(ID int64) (*entities.User, error) {
	return s.userRepository.FindOneByID(ID)
}

func (s *UserService) UpdateOne(user entities.User) (*entities.User, error) {
	return s.userRepository.UpdateOne(user)
}
