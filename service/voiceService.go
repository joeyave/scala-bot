package service

import (
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoiceService struct {
	voiceRepository *repository.VoiceRepository
}

func NewVoiceService(voiceRepository *repository.VoiceRepository) *VoiceService {
	return &VoiceService{
		voiceRepository: voiceRepository,
	}
}

func (s *VoiceService) FindOneByID(ID primitive.ObjectID) (*entity.Voice, error) {
	return s.voiceRepository.FindOneByID(ID)
}

func (s *VoiceService) FindOneByFileID(fileID string) (*entity.Voice, error) {
	return s.voiceRepository.FindOneByFileID(fileID)
}

func (s *VoiceService) UpdateOne(voice entity.Voice) (*entity.Voice, error) {
	return s.voiceRepository.UpdateOne(voice)
}

func (s *VoiceService) DeleteOne(ID primitive.ObjectID) error {
	return s.voiceRepository.DeleteOneByID(ID)
}

func (s *VoiceService) DeleteManyByIDs(IDs []primitive.ObjectID) error {
	return s.voiceRepository.DeleteManyByIDs(IDs)
}

func (s *VoiceService) CloneVoicesForNewSongID(oldSongID, newSongID primitive.ObjectID) error {
	return s.voiceRepository.CloneVoicesForNewSongID(oldSongID, newSongID)
}
