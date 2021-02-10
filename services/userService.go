package services

import (
	"scalaChordsBot/entities"
	"scalaChordsBot/repositories"
)

type UserService struct {
	userRepository *repositories.UserRepository
}

func NewUserService(userRepository *repositories.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) FindOrCreate(ID int64) (entities.User, error) {
	user, err := s.userRepository.FindOneByID(ID)

	// Create User if he doesn't exists or doesn't have states.
	if err != nil || user.States == nil {
		user, err = s.userRepository.UpdateOne(*entities.NewUser(ID))
		if err != nil {
			return entities.User{}, err
		}
	}

	return user, err
}

func (s *UserService) UpdateOne(user entities.User) (entities.User, error) {
	user, err := s.userRepository.UpdateOne(user)
	return user, err
}
