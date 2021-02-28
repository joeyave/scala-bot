package services

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
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

func (s *UserService) FindOrCreate(ID int64) (*entities.User, error) {
	user, err := s.userRepository.FindOneByID(ID)

	// Create User if he doesn't exists or doesn't have states.
	if err != nil {
		user = &entities.User{
			ID: ID,
			State: &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			},
		}

		user, err = s.userRepository.UpdateOne(*user)
		if err != nil {
			return nil, err
		}
	}

	if (user.Bands == nil || len(user.Bands) == 0) &&
		user.State.Name != helpers.ChooseBandState && user.State.Name != helpers.CreateBandState {
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.ChooseBandState,
		}

		user, err = s.userRepository.UpdateOne(*user)
		if err != nil {
			return nil, err
		}
	}

	return user, err
}

func (s *UserService) UpdateOne(user entities.User) (*entities.User, error) {
	return s.userRepository.UpdateOne(user)
}
