package services

import (
	"errors"
	"fmt"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
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
func (s *SongService) QueryDrive(name string, pageToken string, folderIDs ...string) ([]entities.Song, string, error) {
	var songs []entities.Song

	q := fmt.Sprintf("fullText contains '%s'"+
		" and mimeType = 'application/vnd.google-apps.document'", name)

	if folderIDs != nil && len(folderIDs) > 0 {
		for i := range folderIDs {
			if i == 0 {
				q += " and "
			}

			q += fmt.Sprintf("'%s' in parents", folderIDs[i])

			if i != len(folderIDs)-1 {
				q += " or "
			}
		}
	}

	res, err := s.driveClient.Files.List().
		// Use this for precise search.
		//Q(fmt.Sprintf("fullText contains '\"%s\"'", name)).
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, webViewLink)").
		PageSize(90).
		PageToken(pageToken).
		Do()

	if err != nil {
		return nil, "", err
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

	if len(songs) == 0 {
		return nil, "", mongo.ErrEmptySlice
	}

	return songs, res.NextPageToken, nil
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

func (s *SongService) Cache(song entities.Song) (entities.Song, error) {
	if song.ID == "" {
		return song, fmt.Errorf("ID is missing for Song: %v", song)
	}

	oldSong, err := s.FindOneByID(song.ID)
	if err != nil {
		return s.UpdateOne(song)
	}

	song.Voices = oldSong.Voices
	return s.UpdateOne(song)
}

/*
Returns error if cached version is outdated.
*/
func (s *SongService) GetFromCache(song entities.Song) (entities.Song, error) {
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
		return entities.Song{}, errors.New("cached song is not actual")
	}

	cachedSong.Name = song.Name
	cachedSong.WebViewLink = song.WebViewLink

	return cachedSong, err
}

func (s *SongService) DownloadPDF(song entities.Song) (tgbotapi.FileReader, error) {
	if song.ID == "" {
		return tgbotapi.FileReader{}, fmt.Errorf("ID is missing for Song: %v", song)
	}

	res, err := s.driveClient.Files.Export(song.ID, "application/pdf").Download()
	if err != nil {
		return tgbotapi.FileReader{}, err
	}

	fileReader := tgbotapi.FileReader{
		Name:   song.Name + ".pdf",
		Reader: res.Body,
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

	requests, key := s.transposeHeader(doc, sections, sectionIndex, toKey)
	requests = append(requests, s.transposeBody(doc, sections, sectionIndex, key, toKey)...)

	_, err = s.docsClient.Documents.BatchUpdate(doc.DocumentId,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()

	song.ModifiedTime = time.Now().UTC().Format(time.RFC3339)
	return song, err
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

func (s *SongService) Style(song entities.Song) (entities.Song, error) {
	if song.ID == "" {
		return song, fmt.Errorf("ID is missing for Song: %v", song)
	}

	requests := make([]*docs.Request, 0)

	doc, err := s.docsClient.Documents.Get(song.ID).Do()
	if err != nil {
		return entities.Song{}, err
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

		if err == nil {
			doc.DocumentStyle.DefaultHeaderId = res.Replies[0].CreateHeader.HeaderId
			_, _ = s.docsClient.Documents.BatchUpdate(song.ID, &docs.BatchUpdateDocumentRequest{
				Requests: []*docs.Request{
					{
						InsertText: &docs.InsertTextRequest{
							EndOfSegmentLocation: &docs.EndOfSegmentLocation{
								SegmentId: doc.DocumentStyle.DefaultHeaderId,
							},
							Text: "Название - Исполнитель\n" +
								"KEY: ?; BPM: ?; TIME: ?;\n" +
								"структура\n",
						},
					},
				},
			}).Do()
		}
	}

	doc, err = s.docsClient.Documents.Get(song.ID).Do()
	if err != nil {
		return entities.Song{}, err
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
				if element.TextRun.TextStyle.WeightedFontFamily != nil {
					element.TextRun.TextStyle.WeightedFontFamily.FontFamily = "Roboto Mono"
				} else {
					element.TextRun.TextStyle.WeightedFontFamily = &docs.WeightedFontFamily{
						FontFamily: "Roboto Mono",
					}
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
		return entities.Song{}, err
	}

	song.ModifiedTime = time.Now().UTC().Format(time.RFC3339)
	return song, err
}

func composeStyleRequests(content []*docs.StructuralElement, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0)

	makeBoldAndRedRegex := regexp.MustCompile(`(x|х)\d+`)
	sectionNamesRegex := regexp.MustCompile(`\p{L}+(\s\d*)?:|\|`)

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
			style := *element.TextRun.TextStyle
			if style.WeightedFontFamily != nil {
				style.WeightedFontFamily.FontFamily = "Roboto Mono"
			} else {
				style.WeightedFontFamily = &docs.WeightedFontFamily{
					FontFamily: "Roboto Mono",
				}
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
					TextStyle: &style,
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

			matches := makeBoldAndRedRegex.FindAllStringIndex(element.TextRun.Content, -1)
			if matches != nil {
				for _, match := range matches {
					style := *element.TextRun.TextStyle
					style.Bold = true
					style.ForegroundColor = &docs.OptionalColor{
						Color: &docs.Color{
							RgbColor: &docs.RgbColor{Blue: 0, Green: 0, Red: 0.8, ForceSendFields: []string{"blue", "green"}},
						},
					}

					requests = append(requests, &docs.Request{
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
					})
				}
			}

			matches = sectionNamesRegex.FindAllStringIndex(element.TextRun.Content, -1)
			if matches != nil {
				for _, match := range matches {
					style := *element.TextRun.TextStyle
					style.Bold = true
					style.ForegroundColor = &docs.OptionalColor{
						Color: &docs.Color{
							RgbColor: &docs.RgbColor{Blue: 0, Green: 0, Red: 0, ForceSendFields: []string{"blue", "green"}},
						},
					}
					style.Underline = false
					style.Italic = false
					style.Strikethrough = false

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
								Text: strings.ToUpper(element.TextRun.Content[match[0]:match[1]]),
							},
						},
					)
				}
			}
		}
	}

	return requests
}
