package services

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
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

func (s *UserService) FindOneOrCreateByID(ID int64) (*entities.User, error) {
	user, err := s.userRepository.FindOneByID(ID)
	if err != nil {
		user, err = s.userRepository.UpdateOne(entities.User{
			ID: ID,
			State: &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return user, err
}

func (s *UserService) FindOneByName(name string) (*entities.User, error) {
	return s.userRepository.FindOneByName(name)
}

func (s *UserService) FindMultipleByBandID(bandID primitive.ObjectID) ([]*entities.User, error) {
	return s.userRepository.FindManyByBandID(bandID)
}

func (s *UserService) FindMultipleByIDs(IDs []int64) ([]*entities.User, error) {
	return s.userRepository.FindManyByIDs(IDs)
}

func (s *UserService) UpdateOne(user entities.User) (*entities.User, error) {
	return s.userRepository.UpdateOne(user)
}

func (s *UserService) FindManyByBandIDAndRoleID(bandID, roleID primitive.ObjectID) ([]*entities.UserExtra, error) {
	return s.userRepository.FindManyByBandIDAndRoleID(bandID, roleID)
}
