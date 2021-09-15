package services

import (
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoiceService struct {
	voiceRepository *repositories.VoiceRepository
}

func NewVoiceService(voiceRepository *repositories.VoiceRepository) *VoiceService {
	return &VoiceService{
		voiceRepository: voiceRepository,
	}
}

func (s *VoiceService) FindOneByID(ID primitive.ObjectID) (*entities.Voice, error) {
	return s.voiceRepository.FindOneByID(ID)
}

func (s *VoiceService) FindOneByFileID(fileID string) (*entities.Voice, error) {
	return s.voiceRepository.FindOneByFileID(fileID)
}

func (s *VoiceService) UpdateOne(voice entities.Voice) (*entities.Voice, error) {
	return s.voiceRepository.UpdateOne(voice)
}

func (s *VoiceService) DeleteOne(ID primitive.ObjectID) error {
	return s.voiceRepository.DeleteOneByID(ID)
}
