package services

import (
	"errors"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	"github.com/kjk/notionapi"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
)

type SongService struct {
	songRepository  *repositories.SongRepository
	voiceRepository *repositories.VoiceRepository
	bandRepository  *repositories.BandRepository
	driveClient     *drive.Service
	notionClient    *notionapi.Client
}

func NewSongService(songRepository *repositories.SongRepository, voiceRepository *repositories.VoiceRepository, bandRepository *repositories.BandRepository,
	driveClient *drive.Service, notionClient *notionapi.Client) *SongService {
	return &SongService{
		songRepository:  songRepository,
		voiceRepository: voiceRepository,
		bandRepository:  bandRepository,
		driveClient:     driveClient,
		notionClient:    notionClient,
	}
}

func (s *SongService) FindAll() ([]*entities.Song, error) {
	return s.songRepository.FindAll()
}

func (s *SongService) FindOneByID(ID primitive.ObjectID) (*entities.Song, error) {
	return s.songRepository.FindOneByID(ID)
}

func (s *SongService) FindOneByDriveFileID(driveFileID string) (*entities.Song, error) {
	return s.songRepository.FindOneByDriveFileID(driveFileID)
}

func (s *SongService) UpdateOne(song entities.Song) (*entities.Song, error) {
	return s.songRepository.UpdateOne(song)
}

func (s *SongService) DeleteOneByID(ID string) error {
	return s.songRepository.DeleteOneByID(ID)
}

func (s *SongService) FindNotionPageByID(pageID string) (*notionapi.Block, error) {
	res, err := s.notionClient.LoadPageChunk(pageID, 0, nil)
	if err != nil {
		return nil, err
	}

	record, ok := res.RecordMap.Blocks[pageID]
	if !ok {
		return nil, errors.New("TODO")
	}

	block := record.Block
	if block == nil || !block.IsPage() || block.IsSubPage() || block.IsLinkToPage() {
		return nil, errors.New("TODO")
	}

	return block, nil
}
