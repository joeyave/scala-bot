package service

import (
	"errors"
	"fmt"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"time"
)

type SongService struct {
	songRepository   *repository.SongRepository
	voiceRepository  *repository.VoiceRepository
	bandRepository   *repository.BandRepository
	driveRepository  *drive.Service
	driveFileService *DriveFileService
}

func NewSongService(songRepository *repository.SongRepository, voiceRepository *repository.VoiceRepository, bandRepository *repository.BandRepository,
	driveClient *drive.Service, driveFileService *DriveFileService) *SongService {
	return &SongService{
		songRepository:   songRepository,
		voiceRepository:  voiceRepository,
		bandRepository:   bandRepository,
		driveRepository:  driveClient,
		driveFileService: driveFileService,
	}
}

func (s *SongService) FindAll() ([]*entity.Song, error) {
	return s.songRepository.FindAll()
}

func (s *SongService) FindManyLiked(bandID primitive.ObjectID, userID int64) ([]*entity.Song, error) {
	return s.songRepository.FindManyLiked(bandID, userID)
}

func (s *SongService) FindManyByDriveFileIDs(IDs []string) ([]*entity.Song, error) {
	return s.songRepository.FindManyByDriveFileIDs(IDs)
}

func (s *SongService) FindManyExtraLiked(bandID primitive.ObjectID, userID int64, eventsStartDate time.Time, pageNumber int) ([]*entity.SongWithEvents, error) {
	return s.songRepository.FindManyExtraByPageNumberLiked(bandID, userID, eventsStartDate, pageNumber)
}

func (s *SongService) FindOneByID(ID primitive.ObjectID) (*entity.Song, error) {
	return s.songRepository.FindOneByID(ID)
}

func (s *SongService) FindOneByDriveFileID(driveFileID string) (*entity.Song, error) {
	return s.songRepository.FindOneByDriveFileID(driveFileID)
}

func (s *SongService) FindOneByName(driveFileID string) (*entity.Song, error) {
	return s.songRepository.FindOneByName(driveFileID)
}

func (s *SongService) FindOrCreateOneByDriveFileID(driveFileID string) (*entity.Song, *drive.File, error) {
	var driveFile *drive.File
	err := errors.New("fake error")
	for err != nil {
		driveFile, err = s.driveRepository.Files.Get(driveFileID).Fields("id, name, modifiedTime, webViewLink, parents").Do()
	}

	song, err := s.songRepository.FindOneByDriveFileID(driveFileID)
	if err != nil {
		song = &entity.Song{
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

func (s *SongService) FindOrCreateManyByDriveFileIDs(driveFileIDs []string) ([]*entity.Song, []*drive.File, error) {

	errwg := new(errgroup.Group)
	songs := make([]*entity.Song, len(driveFileIDs))
	driveFiles := make([]*drive.File, len(driveFileIDs))

	for i := range driveFileIDs {
		i := i
		errwg.Go(func() error {
			song, driveFile, err := s.FindOrCreateOneByDriveFileID(driveFileIDs[i])
			if err != nil {
				return err
			}
			songs[i] = song
			driveFiles[i] = driveFile
			return nil
		})
	}
	err := errwg.Wait()
	if err != nil {
		return nil, nil, err
	}
	return songs, driveFiles, err
}

func (s *SongService) UpdateOne(song entity.Song) (*entity.Song, error) {
	return s.songRepository.UpdateOne(song)
}

func (s *SongService) UpdateMany(songs []*entity.Song) (*mongo.BulkWriteResult, error) {
	return s.songRepository.UpdateMany(songs)
}

func (s *SongService) DeleteOneByDriveFileID(driveFileID string) error {
	err := s.driveRepository.Files.Delete(driveFileID).Do()
	if err != nil {
		var gErr *googleapi.Error
		if errors.As(err, &gErr) {
			isInsufficientFilePermissions := false
			for _, e := range gErr.Errors {
				if e.Reason == "insufficientFilePermissions" {
					isInsufficientFilePermissions = true
					break
				}
			}
			if !isInsufficientFilePermissions {
				return err
			}
		} else {
			return err
		}
	}

	_, err = s.songRepository.DeleteOneByDriveFileID(driveFileID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SongService) DeleteOneByDriveFileIDFromDatabase(driveFileID string) (bool, error) {
	deletedCount, err := s.songRepository.DeleteOneByDriveFileID(driveFileID)
	deleted := deletedCount > 0
	if err != nil {
		return deleted, err
	}

	return deleted, nil
}

func (s *SongService) Like(songID primitive.ObjectID, userID int64) error {
	return s.songRepository.Like(songID, userID)
}

const archiveFolderName = "Archive"

func (s *SongService) Dislike(songID primitive.ObjectID, userID int64) error {
	return s.songRepository.Dislike(songID, userID)
}

func (s *SongService) Archive(songID primitive.ObjectID) (*drive.File, error) {

	song, err := s.FindOneByID(songID)
	if err != nil {
		return nil, err
	}

	if song.Band == nil {
		return nil, fmt.Errorf("band not found for song %s", songID)
	}

	if song.Band.ArchiveFolderID == "" {
		archiveFolder, err := s.driveFileService.FindOrCreateOneFolderByNameAndFolderID(archiveFolderName, song.Band.DriveFolderID)
		if err != nil {
			return nil, err
		}
		song.Band.ArchiveFolderID = archiveFolder.Id

		_, _ = s.bandRepository.UpdateOne(*song.Band)
	}

	driveFile, err := s.driveFileService.MoveOne(song.DriveFileID, song.Band.ArchiveFolderID)
	if err != nil {
		return nil, err
	}

	_ = s.songRepository.Archive(songID)

	return driveFile, err
}

func (s *SongService) Unarchive(songID primitive.ObjectID) (*drive.File, error) {

	song, err := s.FindOneByID(songID)
	if err != nil {
		return nil, err
	}

	driveFile, err := s.driveFileService.MoveOne(song.DriveFileID, song.Band.DriveFolderID)
	if err != nil {
		return nil, err
	}

	_ = s.songRepository.Unarchive(songID)

	return driveFile, err
}

func (s *SongService) FindAllExtraByPageNumberSortedByEventsNumber(bandID primitive.ObjectID, eventsStartDate time.Time, isAscending bool, pageNumber int) ([]*entity.SongWithEvents, error) {
	return s.songRepository.FindAllExtraByPageNumberSortedByEventsNumber(bandID, eventsStartDate, isAscending, pageNumber)
}

func (s *SongService) FindAllExtraByPageNumberSortedByEventDate(bandID primitive.ObjectID, eventsStartDate time.Time, isAscending bool, pageNumber int) ([]*entity.SongWithEvents, error) {
	return s.songRepository.FindAllExtraByPageNumberSortedByEventDate(bandID, eventsStartDate, isAscending, pageNumber)
}

func (s *SongService) FindManyExtraByTag(tag string, bandID primitive.ObjectID, eventsStartDate time.Time, pageNumber int) ([]*entity.SongWithEvents, error) {
	return s.songRepository.FindManyExtraByTag(tag, bandID, eventsStartDate, pageNumber)
}

func songHasOutdatedPDF(song *entity.Song, driveFile *drive.File) bool {
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

func (s *SongService) GetTags(bandID primitive.ObjectID) ([]string, error) {
	return s.songRepository.GetTags(bandID)
}

func (s *SongService) TagOrUntag(tag string, songID primitive.ObjectID) (*entity.Song, error) {
	return s.songRepository.TagOrUntag(tag, songID)
}
