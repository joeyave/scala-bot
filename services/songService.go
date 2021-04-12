package services

import (
	"errors"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	"github.com/kjk/notionapi"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"sync"
	"time"
)

type SongService struct {
	songRepository   *repositories.SongRepository
	voiceRepository  *repositories.VoiceRepository
	bandRepository   *repositories.BandRepository
	driveClient      *drive.Service
	notionClient     *notionapi.Client
	driveFileService *DriveFileService
}

func NewSongService(songRepository *repositories.SongRepository, voiceRepository *repositories.VoiceRepository, bandRepository *repositories.BandRepository,
	driveClient *drive.Service, notionClient *notionapi.Client, driveFileService *DriveFileService) *SongService {
	return &SongService{
		songRepository:   songRepository,
		voiceRepository:  voiceRepository,
		bandRepository:   bandRepository,
		driveClient:      driveClient,
		notionClient:     notionClient,
		driveFileService: driveFileService,
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

func (s *SongService) FindOrCreateOneByDriveFileID(driveFileID string) (*entities.Song, *drive.File, error) {
	driveFile, err := s.driveClient.Files.Get(driveFileID).Fields("id, name, modifiedTime, webViewLink, parents").Do()

	song, err := s.songRepository.FindOneByDriveFileID(driveFileID)
	if err != nil {
		song = &entities.Song{
			DriveFileID: driveFile.Id,
		}

		for _, parentFolderID := range driveFile.Parents {
			band, err := s.bandRepository.FindOneByDriveFolderID(parentFolderID)
			if err == nil {
				song.BandID = band.ID
				break
			}
		}
	}

	if songHasOutdatedPDF(song, driveFile) ||
		song.PDF.Name == "" || song.PDF.Key == "" || song.PDF.BPM == "" || song.PDF.Time == "" {
		song.PDF.Name = driveFile.Name
		song.PDF.Key, song.PDF.BPM, song.PDF.Time = s.driveFileService.GetMetadata(driveFile.Id)
		song.PDF.TgFileID = ""
		song.PDF.ModifiedTime = driveFile.ModifiedTime
	}

	song, err = s.songRepository.UpdateOne(*song)
	return song, driveFile, err
}

func (s *SongService) FindOrCreateManyByDriveFileIDs(driveFileIDs []string) ([]*entities.Song, []*drive.File, error) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(driveFileIDs))
	songs := make([]*entities.Song, len(driveFileIDs))
	driveFiles := make([]*drive.File, len(driveFileIDs))
	var err error
	for i := range driveFileIDs {
		go func(i int) {
			defer waitGroup.Done()

			song, driveFile, _err := s.FindOrCreateOneByDriveFileID(driveFileIDs[i])
			if _err != nil {
				err = _err
			}
			songs[i] = song
			driveFiles[i] = driveFile
		}(i)
	}
	waitGroup.Wait()

	return songs, driveFiles, err

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

func songHasOutdatedPDF(song *entities.Song, driveFile *drive.File) bool {
	if song.PDF.ModifiedTime == "" || driveFile == nil {
		return true
	}

	pdfModifiedTime, err := time.Parse(time.RFC3339, song.PDF.ModifiedTime)
	if err != nil {
		return true
	}

	driveFileModifiedTime, err := time.Parse(time.RFC3339, driveFile.ModifiedTime)
	if err != nil {
		return true
	}

	if driveFileModifiedTime.After(pdfModifiedTime) {
		return true
	}

	return false
}
