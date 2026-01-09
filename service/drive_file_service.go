package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/flowchartsman/retry"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/helpers"
	"github.com/joeyave/scala-bot/txt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

type DriveFileService struct {
	driveClient    *drive.Service
	docsRepository *docs.Service
}

func NewDriveFileService(driveRepository *drive.Service, docsRepository *docs.Service) *DriveFileService {
	return &DriveFileService{
		driveClient:    driveRepository,
		docsRepository: docsRepository,
	}
}

var newLinesRegex = regexp.MustCompile(`\n{3,}`)

func (s *DriveFileService) FindAllByFolderID(folderID string, nextPageToken string) ([]*drive.File, string, error) {

	q := fmt.Sprintf(`trashed = false and mimeType = 'application/vnd.google-apps.document' and '%s' in parents`, folderID)

	res, err := s.driveClient.Files.List().
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, webViewLink, parents)").
		PageSize(helpers.SongsPageSize).PageToken(nextPageToken).Do()

	if err != nil {
		return nil, "", err
	}

	return res.Files, res.NextPageToken, nil
}

func (s *DriveFileService) FindSomeByFullTextAndFolderID(name string, folderIDs []string, pageToken string) ([]*drive.File, string, error) {
	name = helpers.JsonEscape(name)

	q := fmt.Sprintf(`fullText contains '%s'`+
		` and trashed = false`+
		` and mimeType = 'application/vnd.google-apps.document'`, name)

	var folderIDsCleaned []string
	for _, folderID := range folderIDs {
		if folderID != "" {
			folderIDsCleaned = append(folderIDsCleaned, folderID)
		}
	}
	if len(folderIDsCleaned) != 0 {
		qBuilder := strings.Builder{}
		for i, folderID := range folderIDsCleaned {
			qBuilder.WriteString(fmt.Sprintf("'%s' in parents or '%s' in parents", folderID, folderID))
			if i < len(folderIDsCleaned)-1 {
				qBuilder.WriteString(" or ")
			}
		}
		q += fmt.Sprintf(` and (%s)`, qBuilder.String())
	}

	res, err := s.driveClient.Files.List().
		// Use this for precise search.
		// Q(fmt.Sprintf("fullText contains '\"%s\"'", name)).
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, webViewLink, parents)").
		PageSize(helpers.SongsPageSize).PageToken(pageToken).Do()

	if err != nil {
		return nil, "", err
	}

	return res.Files, res.NextPageToken, nil
}

func (s *DriveFileService) FindOneByNameAndFolderID(name string, folderIDs []string) (*drive.File, error) {
	name = helpers.JsonEscape(name)

	q := fmt.Sprintf(`name = '%s'`+
		` and trashed = false`+
		` and mimeType = 'application/vnd.google-apps.document'`, name)

	var folderIDsCleaned []string
	for _, folderID := range folderIDs {
		if folderID != "" {
			folderIDsCleaned = append(folderIDsCleaned, folderID)
		}
	}
	if len(folderIDsCleaned) != 0 {
		qBuilder := strings.Builder{}
		for i, folderID := range folderIDsCleaned {
			qBuilder.WriteString(fmt.Sprintf("'%s' in parents or '%s' in parents", folderID, folderID))
			if i < len(folderIDsCleaned)-1 {
				qBuilder.WriteString(" or ")
			}
		}
		q += fmt.Sprintf(` and (%s)`, qBuilder.String())
	}

	res, err := s.driveClient.Files.List().
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
	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var driveFile *drive.File
	err := retrier.Run(func() error {
		_driveFile, err := s.driveClient.Files.Get(ID).Fields("id, name, modifiedTime, webViewLink, parents").Do()
		if err != nil {
			return err
		}

		driveFile = _driveFile
		return nil
	})

	return driveFile, err
}

func (s *DriveFileService) FindManyByIDs(IDs []string) ([]*drive.File, error) {

	errwg := new(errgroup.Group)
	driveFiles := make([]*drive.File, len(IDs))
	for i := range IDs {
		i := i
		errwg.Go(func() error {
			driveFile, err := s.FindOneByID(IDs[i])
			if err == nil {
				driveFiles[i] = driveFile
			}
			return err
		})
	}
	err := errwg.Wait()

	return driveFiles, err
}

func (s *DriveFileService) CreateOne(newFile *drive.File, lyrics string, key entity.Key, BPM string, time, lang string) (*drive.File, error) {
	newFile, err := s.driveClient.Files.
		Create(newFile).
		Fields("id, name, modifiedTime, webViewLink, parents").
		Do()
	if err != nil {
		return nil, err
	}

	if len(newFile.Parents) > 0 {
		// TODO: use pagination here.
		folderPermissionsList, err := s.driveClient.Permissions.
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

		// https://stackoverflow.com/questions/71749779/consent-is-required-to-transfer-ownership-of-a-file-to-another-user-google-driv
		// https://developers.google.com/drive/api/guides/manage-sharing
		if folderOwnerPermission != nil {
			permission := &drive.Permission{
				EmailAddress: folderOwnerPermission.EmailAddress,
				Role:         "writer",
				PendingOwner: true,
				Type:         "user",
			}
			_, _ = s.driveClient.Permissions.
				Create(newFile.Id, permission).
				TransferOwnership(false).Do()
			//if err != nil {
			//	return nil, err
			//}
		}
	}

	requests := make([]*docs.Request, 0)

	requests = append(requests, &docs.Request{
		CreateHeader: &docs.CreateHeaderRequest{
			Type: "DEFAULT",
		},
	})

	// todo: test and include.
	//requests = append(requests, &docs.Request{
	//	UpdateDocumentStyle: &docs.UpdateDocumentStyleRequest{
	//		DocumentStyle: &docs.DocumentStyle{
	//			PageSize: &docs.Size{
	//				Width: &docs.Dimension{
	//					Magnitude: 595.276, // A4 width in points
	//					Unit:      "PT",
	//				},
	//				Height: &docs.Dimension{
	//					Magnitude: 841.890, // A4 height in points
	//					Unit:      "PT",
	//				},
	//			},
	//		},
	//		Fields: "*",
	//	},
	//})

	if lyrics != "" {
		requests = append(requests, &docs.Request{
			InsertText: &docs.InsertTextRequest{
				EndOfSegmentLocation: &docs.EndOfSegmentLocation{
					SegmentId: "",
				},
				Text: lyrics,
			},
		})
	}

	res, err := s.docsRepository.Documents.BatchUpdate(newFile.Id,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	if err != nil {
		return nil, err
	}

	if res.Replies[0].CreateHeader.HeaderId != "" {
		_, _ = s.docsRepository.Documents.BatchUpdate(newFile.Id,
			&docs.BatchUpdateDocumentRequest{
				Requests: []*docs.Request{
					getDefaultHeaderRequest(res.Replies[0].CreateHeader.HeaderId, newFile.Name, key, BPM, time, lang),
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
	newFile, err := s.driveClient.Files.
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
	folderPermissionsList, err := s.driveClient.Permissions.
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
			Role:         "writer",
			PendingOwner: true,
			Type:         "user",
		}
		_, err = s.driveClient.Permissions.
			Create(newFile.Id, permission).
			TransferOwnership(false).Do()
		if err != nil {
			return nil, err
		}
	}

	return newFile, nil
}

func (s *DriveFileService) FindOrCreateOneFolderByNameAndFolderID(name string, folderID string) (*drive.File, error) {
	name = helpers.JsonEscape(name)

	q := fmt.Sprintf(`name = '%s'`+
		` and trashed = false`+
		` and mimeType = 'application/vnd.google-apps.folder'`, name)

	if folderID != "" {
		q += fmt.Sprintf(` and '%s' in parents`, folderID)
	}

	res, err := s.driveClient.Files.List().
		Q(q).
		Fields("nextPageToken, files(id, name, modifiedTime, parents)").
		PageSize(1).Do()
	if err != nil {
		return nil, err
	}

	if len(res.Files) == 0 {
		return s.driveClient.Files.Create(&drive.File{
			Name:     name,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{folderID},
		}).Do()
	}

	return res.Files[0], nil
}

func (s *DriveFileService) MoveOne(fileID string, newFolderID string) (*drive.File, error) {
	file, err := s.driveClient.Files.Get(fileID).Fields("parents").Do()
	if err != nil {
		return nil, err
	}

	previousParents := strings.Join(file.Parents, ",")

	newFile, err := s.driveClient.Files.Update(fileID, nil).
		AddParents(newFolderID).RemoveParents(previousParents).Fields("id, name, modifiedTime, webViewLink, parents").Do()
	if err != nil {
		return nil, err
	}
	return newFile, err
}

// DeleteOne deletes a file from Google Drive by its ID
func (s *DriveFileService) DeleteOne(fileID string) error {
	return s.driveClient.Files.Delete(fileID).Do()
}

func (s *DriveFileService) DownloadOneByID(ID string) (io.ReadCloser, error) {
	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var reader io.ReadCloser
	err := retrier.Run(func() error {
		res, err := s.driveClient.Files.Export(ID, "application/pdf").Download()
		if err != nil {
			return err
		}

		reader = res.Body

		return nil
	})

	return reader, err
}

func (s *DriveFileService) DownloadOneByIDWithResp(ID string) (*http.Response, error) {
	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var reader *http.Response
	err := retrier.Run(func() error {
		res, err := s.driveClient.Files.Export(ID, "application/pdf").Download()
		if err != nil {
			return err
		}
		reader = res
		return nil
	})

	return reader, err
}

func (s *DriveFileService) AddLyricsPage(ID string) (*drive.File, error) {
	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return nil, err
	}

	sections := getSections(doc)

	if len(sections) == 1 {
		sections, err = s.appendSectionByID(ID)
		if err != nil {
			return nil, err
		}

		doc, err = s.docsRepository.Documents.Get(ID).Do()
		if err != nil {
			return nil, err
		}
	}

	requests := removeChords(doc, sections, 1)

	_, err = s.docsRepository.Documents.BatchUpdate(doc.DocumentId,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

func (s *DriveFileService) Rename(ID string, newName string) error {
	_, err := s.driveClient.Files.Update(ID, &drive.File{Name: newName}).Do()
	return err
}

func (s *DriveFileService) ReplaceAllTextByRegex(ID string, regex *regexp.Regexp, replaceText string) (int64, error) {
	res, err := s.driveClient.Files.Export(ID, "text/plain").Download()
	if err != nil {
		return 0, err
	}

	var driveFileText string
	b, err := io.ReadAll(res.Body)
	if err == nil {
		driveFileText = string(b)
	}

	textToReplace := regex.FindString(driveFileText)

	request := &docs.BatchUpdateDocumentRequest{Requests: []*docs.Request{
		{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					MatchCase: true,
					Text:      textToReplace,
				},
				ReplaceText: replaceText,
			},
		},
	}}

	replaceAllTextResp, err := s.docsRepository.Documents.BatchUpdate(ID, request).Do()
	if err != nil {
		return 0, err
	}

	return replaceAllTextResp.Replies[0].ReplaceAllText.OccurrencesChanged, err
}

func (s *DriveFileService) GetSectionsNumber(ID string) (int, error) {
	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return 0, err
	}

	return len(getSections(doc)), nil
}

func (s *DriveFileService) GetMetadata(ID string) (entity.Key, string, string) {
	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var reader io.Reader
	err := retrier.Run(func() error {
		res, err := s.driveClient.Files.Export(ID, "text/plain").Download()
		if err != nil {
			return err
		}

		reader = res.Body

		return nil
	})
	if err != nil {
		return "", "", ""
	}

	var driveFileText string
	b, err := io.ReadAll(reader)
	if err == nil {
		driveFileText = string(b)
	}

	key := "?"
	keyRegex := regexp.MustCompile(`(?i)key:(.*?);`)
	keyMatches := keyRegex.FindStringSubmatch(driveFileText)
	if len(keyMatches) > 1 {
		keyTrimmed := strings.TrimSpace(keyMatches[1])
		if keyTrimmed != "" {
			key = keyTrimmed
		}
	}

	BPM := "?"
	BPMRegex := regexp.MustCompile(`(?i)bpm:(.*?);`)
	BPMMatches := BPMRegex.FindStringSubmatch(driveFileText)
	if len(BPMMatches) > 1 {
		BPMTrimmed := strings.TrimSpace(BPMMatches[1])
		if BPMTrimmed != "" {
			BPM = BPMTrimmed
		}
	}

	time := "?"
	timeRegex := regexp.MustCompile(`(?i)time:(.*?);`)
	timeMatches := timeRegex.FindStringSubmatch(driveFileText)
	if len(timeMatches) > 1 {
		timeTrimmed := strings.TrimSpace(timeMatches[1])
		if timeTrimmed != "" {
			time = timeTrimmed
		}
	}

	return entity.Key(key), BPM, time
}

func (s *DriveFileService) GetHTMLTextWithSectionsNumber(ID string) (string, int) {

	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return "", 0
	}

	return docToHTML(doc), len(getSections(doc))
}

func (s *DriveFileService) GetLyrics(ID string) (string, error) {

	retrier := retry.NewRetrier(5, 100*time.Millisecond, time.Second)

	var reader io.Reader
	err := retrier.Run(func() error {
		res, err := s.driveClient.Files.Export(ID, "text/plain").Download()
		if err != nil {
			return err
		}

		reader = res.Body

		return nil
	})
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
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

	sections := getSections(doc)

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
	return getSections(doc), nil
}

func docToHTML(doc *docs.Document) string {
	var sb strings.Builder

	for _, item := range doc.Body.Content {
		if item.SectionBreak != nil && item.SectionBreak.SectionStyle != nil && item.SectionBreak.SectionStyle.SectionType == "NEXT_PAGE" {
			break
		}

		if item.Paragraph != nil && item.Paragraph.Elements != nil {
			for _, element := range item.Paragraph.Elements {
				if element.TextRun != nil && element.TextRun.Content != "" {
					style := element.TextRun.TextStyle
					text := element.TextRun.Content

					if style != nil {
						if style.Bold {
							text = fmt.Sprintf("<b>%s</b>", text)
						}
						if style.Italic {
							text = fmt.Sprintf("<i>%s</i>", text)
						}
						if style.ForegroundColor != nil && style.ForegroundColor.Color != nil && style.ForegroundColor.Color.RgbColor != nil {
							text = fmt.Sprintf(`<span class="chord">%s</span>`, text)
							//text = fmt.Sprintf(`<span class="chord" style="color: rgb(%d%%, %d%%, %d%%)">%s</span>`, int(style.ForegroundColor.Color.RgbColor.Red*100), int(style.ForegroundColor.Color.RgbColor.Green*100), int(style.ForegroundColor.Color.RgbColor.Blue*100), text)
						}
					}

					sb.WriteString(text)
				}
			}
		}
	}

	text := newLinesRegex.ReplaceAllString(sb.String(), "\n\n")
	return strings.TrimSpace(text)
}

func removeChords(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int) []*docs.Request {
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

	bodyCloneRequests := composeCloneWithoutChordsRequests(content, sectionToInsertStartIndex, "")
	requests = append(requests, bodyCloneRequests...)

	if doc.DocumentStyle.DefaultHeaderId == "" {
		return requests
	}

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

	headerCloneRequests := composeCloneWithoutChordsRequests(doc.Headers[doc.DocumentStyle.DefaultHeaderId].Content, 0, doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId].HeaderId)

	requests = append(requests, headerCloneRequests...)

	return requests
}

func composeCloneWithoutChordsRequests(content []*docs.StructuralElement, index int64, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0)

	for _, item := range content {
		if item.Paragraph != nil && item.Paragraph.Elements != nil {
			for _, element := range item.Paragraph.Elements {

				var sb strings.Builder
				for _, element := range item.Paragraph.Elements {
					if element.TextRun != nil && element.TextRun.Content != "" {
						sb.WriteString(element.TextRun.Content)
					}
				}
				_, err := transposer.GuessKeyFromText(sb.String())
				if err == nil && segmentID == "" {
					continue
				}

				if element.TextRun != nil && element.TextRun.Content != "" {

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
									SegmentId: segmentID,
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
									SegmentId:  segmentID,
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
									SegmentId:  segmentID,
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

	return requests
}

func getSections(doc *docs.Document) []docs.StructuralElement {
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

func getDefaultHeaderRequest(headerID, name string, key entity.Key, BPM, time, lang string) *docs.Request {

	if name == "" {
		// TODO: replace "uk" with user language if available in context
		name = txt.Get("text.defaultDocTitle", lang)
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

	text := fmt.Sprintf("%s\nKEY: %s; BPM: %s; TIME: %s;\n",
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
