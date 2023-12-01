package service

import (
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BandService struct {
	bandRepository *repository.BandRepository
}

func NewBandService(bandRepository *repository.BandRepository) *BandService {
	return &BandService{
		bandRepository: bandRepository,
	}
}

func (s *BandService) FindAll() ([]*entity.Band, error) {
	return s.bandRepository.FindAll()
}

func (s *BandService) FindManyByIDs(ids []primitive.ObjectID) ([]*entity.Band, error) {
	return s.bandRepository.FindManyByIDs(ids)
}

func (s *BandService) FindOneByID(ID primitive.ObjectID) (*entity.Band, error) {
	return s.bandRepository.FindOneByID(ID)
}

func (s *BandService) FindOneByDriveFolderID(driveFolderID string) (*entity.Band, error) {
	return s.bandRepository.FindOneByDriveFolderID(driveFolderID)
}

func (s *BandService) UpdateOne(band entity.Band) (*entity.Band, error) {
	return s.bandRepository.UpdateOne(band)
}
