package services

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	userRepository *repositories.UserRepository
}

func NewUserService(userRepository *repositories.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) FindOneByID(ID int64) (*entities.User, error) {
	return s.userRepository.FindOneByID(ID)
}

func (s *UserService) FindMultipleByBandID(bandID primitive.ObjectID) ([]*entities.User, error) {
	return s.userRepository.FindMultipleByBandID(bandID)
}

func (s *UserService) UpdateOne(user entities.User) (*entities.User, error) {
	return s.userRepository.UpdateOne(user)
}
