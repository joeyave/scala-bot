package services

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
)

type BandService struct {
	bandRepository *repositories.BandRepository
}

func NewBandService(bandRepository *repositories.BandRepository) *BandService {
	return &BandService{
		bandRepository: bandRepository,
	}
}

func (s *BandService) FindAll() ([]entities.Band, error) {
	return s.bandRepository.FindAll()
}

func (s *BandService) UpdateOne(band entities.Band) (entities.Band, error) {
	return s.bandRepository.UpdateOne(band)
}
