package services

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joeyave/chords-transposer/transposer"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"regexp"
	"scala-chords-bot/entities"
	"scala-chords-bot/repositories"
	"time"
)

type SongService struct {
	songRepository *repositories.SongRepository
	driveClient    *drive.Service
	docsClient     *docs.Service
}

func NewSongService(songRepository *repositories.SongRepository, driveClient *drive.Service, docsClient *docs.Service) *SongService {
	return &SongService{
		songRepository: songRepository,
		driveClient:    driveClient,
		docsClient:     docsClient,
	}
}

/*
Searches for Song on Google Drive then returns uncached versions of Songs for performance reasons.
*/
func (s *SongService) FindByName(name string) ([]entities.Song, error) {
	var songs []entities.Song

	var pageToken string

	for {
		res, err := s.driveClient.Files.List().
			Q(fmt.Sprintf("fullText contains '\"%s\"'", name)).
			Fields("nextPageToken, files(id, name, modifiedTime, webViewLink)").
			PageToken(pageToken).
			Do()

		if err != nil {
			return nil, err
		}

		for _, file := range res.Files {
			actualSong := entities.Song{
				ID:           file.Id,
				Name:         file.Name,
				ModifiedTime: file.ModifiedTime,
				WebViewLink:  file.WebViewLink,
			}

			songs = append(songs, actualSong)
		}

		pageToken = res.NextPageToken

		if pageToken == "" {
			break
		}
	}

	if len(songs) == 0 {
		return nil, mongo.ErrEmptySlice
	}

	return songs, nil
}

func (s *SongService) FindOneByID(ID string) (entities.Song, error) {
	song, err := s.songRepository.FindOneByID(ID)
	return song, err
}

func (s *SongService) UpdateOne(song entities.Song) (entities.Song, error) {
	if song.ID == "" {
		return song, fmt.Errorf("ID is missing for Song: %v", song)
	}

	newSong, err := s.songRepository.UpdateOne(song)
	return newSong, err
}

func (s *SongService) GetWithActualTgFileID(song entities.Song) (entities.Song, error) {
	if song.ID == "" {
		return song, fmt.Errorf("ID is missing for Song: %v", song)
	}

	cachedSong, err := s.songRepository.FindOneByID(song.ID)
	if err != nil || cachedSong.TgFileID == "" {
		return entities.Song{}, errors.New("TgFileID is missing")
	}

	cachedModifiedTime, err := time.Parse(time.RFC3339, cachedSong.ModifiedTime)
	actualModifiedTime, err := time.Parse(time.RFC3339, song.ModifiedTime)

	if err != nil || actualModifiedTime.After(cachedModifiedTime) {
		return entities.Song{}, err
	}

	return cachedSong, err
}

func (s *SongService) DownloadPDF(song entities.Song) (*tgbotapi.FileReader, error) {
	if song.ID == "" {
		return nil, fmt.Errorf("ID is missing for Song: %v", song)
	}

	res, err := s.driveClient.Files.Export(song.ID, "application/pdf").Download()
	if err != nil {
		return nil, err
	}

	fileReader := &tgbotapi.FileReader{
		Name:   song.Name + ".pdf",
		Reader: res.Body,
		Size:   res.ContentLength,
	}

	return fileReader, err
}

func (s *SongService) GetSections(song entities.Song) ([]docs.StructuralElement, error) {
	sections := make([]docs.StructuralElement, 0)

	if song.ID == "" {
		return sections, fmt.Errorf("ID is missing for Song: %v", song)
	}

	doc, err := s.docsClient.Documents.Get(song.ID).Do()
	if err != nil {
		return sections, err
	}

	for i, section := range doc.Body.Content {
		if section.SectionBreak != nil &&
			section.SectionBreak.SectionStyle != nil &&
			section.SectionBreak.SectionStyle.SectionType == "NEXT_PAGE" ||
			i == 0 {
			if i == 0 {
				section.StartIndex = 0
				section.SectionBreak.SectionStyle.DefaultHeaderId = doc.DocumentStyle.DefaultHeaderId
			}

			sections = append(sections, *section)
		}
	}

	return sections, err
}

func (s *SongService) AppendSection(song entities.Song) ([]docs.StructuralElement, error) {
	sections := make([]docs.StructuralElement, 0)

	if song.ID == "" {
		return sections, fmt.Errorf("ID is missing for Song: %v", song)
	}

	requests := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertSectionBreak: &docs.InsertSectionBreakRequest{
					EndOfSegmentLocation: &docs.EndOfSegmentLocation{
						SegmentId: "",
					},
					SectionType: "NEXT_PAGE",
				},
			},
		},
	}

	_, err := s.docsClient.Documents.BatchUpdate(song.ID, requests).Do()
	if err != nil {
		return nil, err
	}

	sections, err = s.GetSections(song)
	if err != nil {
		return nil, err
	}

	requests = &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				CreateHeader: &docs.CreateHeaderRequest{
					SectionBreakLocation: &docs.Location{
						Index:     sections[len(sections)-1].StartIndex,
						SegmentId: "",
					},
					Type: "DEFAULT",
				},
			},
		},
	}

	_, err = s.docsClient.Documents.BatchUpdate(song.ID, requests).Do()
	if err != nil {
		return nil, err
	}

	return s.GetSections(song)
}

func (s *SongService) Transpose(song entities.Song, toKey string, sectionIndex int) (entities.Song, error) {
	if song.ID == "" {
		return song, fmt.Errorf("ID is missing for Song: %v", song)
	}

	doc, err := s.docsClient.Documents.Get(song.ID).Do()
	if err != nil {
		return entities.Song{}, err
	}

	sections, err := s.GetSections(song)
	if err != nil {
		return entities.Song{}, err
	}

	requests, _ := s.transposeHeader(doc, sections, sectionIndex, toKey)

	_, err = s.docsClient.Documents.BatchUpdate(doc.DocumentId,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	fmt.Println(err)

	//song.ModifiedTime = time.Now().Format(time.RFC3339)
	return song, err
}

func (s *SongService) transposeHeader(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, toKey string) ([]*docs.Request, string) {
	if doc.DocumentStyle.DefaultHeaderId == "" {
		return nil, ""
	}

	// Create header if section doesn't have it.
	if sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId == "" {
		requests := &docs.BatchUpdateDocumentRequest{
			Requests: []*docs.Request{
				{
					CreateHeader: &docs.CreateHeaderRequest{
						SectionBreakLocation: &docs.Location{
							SegmentId: "",
							Index:     sections[sectionIndex].StartIndex,
						},
						Type: "DEFAULT",
					},
				},
			},
		}

		_, err := s.docsClient.Documents.BatchUpdate(doc.DocumentId, requests).Do()
		if err != nil {
			return nil, ""
		}
	} else {
		header := doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId]
		if header.Content[len(header.Content)-1].EndIndex-1 > 0 {
			requests := &docs.BatchUpdateDocumentRequest{
				Requests: []*docs.Request{
					{
						DeleteContentRange: &docs.DeleteContentRangeRequest{
							Range: &docs.Range{
								StartIndex:      0,
								EndIndex:        header.Content[len(header.Content)-1].EndIndex - 1,
								SegmentId:       header.HeaderId,
								ForceSendFields: []string{"StartIndex"},
							},
						},
					},
				},
			}

			_, err := s.docsClient.Documents.BatchUpdate(doc.DocumentId, requests).Do()
			if err != nil {
				return nil, ""
			}
		}
	}

	requests := make([]*docs.Request, 0)

	key := ""
	var index int64 = 0

	for i, item := range doc.Headers[doc.DocumentStyle.DefaultHeaderId].Content {
		if item.Paragraph != nil && item.Paragraph.Elements != nil {
			for _, element := range item.Paragraph.Elements {
				if element.TextRun != nil && element.TextRun.Content != "" {
					if key == "" {
						guessedKey, err := transposer.GuessKeyFromText(element.TextRun.Content)
						if err == nil {
							key = guessedKey.String()
						}
					}

					transposedText, err := transposer.TransposeToKey(element.TextRun.Content, key, toKey)
					if err == nil {
						element.TextRun.Content = transposedText
					}

					if i == len(doc.Headers[doc.DocumentStyle.DefaultHeaderId].Content)-1 {
						re := regexp.MustCompile("[\\r\\n]$")
						element.TextRun.Content = re.ReplaceAllString(element.TextRun.Content, " ")
					}

					if len([]rune(element.TextRun.Content)) == 0 {
						continue
					}

					if element.TextRun.TextStyle.ForegroundColor == nil {
						element.TextRun.TextStyle.ForegroundColor = &docs.OptionalColor{
							Color: &docs.Color{
								RgbColor: &docs.RgbColor{
									Blue:  0,
									Green: 0,
									Red:   0,
								},
							},
						}
					}

					requests = append(requests,
						&docs.Request{
							InsertText: &docs.InsertTextRequest{
								Location: &docs.Location{
									Index:     index,
									SegmentId: doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId].HeaderId,
								},
								Text: element.TextRun.Content,
							},
						},
						&docs.Request{
							UpdateTextStyle: &docs.UpdateTextStyleRequest{
								Fields: "*",
								Range: &docs.Range{
									StartIndex: index,
									EndIndex:   index + int64(len([]rune(element.TextRun.Content))),
									SegmentId:  doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId].HeaderId,
									ForceSendFields: func() []string {
										if index == 0 {
											return []string{"StartIndex"}
										} else {
											return nil
										}
									}(),
								},
								TextStyle: element.TextRun.TextStyle,
							},
						},
						&docs.Request{
							UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
								Fields:         "alignment, lineSpacing, direction, spaceAbove, spaceBelow",
								ParagraphStyle: item.Paragraph.ParagraphStyle,
								Range: &docs.Range{
									StartIndex: index,
									EndIndex:   index + int64(len([]rune(element.TextRun.Content))),
									SegmentId:  doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId].HeaderId,
									ForceSendFields: func() []string {
										if index == 0 {
											return []string{"StartIndex"}
										} else {
											return nil
										}
									}(),
								},
							},
						},
					)

					index += int64(len([]rune(element.TextRun.Content)))
				}
			}
		}
	}

	return requests, key
}
