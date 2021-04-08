package services

import (
	"errors"
	"fmt"
	"github.com/flowchartsman/retry"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-chords-bot/helpers"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

const PageSize = 50

type DriveFileService struct {
	driveRepository *drive.Service
	docsRepository  *docs.Service
}

func NewDriveFileService(driveRepository *drive.Service, docsRepository *docs.Service) *DriveFileService {
	return &DriveFileService{
		driveRepository: driveRepository,
		docsRepository:  docsRepository,
	}
}

func (s *DriveFileService) FindAllByFolderID(folderID string, nextPageToken string) ([]*drive.File, string, error) {

	q := fmt.Sprintf(`trashed = false and mimeType = 'application/vnd.google-apps.document' and '%s' in parents`, folderID)

	res, err := s.driveRepository.Files.List().
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, webViewLink, parents)").
		PageSize(PageSize).PageToken(nextPageToken).Do()

	if err != nil {
		return nil, "", err
	}

	return res.Files, res.NextPageToken, nil
}

func (s *DriveFileService) FindSomeByFullTextAndFolderID(name string, folderID string, pageToken string) ([]*drive.File, string, error) {
	name = helpers.JsonEscape(name)

	q := fmt.Sprintf(`fullText contains '%s'`+
		` and trashed = false`+
		` and mimeType = 'application/vnd.google-apps.document'`, name)

	if folderID != "" {
		q += fmt.Sprintf(` and '%s' in parents`, folderID)
	}

	res, err := s.driveRepository.Files.List().
		// Use this for precise search.
		//Q(fmt.Sprintf("fullText contains '\"%s\"'", name)).
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, webViewLink, parents)").
		PageSize(PageSize).PageToken(pageToken).Do()

	if err != nil {
		return nil, "", err
	}

	return res.Files, res.NextPageToken, nil
}

func (s *DriveFileService) FindOneByName(name string, folderID string) (*drive.File, error) {
	name = helpers.JsonEscape(name)

	q := fmt.Sprintf(`name = '%s'`+
		` and trashed = false`+
		` and mimeType = 'application/vnd.google-apps.document'`, name)

	if folderID != "" {
		q += fmt.Sprintf(` and '%s' in parents`, folderID)
	}

	res, err := s.driveRepository.Files.List().
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, webViewLink, parents)").
		PageSize(1).Do()
	if err != nil {
		return nil, err
	}

	if len(res.Files) == 0 {
		return nil, errors.New("not found")
	}

	return res.Files[0], nil
}

func (s *DriveFileService) FindOneByID(ID string) (*drive.File, error) {
	return s.driveRepository.Files.Get(ID).Fields("id, name, modifiedTime, webViewLink, parents").Do()
}

func (s *DriveFileService) CreateOne(newFile *drive.File, lyrics string, key string, BPM string, time string) (*drive.File, error) {
	newFile, err := s.driveRepository.Files.Create(newFile).Do()
	if err != nil {
		return nil, err
	}

	if len(newFile.Parents) > 0 {
		// TODO: use pagination here.
		folderPermissionsList, err := s.driveRepository.Permissions.
			List(newFile.Parents[0]).
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
			_, err = s.driveRepository.Permissions.
				Create(newFile.Id, permission).
				TransferOwnership(true).Do()
			if err != nil {
				return nil, err
			}
		}
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

	res, err := s.docsRepository.Documents.BatchUpdate(newFile.Id,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	if err != nil {
		return nil, err
	}

	if res.Replies[0].CreateHeader.HeaderId != "" {
		_, _ = s.docsRepository.Documents.BatchUpdate(newFile.Id,
			&docs.BatchUpdateDocumentRequest{
				Requests: []*docs.Request{
					getDefaultHeaderRequest(res.Replies[0].CreateHeader.HeaderId, newFile.Name, key, BPM, time),
				},
			}).Do()
	}

	doc, err := s.docsRepository.Documents.Get(newFile.Id).Do()
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

	_, _ = s.docsRepository.Documents.BatchUpdate(newFile.Id,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()

	return s.FindOneByID(newFile.Id)
}

func (s *DriveFileService) CloneOne(fileToCloneID string, newFile *drive.File) (*drive.File, error) {
	newFile, err := s.driveRepository.Files.
		Copy(fileToCloneID, newFile).
		Fields("id, name, modifiedTime, webViewLink, parents").
		Do()
	if err != nil {
		return nil, err
	}

	if len(newFile.Parents) < 1 {
		return newFile, nil
	}

	// TODO: use pagination here.
	folderPermissionsList, err := s.driveRepository.Permissions.
		List(newFile.Parents[0]).
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
		_, err = s.driveRepository.Permissions.
			Create(newFile.Id, permission).
			TransferOwnership(true).Do()
		if err != nil {
			return nil, err
		}
	}

	return newFile, nil
}

func (s *DriveFileService) DownloadOneByID(ID string) (*io.Reader, error) {
	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var reader io.Reader
	err := retrier.Run(func() error {
		res, err := s.driveRepository.Files.Export(ID, "application/pdf").Download()
		if err != nil {
			return err
		}

		reader = res.Body

		return nil
	})

	return &reader, err
}

func (s *DriveFileService) DeleteOneByID(ID string) error {
	return s.driveRepository.Files.Delete(ID).Do()
}

func (s *DriveFileService) TransposeOne(ID string, toKey string, sectionIndex int) (*drive.File, error) {
	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return nil, err
	}

	sections := s.getSections(doc)

	if len(sections) <= sectionIndex || sectionIndex < 0 {
		sections, err = s.appendSectionByID(ID)
		if err != nil {
			return nil, err
		}

		doc, err = s.docsRepository.Documents.Get(ID).Do()
		if err != nil {
			return nil, err
		}

		sectionIndex = len(sections) - 1
	}

	requests, key := s.transposeHeader(doc, sections, sectionIndex, toKey)
	requests = append(requests, s.transposeBody(doc, sections, sectionIndex, key, toKey)...)

	_, err = s.docsRepository.Documents.BatchUpdate(doc.DocumentId,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()

	return s.FindOneByID(ID)
}

func (s *DriveFileService) StyleOne(ID string) (*drive.File, error) {
	requests := make([]*docs.Request, 0)

	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return nil, err
	}

	if doc.DocumentStyle.DefaultHeaderId == "" {
		res, err := s.docsRepository.Documents.BatchUpdate(ID, &docs.BatchUpdateDocumentRequest{
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
			_, _ = s.docsRepository.Documents.BatchUpdate(ID, &docs.BatchUpdateDocumentRequest{
				Requests: []*docs.Request{
					getDefaultHeaderRequest(doc.DocumentStyle.DefaultHeaderId, doc.Title, "", "", ""),
				},
			}).Do()
		}
	}

	doc, err = s.docsRepository.Documents.Get(ID).Do()
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

	_, err = s.docsRepository.Documents.BatchUpdate(ID, &docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

func (s *DriveFileService) GetSectionsNumber(ID string) (int, error) {
	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return 0, err
	}

	return len(s.getSections(doc)), nil
}

func (s *DriveFileService) GetMetadata(ID string) (string, string, string) {
	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var reader io.Reader
	err := retrier.Run(func() error {
		res, err := s.driveRepository.Files.Export(ID, "text/plain").Download()
		if err != nil {
			return err
		}

		reader = res.Body

		return nil
	})

	var driveFileText string
	b, err := ioutil.ReadAll(reader)
	if err == nil {
		driveFileText = string(b)
	}

	key := ""
	keyRegex := regexp.MustCompile(`(?i)key:(.*?);`)
	keyMatches := keyRegex.FindStringSubmatch(driveFileText)
	if len(keyMatches) > 1 {
		key = strings.TrimSpace(keyMatches[1])
	}

	BPM := ""
	BPMRegex := regexp.MustCompile(`(?i)bpm:(.*?);`)
	BPMMatches := BPMRegex.FindStringSubmatch(driveFileText)
	if len(BPMMatches) > 1 {
		BPM = strings.TrimSpace(BPMMatches[1])
	}

	time := ""
	timeRegex := regexp.MustCompile(`(?i)time:(.*?);`)
	timeMatches := timeRegex.FindStringSubmatch(driveFileText)
	if len(timeMatches) > 1 {
		time = strings.TrimSpace(timeMatches[1])
	}

	return key, BPM, time
}

//
// Helper functions. -----------------------------------
//
func (s *DriveFileService) getSections(doc *docs.Document) []docs.StructuralElement {
	sections := make([]docs.StructuralElement, 0)

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

	return sections
}

func (s *DriveFileService) appendSectionByID(ID string) ([]docs.StructuralElement, error) {
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

	_, err := s.docsRepository.Documents.BatchUpdate(ID, requests).Do()
	if err != nil {
		return nil, err
	}

	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return nil, err
	}

	sections := s.getSections(doc)

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

	_, err = s.docsRepository.Documents.BatchUpdate(ID, requests).Do()
	if err != nil {
		return nil, err
	}

	doc, err = s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return nil, err
	}
	return s.getSections(doc), nil
}

func (s *DriveFileService) transposeHeader(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, toKey string) ([]*docs.Request, string) {
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

	transposeRequests, key := composeTransposeRequests(doc.Headers[doc.DocumentStyle.DefaultHeaderId].Content,
		0, "", toKey, doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId].HeaderId)
	requests = append(requests, transposeRequests...)

	return requests, key
}

func (s *DriveFileService) transposeBody(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, key string, toKey string) []*docs.Request {
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

	transposeRequests, _ := composeTransposeRequests(content, sectionToInsertStartIndex, key, toKey, "")
	requests = append(requests, transposeRequests...)

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
