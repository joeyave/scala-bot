package service

import (
	"errors"
	"fmt"
	"github.com/flowchartsman/retry"
	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-bot/helpers"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
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

func (s *DriveFileService) CreateOne(newFile *drive.File, lyrics string, key string, BPM string, time string) (*drive.File, error) {
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
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

func (s *DriveFileService) TransposeHeader(ID string, toKey string, sectionIndex int) (*drive.File, error) {
	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return nil, err
	}

	sections := s.getSections(doc)

	if len(sections) <= sectionIndex || sectionIndex < 0 {
		return nil, fmt.Errorf("section index %d is out of bounds", sectionIndex)
	}

	requests, _ := s.transposeHeader(doc, sections, sectionIndex, toKey)

	_, err = s.docsRepository.Documents.BatchUpdate(doc.DocumentId,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

func (s *DriveFileService) AddLyricsPage(ID string) (*drive.File, error) {
	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return nil, err
	}

	sections := s.getSections(doc)

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

	requests := s.removeChords(doc, sections, 1)

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
					Fields:         "alignment",
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

	transposeRequests, key := composeTransposeRequests(doc.Headers[doc.DocumentStyle.DefaultHeaderId].Content, 0, "", toKey, doc.Headers[sections[sectionIndex].SectionBreak.SectionStyle.DefaultHeaderId].HeaderId, true)
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

	transposeRequests, _ := composeTransposeRequests(content, sectionToInsertStartIndex, key, toKey, "", false)
	requests = append(requests, transposeRequests...)

	return requests
}

func (s *DriveFileService) removeChords(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int) []*docs.Request {
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

var keyRegex = regexp.MustCompile(`(?i)key:(.*?);`)

func composeTransposeRequests(content []*docs.StructuralElement, index int64, key string, toKey string, segmentId string, replaceKeyOnErr bool) ([]*docs.Request, string) {

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
					} else if replaceKeyOnErr && keyRegex.FindString(element.TextRun.Content) != "" {
						element.TextRun.Content = keyRegex.ReplaceAllStringFunc(element.TextRun.Content, func(match string) string {
							// match == "KEY:oldValue;"  or  "key:another;"
							idx := keyRegex.FindStringSubmatchIndex(match)
							// idx = [ fullStart, fullEnd, groupStart, groupEnd ]
							prefix := match[:idx[2]] // up through the ':'
							suffix := match[idx[3]:] // from ';' onward
							return fmt.Sprintf("%s %s%s", strings.TrimSpace(prefix), strings.TrimSpace(toKey), strings.TrimSpace(suffix))
						})
					}

					if i == len(content)-1 {
						re := regexp.MustCompile(`\s*[\r\n]$`)
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

func (s *DriveFileService) GetHTMLTextWithSectionsNumber(ID string) (string, int) {

	doc, err := s.docsRepository.Documents.Get(ID).Do()
	if err != nil {
		return "", 0
	}

	return docToHTML(doc), len(s.getSections(doc))
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

var newLinesRegex = regexp.MustCompile(`\n{3,}`)

func composeStyleRequests(content []*docs.StructuralElement, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0)
	// makeBoldAndRedRegex := regexp.MustCompile(`(x|х)\d+`)
	// sectionNamesRegex := regexp.MustCompile(`\p{L}+(\s\d*)?:|\|`)

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
				Fields:         "alignment, lineSpacing, direction, spaceAbove, spaceBelow",
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

			// requests = append(requests, changeStyleByRegex(regexp.MustCompile(`\p{L}+(\s\d*)?:`), *element, style, strings.ToUpper, segmentID)...)
			requests = append(requests, changeStyleByRegex(regexp.MustCompile(`^[\d\s]*\p{L}+(\s\d*)?:`), *element, style, strings.ToUpper, segmentID)...)
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
