package services

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
)

type RoleService struct {
	roleRepository *repositories.RoleRepository
}

func NewRoleService(roleRepository *repositories.RoleRepository) *RoleService {
	return &RoleService{
		roleRepository: roleRepository,
	}
}

func (s *RoleService) FindAll() ([]*entities.Role, error) {
	return s.roleRepository.FindAll()
}

func (s *RoleService) UpdateOne(role entities.Role) (*entities.Role, error) {
	return s.roleRepository.UpdateOne(role)
}
