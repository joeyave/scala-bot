package services

import (
	"errors"
	"fmt"
	"github.com/flowchartsman/retry"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/scala-chords-bot/repositories"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"github.com/kjk/notionapi"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"regexp"
	"strings"
	"time"
)

type SongService struct {
	songRepository *repositories.SongRepository
	driveClient    *drive.Service
	docsClient     *docs.Service
	notionClient   *notionapi.Client
}

func NewSongService(songRepository *repositories.SongRepository, driveClient *drive.Service, docsClient *docs.Service, notionClient *notionapi.Client) *SongService {
	return &SongService{
		songRepository: songRepository,
		driveClient:    driveClient,
		docsClient:     docsClient,
		notionClient:   notionClient,
	}
}

/*
Searches for Song on Google Drive then returns uncached versions of Songs for performance reasons.
*/
func (s *SongService) QueryDrive(name string, pageToken string, folderID string) ([]*drive.File, string, error) {
	name = helpers.JsonEscape(name)

	q := fmt.Sprintf(`fullText contains '%s'`+
		` and trashed = false`+
		` and mimeType = 'application/vnd.google-apps.document'`, name)

	if folderID != "" {
		q += fmt.Sprintf(` and '%s' in parents`, folderID)
	}

	res, err := s.driveClient.Files.List().
		// Use this for precise search.
		//Q(fmt.Sprintf("fullText contains '\"%s\"'", name)).
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, webViewLink, parents)").
		PageSize(90).PageToken(pageToken).Do()

	if err != nil {
		return nil, "", err
	}

	return res.Files, res.NextPageToken, nil
}

func (s *SongService) FindOneByID(ID string) (*entities.Song, error) {
	file, err := s.driveClient.Files.
		Get(ID).
		Fields("id, name, modifiedTime, webViewLink, parents").Do()
	if err != nil {
		return nil, err
	}

	song, err := s.songRepository.FindOneByID(ID)
	if err != nil {
		err = nil
		song = &entities.Song{
			ID: file.Id,
		}
	}

	song.DriveFile = file

	return song, err
}

func (s *SongService) UpdateOne(song entities.Song) (*entities.Song, error) {
	if song.ID == "" {
		return nil, fmt.Errorf("ID is missing for Song: %v", song)
	}

	return s.songRepository.UpdateOne(song)
}

func (s *SongService) CreateOne(name string, lyrics string, key string, BPM string, time string, folderID string) (*entities.Song, error) {

	fileToCreate := &drive.File{
		Name:     name,
		Parents:  []string{folderID},
		MimeType: "application/vnd.google-apps.document",
	}

	newFile, err := s.driveClient.Files.Create(fileToCreate).Do()
	if err != nil {
		return nil, err
	}

	requests := make([]*docs.Request, 0)

	requests = append(requests,
		&docs.Request{
			CreateHeader: &docs.CreateHeaderRequest{
				Type: "DEFAULT",
			},
		},
		&docs.Request{
			InsertText: &docs.InsertTextRequest{
				EndOfSegmentLocation: &docs.EndOfSegmentLocation{
					SegmentId: "",
				},
				Text: lyrics,
			},
		})

	res, err := s.docsClient.Documents.BatchUpdate(newFile.Id,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	if err != nil {
		return nil, err
	}

	if res.Replies[0].CreateHeader.HeaderId != "" {
		_, _ = s.docsClient.Documents.BatchUpdate(newFile.Id,
			&docs.BatchUpdateDocumentRequest{
				Requests: []*docs.Request{
					getDefaultHeaderRequest(res.Replies[0].CreateHeader.HeaderId, name, key, BPM, time),
				},
			}).Do()
	}

	doc, err := s.docsClient.Documents.Get(newFile.Id).Do()
	if err != nil {
		return nil, err
	}

	requests = nil
	for _, paragraph := range doc.Body.Content {
		if paragraph.Paragraph == nil {
			continue
		}

		for _, element := range paragraph.Paragraph.Elements {
			if element.TextRun == nil || element.TextRun.TextStyle == nil {
				continue
			}

			element.TextRun.TextStyle.FontSize = &docs.Dimension{
				Magnitude: 14,
				Unit:      "PT",
			}

			requests = append(requests, &docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Fields: "*",
					Range: &docs.Range{
						SegmentId:       "",
						StartIndex:      element.StartIndex,
						EndIndex:        element.EndIndex,
						ForceSendFields: []string{"StartIndex"},
					},
					TextStyle: element.TextRun.TextStyle,
				},
			})
		}
	}

	_, _ = s.docsClient.Documents.BatchUpdate(newFile.Id,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()

	newSong := entities.Song{
		ID:        newFile.Id,
		DriveFile: newFile,
	}

	return &newSong, err
}

func (s *SongService) DeleteOne(ID string) error {
	err := s.driveClient.Files.Delete(ID).Do()
	if err != nil {
		return err
	}

	return s.songRepository.DeleteOneByID(ID)
}

func (s *SongService) DownloadPDFByID(songID string) (tgbotapi.FileReader, error) {
	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var fileReader tgbotapi.FileReader
	err := retrier.Run(func() error {
		file, err := s.driveClient.Files.Get(songID).Fields("id, name").Do()
		if err != nil {
			return err
		}

		res, err := s.driveClient.Files.Export(songID, "application/pdf").Download()
		if err != nil {
			return err
		}

		fileReader = tgbotapi.FileReader{
			Name:   file.Name + ".pdf",
			Reader: res.Body,
		}
		return nil
	})
	if err != nil {
		return tgbotapi.FileReader{}, err
	}

	return fileReader, err
}

func (s *SongService) DeepCopyToFolder(song entities.Song, folderID string) (*entities.Song, error) {
	if song.ID == "" {
		return nil, fmt.Errorf("ID is missing for Song: %v", song)
	}

	file := &drive.File{
		Name:    song.DriveFile.Name,
		Parents: []string{folderID},
	}
	newFile, err := s.driveClient.Files.Copy(song.ID, file).Fields("id, name, modifiedTime, webViewLink, parents").Do()
	if err != nil {
		return nil, err
	}

	folderPermissionsList, err := s.driveClient.Permissions.List(folderID).
		Fields("*").
		PageSize(100).Do()
	if err != nil {
		return nil, err
	}

	var folderOwnerPermission *drive.Permission
	for _, permission := range folderPermissionsList.Permissions {
		if permission.Role == "owner" {
			folderOwnerPermission = permission
		}
	}

	if folderOwnerPermission != nil {
		permission := &drive.Permission{
			EmailAddress: folderOwnerPermission.EmailAddress,
			Role:         "owner",
			Type:         "user",
		}
		_, err = s.driveClient.Permissions.
			Create(newFile.Id, permission).
			TransferOwnership(true).Do()
		if err != nil {
			return nil, err
		}

	}

	newSong := entities.Song{
		ID:        newFile.Id,
		DriveFile: newFile,
		PDF:       song.PDF,
		Voices:    song.Voices,
	}

	return s.UpdateOne(newSong)
}

func (s *SongService) GetSectionsByID(songID string) ([]docs.StructuralElement, error) {
	sections := make([]docs.StructuralElement, 0)

	doc, err := s.docsClient.Documents.Get(songID).Do()
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

func (s *SongService) AppendSection(songID string) ([]docs.StructuralElement, error) {
	sections := make([]docs.StructuralElement, 0)

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

	_, err := s.docsClient.Documents.BatchUpdate(songID, requests).Do()
	if err != nil {
		return nil, err
	}

	sections, err = s.GetSectionsByID(songID)
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

	_, err = s.docsClient.Documents.BatchUpdate(songID, requests).Do()
	if err != nil {
		return nil, err
	}

	return s.GetSectionsByID(songID)
}

func (s *SongService) Transpose(song entities.Song, toKey string, sectionIndex int) (*entities.Song, error) {
	if song.ID == "" {
		return nil, fmt.Errorf("ID is missing for Song: %v", song)
	}

	doc, err := s.docsClient.Documents.Get(song.ID).Do()
	if err != nil {
		return nil, err
	}

	sections, err := s.GetSectionsByID(song.ID)
	if err != nil {
		return nil, err
	}

	requests, key := s.transposeHeader(doc, sections, sectionIndex, toKey)
	requests = append(requests, s.transposeBody(doc, sections, sectionIndex, key, toKey)...)

	_, err = s.docsClient.Documents.BatchUpdate(doc.DocumentId,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()

	fakeTime, _ := time.Parse("2006", "2006")
	song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)
	return &song, err
}

func (s *SongService) transposeHeader(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, toKey string) ([]*docs.Request, string) {
	if doc.DocumentStyle.DefaultHeaderId == "" {
		return nil, ""
	}

	requests := make([]*docs.Request, 0)

	// Create header if section doesn't have it.
	if sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId == "" {
		requests = append(requests, &docs.Request{
			CreateHeader: &docs.CreateHeaderRequest{
				SectionBreakLocation: &docs.Location{
					SegmentId: "",
					Index:     sections[sectionIndex].StartIndex,
				},
				Type: "DEFAULT",
			},
		})
	} else {
		header := doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId]
		if header.Content[len(header.Content)-1].EndIndex-1 > 0 {
			requests = append(requests, &docs.Request{
				DeleteContentRange: &docs.DeleteContentRangeRequest{
					Range: &docs.Range{
						StartIndex:      0,
						EndIndex:        header.Content[len(header.Content)-1].EndIndex - 1,
						SegmentId:       header.HeaderId,
						ForceSendFields: []string{"StartIndex"},
					},
				},
			})

		}
	}

	r, key := composeTransposeRequests(doc.Headers[doc.DocumentStyle.DefaultHeaderId].Content,
		0, "", toKey, doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId].HeaderId)
	requests = append(requests, r...)

	return requests, key
}

func (s *SongService) transposeBody(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, key string, toKey string) []*docs.Request {
	requests := make([]*docs.Request, 0)

	sectionToInsertStartIndex := sections[sectionIndex].StartIndex + 1
	var sectionToInsertEndIndex int64

	if len(sections) > sectionIndex+1 {
		sectionToInsertEndIndex = sections[sectionIndex+1].StartIndex - 1
	} else {
		sectionToInsertEndIndex = doc.Body.Content[len(doc.Body.Content)-1].EndIndex - 1
	}

	var content []*docs.StructuralElement
	if len(sections) > 1 {
		index := len(doc.Body.Content)
		for i := range doc.Body.Content {
			if doc.Body.Content[i].StartIndex == sections[1].StartIndex {
				index = i
				break
			}
		}
		content = doc.Body.Content[:index]
	} else {
		content = doc.Body.Content
	}

	if sectionToInsertEndIndex-sectionToInsertStartIndex > 0 {
		requests = append(requests, &docs.Request{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex:      sectionToInsertStartIndex,
					EndIndex:        sectionToInsertEndIndex,
					SegmentId:       "",
					ForceSendFields: []string{"StartIndex"},
				},
			},
		})
	}

	r, _ := composeTransposeRequests(content, sectionToInsertStartIndex, key, toKey, "")
	requests = append(requests, r...)

	return requests
}

func composeTransposeRequests(content []*docs.StructuralElement, index int64, key string, toKey string, segmentId string) ([]*docs.Request, string) {
	requests := make([]*docs.Request, 0)

	for i, item := range content {
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

					if i == len(content)-1 {
						re := regexp.MustCompile("\\s*[\\r\\n]$")
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
									SegmentId: segmentId,
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
									SegmentId:  segmentId,
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
									SegmentId:  segmentId,
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

func (s *SongService) Style(song entities.Song) (*entities.Song, error) {
	if song.ID == "" {
		return nil, fmt.Errorf("ID is missing for Song: %v", song)
	}

	requests := make([]*docs.Request, 0)

	doc, err := s.docsClient.Documents.Get(song.ID).Do()
	if err != nil {
		return nil, err
	}

	if doc.DocumentStyle.DefaultHeaderId == "" {
		res, err := s.docsClient.Documents.BatchUpdate(song.ID, &docs.BatchUpdateDocumentRequest{
			Requests: []*docs.Request{
				{
					CreateHeader: &docs.CreateHeaderRequest{
						Type: "DEFAULT",
					},
				},
			},
		}).Do()

		if err == nil && res.Replies[0].CreateHeader.HeaderId != "" {
			doc.DocumentStyle.DefaultHeaderId = res.Replies[0].CreateHeader.HeaderId
			_, _ = s.docsClient.Documents.BatchUpdate(song.ID, &docs.BatchUpdateDocumentRequest{
				Requests: []*docs.Request{
					getDefaultHeaderRequest(doc.DocumentStyle.DefaultHeaderId, doc.Title, "", "", ""),
				},
			}).Do()
		}
	}

	doc, err = s.docsClient.Documents.Get(song.ID).Do()
	if err != nil {
		return nil, err
	}

	for _, header := range doc.Headers {
		for j, paragraph := range header.Content {
			if paragraph.Paragraph == nil {
				continue
			}

			style := *paragraph.Paragraph.ParagraphStyle

			if j == 0 || j == 2 {
				paragraph.Paragraph.ParagraphStyle.Alignment = "CENTER"
			}
			if j == 1 {
				paragraph.Paragraph.ParagraphStyle.Alignment = "END"
			}

			requests = append(requests, &docs.Request{
				UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
					Fields:         "*",
					ParagraphStyle: &style,
					Range: &docs.Range{
						StartIndex:      paragraph.StartIndex,
						EndIndex:        paragraph.EndIndex,
						SegmentId:       header.HeaderId,
						ForceSendFields: []string{"StartIndex"},
					},
				},
			})

			for _, element := range paragraph.Paragraph.Elements {

				element.TextRun.TextStyle.WeightedFontFamily = &docs.WeightedFontFamily{
					FontFamily: "Roboto Mono",
				}

				if j == 0 {
					element.TextRun.TextStyle.Bold = true
					element.TextRun.TextStyle.FontSize = &docs.Dimension{
						Magnitude: 20,
						Unit:      "PT",
					}
				}
				if j == 1 {
					element.TextRun.TextStyle.Bold = true
					element.TextRun.TextStyle.FontSize = &docs.Dimension{
						Magnitude: 14,
						Unit:      "PT",
					}
				}
				if j == 2 {
					element.TextRun.TextStyle.Bold = true
					element.TextRun.TextStyle.FontSize = &docs.Dimension{
						Magnitude: 11,
						Unit:      "PT",
					}
				}

				requests = append(requests, &docs.Request{
					UpdateTextStyle: &docs.UpdateTextStyleRequest{
						Fields: "*",
						Range: &docs.Range{
							StartIndex:      element.StartIndex,
							EndIndex:        element.EndIndex,
							SegmentId:       header.HeaderId,
							ForceSendFields: []string{"StartIndex"},
						},
						TextStyle: element.TextRun.TextStyle,
					},
				})
			}
		}

		requests = append(requests, composeStyleRequests(header.Content, header.HeaderId)...)
	}

	requests = append(requests, composeStyleRequests(doc.Body.Content, "")...)

	requests = append(requests, &docs.Request{
		UpdateDocumentStyle: &docs.UpdateDocumentStyleRequest{
			DocumentStyle: &docs.DocumentStyle{
				MarginBottom: &docs.Dimension{
					Magnitude: 14,
					Unit:      "PT",
				},
				MarginHeader: &docs.Dimension{
					Magnitude: 18,
					Unit:      "PT",
				},
				MarginLeft: &docs.Dimension{
					Magnitude: 30,
					Unit:      "PT",
				},
				MarginRight: &docs.Dimension{
					Magnitude: 30,
					Unit:      "PT",
				},
				MarginTop: &docs.Dimension{
					Magnitude: 14,
					Unit:      "PT",
				},
				UseFirstPageHeaderFooter: false,
			},
			Fields: "marginBottom, marginLeft, marginRight, marginTop, marginHeader",
		},
	})

	_, err = s.docsClient.Documents.BatchUpdate(song.ID, &docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	if err != nil {
		return nil, err
	}

	fakeTime, _ := time.Parse("2006", "2006")
	song.PDF.ModifiedTime = fakeTime.Format(time.RFC3339)
	return &song, err
}

func composeStyleRequests(content []*docs.StructuralElement, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0)
	//makeBoldAndRedRegex := regexp.MustCompile(`(x|х)\d+`)
	//sectionNamesRegex := regexp.MustCompile(`\p{L}+(\s\d*)?:|\|`)

	for _, paragraph := range content {
		if paragraph.Paragraph == nil {
			continue
		}

		style := *paragraph.Paragraph.ParagraphStyle

		style.SpaceAbove = &docs.Dimension{
			Magnitude:       0,
			Unit:            "PT",
			ForceSendFields: []string{"Magnitude"},
		}
		style.SpaceBelow = &docs.Dimension{
			Magnitude:       0,
			Unit:            "PT",
			ForceSendFields: []string{"Magnitude"},
		}
		style.LineSpacing = 90

		requests = append(requests, &docs.Request{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Fields:         "*",
				ParagraphStyle: &style,
				Range: &docs.Range{
					EndIndex:        paragraph.EndIndex,
					SegmentId:       segmentID,
					StartIndex:      paragraph.StartIndex,
					ForceSendFields: []string{"StartIndex"},
				},
			},
		})

		for _, element := range paragraph.Paragraph.Elements {
			if element.TextRun == nil {
				continue
			}

			element.TextRun.TextStyle.WeightedFontFamily = &docs.WeightedFontFamily{
				FontFamily: "Roboto Mono",
			}

			requests = append(requests, &docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Fields: "*",
					Range: &docs.Range{
						StartIndex:      element.StartIndex,
						EndIndex:        element.EndIndex,
						SegmentId:       segmentID,
						ForceSendFields: []string{"StartIndex"},
					},
					TextStyle: element.TextRun.TextStyle,
				},
			})

			tokens := transposer.Tokenize(element.TextRun.Content)
			for _, line := range tokens {
				for _, token := range line {
					if token.Chord != nil {
						style := *element.TextRun.TextStyle

						style.Bold = true
						style.ForegroundColor = &docs.OptionalColor{
							Color: &docs.Color{
								RgbColor: &docs.RgbColor{
									Blue:            0,
									Green:           0,
									Red:             0.8,
									ForceSendFields: []string{"blue", "green"},
								},
							},
						}

						requests = append(requests, &docs.Request{
							UpdateTextStyle: &docs.UpdateTextStyleRequest{
								Fields: "*",
								Range: &docs.Range{
									StartIndex:      element.StartIndex + token.Offset,
									EndIndex:        element.StartIndex + token.Offset + int64(len([]rune(token.Chord.String()))),
									SegmentId:       segmentID,
									ForceSendFields: []string{"StartIndex"},
								},
								TextStyle: &style,
							},
						})
					}
				}
			}

			style := *element.TextRun.TextStyle

			style.Bold = true
			style.ForegroundColor = &docs.OptionalColor{
				Color: &docs.Color{
					RgbColor: &docs.RgbColor{Blue: 0, Green: 0, Red: 0, ForceSendFields: []string{"blue", "green", "red"}},
				},
			}

			requests = append(requests, changeStyleByRegex(regexp.MustCompile(`[|]`), *element, style, nil, segmentID)...)

			style = *element.TextRun.TextStyle

			style.Bold = true
			style.ForegroundColor = &docs.OptionalColor{
				Color: &docs.Color{
					RgbColor: &docs.RgbColor{Blue: 0, Green: 0, Red: 0.8, ForceSendFields: []string{"blue", "green"}},
				},
			}

			requests = append(requests, changeStyleByRegex(regexp.MustCompile(`(x|х)\d+`), *element, style, nil, segmentID)...)

			style = *element.TextRun.TextStyle

			style.Bold = true
			style.ForegroundColor = &docs.OptionalColor{
				Color: &docs.Color{
					RgbColor: &docs.RgbColor{Blue: 0, Green: 0, Red: 0, ForceSendFields: []string{"blue", "green", "red"}},
				},
			}
			style.Underline = false
			style.Italic = false
			style.Strikethrough = false

			requests = append(requests, changeStyleByRegex(regexp.MustCompile(`\p{L}+(\s\d*)?:`), *element, style, strings.ToUpper, segmentID)...)

			//matches := makeBoldAndRedRegex.FindAllStringIndex(element.TextRun.Content, -1)
			//if matches != nil {
			//	for _, match := range matches {
			//		style := *element.TextRun.TextStyle
			//
			//		style.Bold = true
			//		style.ForegroundColor = &docs.OptionalColor{
			//			Color: &docs.Color{
			//				RgbColor: &docs.RgbColor{Blue: 0, Green: 0, Red: 0.8, ForceSendFields: []string{"blue", "green"}},
			//			},
			//		}
			//
			//		requests = append(requests, &docs.Request{
			//			UpdateTextStyle: &docs.UpdateTextStyleRequest{
			//				Fields: "*",
			//				Range: &docs.Range{
			//					StartIndex:      element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[0]]))),
			//					EndIndex:        element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[1]]))),
			//					SegmentId:       segmentID,
			//					ForceSendFields: []string{"StartIndex"},
			//				},
			//				TextStyle: &style,
			//			},
			//		})
			//	}
			//}
			//
			//matches = sectionNamesRegex.FindAllStringIndex(element.TextRun.Content, -1)
			//if matches != nil {
			//	for _, match := range matches {
			//		style := *element.TextRun.TextStyle
			//
			//		style.Bold = true
			//		style.ForegroundColor = &docs.OptionalColor{
			//			Color: &docs.Color{
			//				RgbColor: &docs.RgbColor{Blue: 0, Green: 0, Red: 0, ForceSendFields: []string{"blue", "green"}},
			//			},
			//		}
			//		style.Underline = false
			//		style.Italic = false
			//		style.Strikethrough = false
			//
			//		requests = append(requests,
			//			&docs.Request{
			//				UpdateTextStyle: &docs.UpdateTextStyleRequest{
			//					Fields: "*",
			//					Range: &docs.Range{
			//						StartIndex:      element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[0]]))),
			//						EndIndex:        element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[1]]))),
			//						SegmentId:       segmentID,
			//						ForceSendFields: []string{"StartIndex"},
			//					},
			//					TextStyle: &style,
			//				},
			//			},
			//			&docs.Request{
			//				DeleteContentRange: &docs.DeleteContentRangeRequest{
			//					Range: &docs.Range{
			//						StartIndex:      element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[0]]))),
			//						EndIndex:        element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[1]]))),
			//						SegmentId:       segmentID,
			//						ForceSendFields: []string{"StartIndex"},
			//					},
			//				},
			//			},
			//			&docs.Request{
			//				InsertText: &docs.InsertTextRequest{
			//					Location: &docs.Location{
			//						Index:           element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[0]]))),
			//						SegmentId:       segmentID,
			//						ForceSendFields: []string{"StartIndex"},
			//					},
			//					Text: strings.ToUpper(element.TextRun.Content[match[0]:match[1]]),
			//				},
			//			},
			//		)
			//	}
			//}
		}
	}

	return requests
}

func changeStyleByRegex(re *regexp.Regexp, element docs.ParagraphElement, style docs.TextStyle, textFunc func(string) string, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0)

	matches := re.FindAllStringIndex(element.TextRun.Content, -1)
	if matches == nil {
		return requests
	}

	for _, match := range matches {
		requests = append(requests,
			&docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Fields: "*",
					Range: &docs.Range{
						StartIndex:      element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[0]]))),
						EndIndex:        element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[1]]))),
						SegmentId:       segmentID,
						ForceSendFields: []string{"StartIndex"},
					},
					TextStyle: &style,
				},
			},
		)

		if textFunc != nil {
			requests = append(requests,
				&docs.Request{
					DeleteContentRange: &docs.DeleteContentRangeRequest{
						Range: &docs.Range{
							StartIndex:      element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[0]]))),
							EndIndex:        element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[1]]))),
							SegmentId:       segmentID,
							ForceSendFields: []string{"StartIndex"},
						},
					},
				},
				&docs.Request{
					InsertText: &docs.InsertTextRequest{
						Location: &docs.Location{
							Index:           element.StartIndex + int64(len([]rune(element.TextRun.Content[:match[0]]))),
							SegmentId:       segmentID,
							ForceSendFields: []string{"StartIndex"},
						},
						Text: textFunc(element.TextRun.Content[match[0]:match[1]]),
					},
				})
		}
	}

	return requests
}

func getDefaultHeaderRequest(headerID string, name string, key string, BPM string, time string) *docs.Request {

	if name == "" {
		name = "Название - Исполнитель"
	}

	if key == "" {
		key = "?"
	}

	if BPM == "" {
		BPM = "?"
	}

	if time == "" {
		time = "?"
	}

	text := fmt.Sprintf("%s\nKEY: %s; BPM: %s; TIME: %s;\nструктура\n",
		name, key, BPM, time)

	return &docs.Request{
		InsertText: &docs.InsertTextRequest{
			EndOfSegmentLocation: &docs.EndOfSegmentLocation{
				SegmentId: headerID,
			},
			Text: text,
		},
	}
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
