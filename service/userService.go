package service

import (
	"time"

	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) FindOneByID(ID int64) (*entity.User, error) {
	return s.userRepository.FindOneByID(ID)
}

func (s *UserService) FindOneOrCreateByID(ID int64) (*entity.User, error) {
	user, err := s.userRepository.FindOneByID(ID)
	if err != nil {
		user, err = s.userRepository.UpdateOne(entity.User{ID: ID})
		if err != nil {
			return nil, err
		}
	}

	return user, err
}

func (s *UserService) FindOneByName(name string) (*entity.User, error) {
	return s.userRepository.FindOneByName(name)
}

func (s *UserService) FindMultipleByBandID(bandID bson.ObjectID) ([]*entity.User, error) {
	return s.userRepository.FindManyByBandID(bandID)
}

func (s *UserService) FindMultipleByIDs(IDs []int64) ([]*entity.User, error) {
	return s.userRepository.FindManyByIDs(IDs)
}

func (s *UserService) UpdateOne(user entity.User) (*entity.User, error) {
	return s.userRepository.UpdateOne(user)
}

func (s *UserService) FindManyByBandIDAndRoleID(bandID, roleID bson.ObjectID, from time.Time) ([]*entity.UserWithEvents, error) {
	return s.userRepository.FindManyExtraByBandIDAndRoleID(bandID, roleID, from)
}

func (s *UserService) FindManyExtraByBandID(bandID bson.ObjectID, from, to time.Time) ([]*entity.UserWithEvents, error) {
	return s.userRepository.FindManyExtraByBandID(bandID, from, to)
}
