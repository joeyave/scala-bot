package services

import (
	"errors"
	"github.com/joeyave/scala-bot/entities"
	"github.com/joeyave/scala-bot/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"sync"
	"time"
)

type SongService struct {
	songRepository   *repositories.SongRepository
	voiceRepository  *repositories.VoiceRepository
	bandRepository   *repositories.BandRepository
	driveRepository  *drive.Service
	driveFileService *DriveFileService
}

func NewSongService(songRepository *repositories.SongRepository, voiceRepository *repositories.VoiceRepository, bandRepository *repositories.BandRepository,
	driveClient *drive.Service, driveFileService *DriveFileService) *SongService {
	return &SongService{
		songRepository:   songRepository,
		voiceRepository:  voiceRepository,
		bandRepository:   bandRepository,
		driveRepository:  driveClient,
		driveFileService: driveFileService,
	}
}

func (s *SongService) FindAll() ([]*entities.Song, error) {
	return s.songRepository.FindAll()
}

func (s *SongService) FindManyLiked(userID int64) ([]*entities.Song, error) {
	return s.songRepository.FindManyLiked(userID)
}

func (s *SongService) FindManyByDriveFileIDs(IDs []string) ([]*entities.Song, error) {
	return s.songRepository.FindManyByDriveFileIDs(IDs)
}

func (s *SongService) FindManyExtraLiked(userID int64, pageNumber int) ([]*entities.SongExtra, error) {
	return s.songRepository.FindManyExtraByPageNumberLiked(userID, pageNumber)
}

func (s *SongService) FindOneByID(ID primitive.ObjectID) (*entities.Song, error) {
	return s.songRepository.FindOneByID(ID)
}

func (s *SongService) FindOneByDriveFileID(driveFileID string) (*entities.Song, error) {
	return s.songRepository.FindOneByDriveFileID(driveFileID)
}

func (s *SongService) FindOneByName(driveFileID string) (*entities.Song, error) {
	return s.songRepository.FindOneByName(driveFileID)
}

func (s *SongService) FindOrCreateOneByDriveFileID(driveFileID string) (*entities.Song, *drive.File, error) {
	var driveFile *drive.File
	err := errors.New("fake error")
	for err != nil {
		driveFile, err = s.driveRepository.Files.Get(driveFileID).Fields("id, name, modifiedTime, webViewLink, parents").Do()
	}

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
		song.PDF.Name == "" || song.PDF.Key == "" || song.PDF.BPM == "" || song.PDF.Time == "" || song.PDF.WebViewLink == "" {
		song.PDF.Name = driveFile.Name
		song.PDF.Key, song.PDF.BPM, song.PDF.Time = s.driveFileService.GetMetadata(driveFile.Id)
		song.PDF.TgFileID = ""
		song.PDF.ModifiedTime = driveFile.ModifiedTime
		song.PDF.WebViewLink = driveFile.WebViewLink
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

func (s *SongService) DeleteOneByDriveFileID(driveFileID string) error {
	err := s.driveRepository.Files.Delete(driveFileID).Do()
	if err != nil {
		return err
	}

	err = s.songRepository.DeleteOneByDriveFileID(driveFileID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SongService) Like(songID primitive.ObjectID, userID int64) error {
	return s.songRepository.Like(songID, userID)
}

func (s *SongService) Dislike(songID primitive.ObjectID, userID int64) error {
	return s.songRepository.Dislike(songID, userID)
}

func (s *SongService) FindAllExtraByPageNumberSortedByEventsNumber(bandID primitive.ObjectID, pageNumber int) ([]*entities.SongExtra, error) {
	return s.songRepository.FindAllExtraByPageNumberSortedByEventsNumber(bandID, pageNumber)
}

func (s *SongService) FindAllExtraByPageNumberSortedByLatestEventDate(bandID primitive.ObjectID, pageNumber int) ([]*entities.SongExtra, error) {
	return s.songRepository.FindAllExtraByPageNumberSortedByLatestEventDate(bandID, pageNumber)
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
