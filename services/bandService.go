package services

import (
	"github.com/joeyave/scala-bot/entities"
	"github.com/joeyave/scala-bot/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BandService struct {
	bandRepository *repositories.BandRepository
}

func NewBandService(bandRepository *repositories.BandRepository) *BandService {
	return &BandService{
		bandRepository: bandRepository,
	}
}

func (s *BandService) FindAll() ([]*entities.Band, error) {
	return s.bandRepository.FindAll()
}

func (s *BandService) FindOneByID(ID primitive.ObjectID) (*entities.Band, error) {
	return s.bandRepository.FindOneByID(ID)
}

func (s *BandService) FindOneByDriveFolderID(driveFolderID string) (*entities.Band, error) {
	return s.bandRepository.FindOneByDriveFolderID(driveFolderID)
}

func (s *BandService) UpdateOne(band entities.Band) (*entities.Band, error) {
	return s.bandRepository.UpdateOne(band)
}
