package service

import (
	"github.com/joeyave/scala-bot-v2/entity"
	"github.com/joeyave/scala-bot-v2/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MembershipService struct {
	membershipRepository *repository.MembershipRepository
}

func NewMembershipService(membershipRepository *repository.MembershipRepository) *MembershipService {
	return &MembershipService{
		membershipRepository: membershipRepository,
	}
}

func (s *MembershipService) FindAll() ([]*entity.Membership, error) {
	return s.membershipRepository.FindAll()
}

func (s *MembershipService) FindOneByID(ID primitive.ObjectID) (*entity.Membership, error) {
	return s.membershipRepository.FindOneByID(ID)
}
func (s *MembershipService) FindMultipleByEventID(ID primitive.ObjectID) ([]*entity.Membership, error) {
	return s.membershipRepository.FindMultipleByEventID(ID)
}

func (s *MembershipService) FindSimilar(m *entity.Membership) (*entity.Membership, error) {
	memberships, err := s.membershipRepository.FindMultipleByUserIDAndEventIDAndRoleID(m.UserID, m.EventID, m.RoleID)
	if err != nil {
		return nil, err
	}
	return memberships[0], nil
}

func (s *MembershipService) UpdateOne(membership entity.Membership) (*entity.Membership, error) {
	memberships, err := s.membershipRepository.FindMultipleByUserIDAndEventID(membership.UserID, membership.EventID)
	if err == nil {
		for _, mb := range memberships {
			if mb.RoleID == membership.RoleID {
				membership.ID = mb.ID
				break
			}
		}
	}

	return s.membershipRepository.UpdateOne(membership)
}

func (s *MembershipService) DeleteOneByID(ID primitive.ObjectID) error {
	return s.membershipRepository.DeleteOneByID(ID)
}
