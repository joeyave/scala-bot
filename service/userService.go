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

func (s *UserService) AddToBand(user *entity.User, bandID bson.ObjectID) (*entity.User, error) {
	if user == nil {
		return nil, ErrInvalidOperation
	}

	user.BandIDs = appendUniqueBandID(user.BandIDs, bandID)
	if user.BandID.IsZero() {
		user.BandID = bandID
	}

	return s.UpdateOne(*user)
}

func (s *UserService) RemoveFromBand(user *entity.User, bandID bson.ObjectID) (*entity.User, error) {
	if user == nil {
		return nil, ErrInvalidOperation
	}

	remainingBandIDs := make([]bson.ObjectID, 0, len(user.BandIDs))
	for _, userBandID := range user.BandIDs {
		if userBandID != bandID {
			remainingBandIDs = appendUniqueBandID(remainingBandIDs, userBandID)
		}
	}

	user.BandIDs = remainingBandIDs
	if user.BandID == bandID {
		if len(user.BandIDs) > 0 {
			user.BandID = user.BandIDs[0]
		} else {
			user.BandID = bson.NilObjectID
		}
	}

	return s.UpdateOne(*user)
}

func (s *UserService) SetActiveBand(user *entity.User, bandID bson.ObjectID) (*entity.User, error) {
	if user == nil {
		return nil, ErrInvalidOperation
	}
	if !user.BelongsToBand(bandID) {
		return nil, ErrForbidden
	}

	user.BandID = bandID
	user.BandIDs = appendUniqueBandID(user.BandIDs, bandID)
	return s.UpdateOne(*user)
}

func (s *UserService) FindManyByBandIDAndRoleID(bandID, roleID bson.ObjectID, from time.Time) ([]*entity.UserWithEvents, error) {
	return s.userRepository.FindManyExtraByBandIDAndRoleID(bandID, roleID, from)
}

func (s *UserService) FindManyExtraByBandID(bandID bson.ObjectID, from, to time.Time) ([]*entity.UserWithEvents, error) {
	return s.userRepository.FindManyExtraByBandID(bandID, from, to)
}

func appendUniqueBandID(ids []bson.ObjectID, id bson.ObjectID) []bson.ObjectID {
	for _, existingID := range ids {
		if existingID == id {
			return ids
		}
	}
	return append(ids, id)
}
