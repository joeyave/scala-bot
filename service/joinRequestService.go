package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type JoinRequestService struct {
	joinRequestRepository *repository.JoinRequestRepository
	userService           *UserService
}

func NewJoinRequestService(joinRequestRepository *repository.JoinRequestRepository, userService *UserService) *JoinRequestService {
	return &JoinRequestService{
		joinRequestRepository: joinRequestRepository,
		userService:           userService,
	}
}

type CreateJoinRequestInput struct {
	UserID   int64
	UserName string
	Band     *entity.Band
}

func (s *JoinRequestService) FindOneByID(ID bson.ObjectID) (*entity.JoinRequest, error) {
	return s.joinRequestRepository.FindOneByID(ID)
}

func (s *JoinRequestService) FindPendingByUserID(userID int64) ([]*entity.JoinRequest, error) {
	requests, err := s.joinRequestRepository.FindPendingByUserID(userID)
	if errors.Is(err, repository.ErrNotFound) {
		return []*entity.JoinRequest{}, nil
	}
	return requests, err
}

func (s *JoinRequestService) Create(input CreateJoinRequestInput) (*entity.JoinRequest, bool, error) {
	if input.Band == nil {
		return nil, false, ErrInvalidOperation
	}

	existingRequest, err := s.joinRequestRepository.FindPendingByUserIDAndBandID(input.UserID, input.Band.ID)
	if err == nil {
		return existingRequest, false, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, false, fmt.Errorf("finding pending join request: %w", err)
	}

	now := time.Now()
	request := entity.JoinRequest{
		UserID:    input.UserID,
		UserName:  input.UserName,
		BandID:    input.Band.ID,
		BandName:  input.Band.Name,
		Status:    entity.JoinRequestPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdRequest, err := s.joinRequestRepository.UpdateOne(request)
	if err != nil {
		return nil, false, fmt.Errorf("creating join request: %w", err)
	}

	return createdRequest, true, nil
}

func (s *JoinRequestService) Approve(requestID bson.ObjectID, decidedByUserID int64) (*entity.JoinRequest, *entity.User, error) {
	request, err := s.joinRequestRepository.FindOneByID(requestID)
	if err != nil {
		return nil, nil, fmt.Errorf("finding join request: %w", err)
	}
	if request.Status != entity.JoinRequestPending {
		return nil, nil, ErrInvalidOperation
	}

	user, err := s.userService.FindOneOrCreateByID(request.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("finding request user: %w", err)
	}

	user, err = s.userService.AddToBand(user, request.BandID)
	if err != nil {
		return nil, nil, fmt.Errorf("adding user to band: %w", err)
	}

	now := time.Now()
	request.Status = entity.JoinRequestApproved
	request.UpdatedAt = now
	request.DecidedAt = &now
	request.DecidedByUserID = decidedByUserID

	request, err = s.joinRequestRepository.UpdateOne(*request)
	if err != nil {
		return nil, nil, fmt.Errorf("approving join request: %w", err)
	}

	return request, user, nil
}

func (s *JoinRequestService) Decline(requestID bson.ObjectID, decidedByUserID int64) (*entity.JoinRequest, error) {
	request, err := s.joinRequestRepository.FindOneByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("finding join request: %w", err)
	}
	if request.Status != entity.JoinRequestPending {
		return nil, ErrInvalidOperation
	}

	now := time.Now()
	request.Status = entity.JoinRequestDeclined
	request.UpdatedAt = now
	request.DecidedAt = &now
	request.DecidedByUserID = decidedByUserID

	request, err = s.joinRequestRepository.UpdateOne(*request)
	if err != nil {
		return nil, fmt.Errorf("declining join request: %w", err)
	}

	return request, nil
}

func (s *JoinRequestService) Cancel(userID int64, bandID bson.ObjectID) (*entity.JoinRequest, error) {
	request, err := s.joinRequestRepository.FindPendingByUserIDAndBandID(userID, bandID)
	if err != nil {
		return nil, fmt.Errorf("finding pending join request: %w", err)
	}

	now := time.Now()
	request.Status = entity.JoinRequestCanceled
	request.UpdatedAt = now
	request.DecidedAt = &now
	request.DecidedByUserID = userID

	request, err = s.joinRequestRepository.UpdateOne(*request)
	if err != nil {
		return nil, fmt.Errorf("canceling join request: %w", err)
	}

	return request, nil
}
