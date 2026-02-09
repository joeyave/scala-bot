package service

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/joeyave/scala-bot/entity"
	"google.golang.org/api/docs/v1"
)

var metadataLineRegex = regexp.MustCompile(`(?i)^\s*key:\s*(.*?);\s*bpm:\s*(.*?);\s*time:\s*(.*?);\s*$`)

type SectionMetadata struct {
	Title string
	Key   entity.Key
	BPM   string
	Time  string
}

type MetadataPatch struct {
	Title *string
	Key   *entity.Key
	BPM   *string
	Time  *string
}

type MetadataNormalizeResult struct {
	SectionsNormalized int
	HeadersDeleted     int
}

type normalizeMetadataLayoutOptions struct {
	applyMetadataStyles bool
}

const (
	metadataAlignmentCenter = "CENTER"
	metadataAlignmentEnd    = "END"

	metadataFontSizeTitle    float64 = 20
	metadataFontSizeLine     float64 = 14
	metadataFontSizeLastLine float64 = 11
)

func normalizeTextValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "?"
	}
	return s
}

func normalizeComparableText(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeMetadata(md SectionMetadata) SectionMetadata {
	md.Title = normalizeTextValue(md.Title)
	md.Key = entity.Key(normalizeTextValue(string(md.Key)))
	md.BPM = normalizeTextValue(md.BPM)
	md.Time = normalizeTextValue(md.Time)
	return md
}

func composeCanonicalMetadataText(md SectionMetadata) string {
	md = normalizeMetadata(md)
	return fmt.Sprintf("%s\nKEY: %s; BPM: %s; TIME: %s;\n\n", md.Title, md.Key, md.BPM, md.Time)
}

func composeMetadataTextWithoutTrailingEmpty(md SectionMetadata) string {
	md = normalizeMetadata(md)
	return fmt.Sprintf("%s\nKEY: %s; BPM: %s; TIME: %s;\n", md.Title, md.Key, md.BPM, md.Time)
}

func parseMetadataLine(text string) (entity.Key, string, string, bool) {
	matches := metadataLineRegex.FindStringSubmatch(strings.TrimSpace(text))
	if len(matches) != 4 {
		return "", "", "", false
	}
	return entity.Key(normalizeTextValue(matches[1])), normalizeTextValue(matches[2]), normalizeTextValue(matches[3]), true
}

func paragraphToPlainText(paragraph *docs.StructuralElement) string {
	if paragraph == nil || paragraph.Paragraph == nil {
		return ""
	}

	var sb strings.Builder
	for _, el := range paragraph.Paragraph.Elements {
		if el.TextRun != nil {
			sb.WriteString(el.TextRun.Content)
		}
	}
	return strings.TrimSpace(sb.String())
}

func paragraphToRawText(paragraph *docs.StructuralElement) string {
	if paragraph == nil || paragraph.Paragraph == nil {
		return ""
	}

	var sb strings.Builder
	for _, el := range paragraph.Paragraph.Elements {
		if el.TextRun != nil {
			sb.WriteString(el.TextRun.Content)
		}
	}
	return sb.String()
}

func paragraphToExactLineText(paragraph *docs.StructuralElement) string {
	text := paragraphToRawText(paragraph)
	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimSuffix(text, "\r")
	return text
}

func contentEndForSection(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int) int64 {
	if len(sections) > sectionIndex+1 {
		return sections[sectionIndex+1].StartIndex - 1
	}
	if len(doc.Body.Content) == 0 {
		return 0
	}
	return doc.Body.Content[len(doc.Body.Content)-1].EndIndex - 1
}

func findFirstContinuousBreakInRange(doc *docs.Document, start, end int64) *docs.StructuralElement {
	for _, item := range doc.Body.Content {
		if item.SectionBreak == nil || item.SectionBreak.SectionStyle == nil {
			continue
		}
		if item.StartIndex <= start || item.StartIndex >= end {
			continue
		}
		if item.SectionBreak.SectionStyle.SectionType == "CONTINUOUS" {
			copyItem := item
			return copyItem
		}
	}
	return nil
}

func paragraphElementsInRange(doc *docs.Document, start, end int64) []*docs.StructuralElement {
	result := make([]*docs.StructuralElement, 0)
	for _, item := range doc.Body.Content {
		if item.Paragraph == nil {
			continue
		}
		if item.StartIndex < start || item.EndIndex > end {
			continue
		}
		result = append(result, item)
	}
	return result
}

func copyParagraphElements(elements []*docs.StructuralElement) []*docs.StructuralElement {
	if len(elements) == 0 {
		return nil
	}
	copied := make([]*docs.StructuralElement, 0, len(elements))
	for _, el := range elements {
		if el == nil || el.Paragraph == nil {
			continue
		}
		copied = append(copied, el)
	}
	return copied
}

func keepTailParagraphsForBody(elements []*docs.StructuralElement) []*docs.StructuralElement {
	if len(elements) == 0 {
		return nil
	}
	hasNonEmpty := false
	for _, item := range elements {
		if strings.TrimSpace(paragraphToPlainText(item)) != "" {
			hasNonEmpty = true
			break
		}
	}
	if !hasNonEmpty {
		return nil
	}
	return copyParagraphElements(elements)
}

func cloneParagraphElementsRequests(content []*docs.StructuralElement, index int64, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0)

	for _, item := range content {
		if item == nil || item.Paragraph == nil || item.Paragraph.Elements == nil {
			continue
		}
		for _, element := range item.Paragraph.Elements {
			if element.TextRun == nil || element.TextRun.Content == "" {
				continue
			}
			if element.TextRun.TextStyle == nil {
				element.TextRun.TextStyle = &docs.TextStyle{}
			}
			if element.TextRun.TextStyle.ForegroundColor == nil {
				element.TextRun.TextStyle.ForegroundColor = newOptionalColor(rgbColorBlack)
			}

			textLen := int64(len([]rune(element.TextRun.Content)))
			endIndex := index + textLen
			requests = append(requests,
				newInsertTextRequest(element.TextRun.Content, index, segmentID),
				newUpdateTextStyleRequest(element.TextRun.TextStyle, "*", index, endIndex, segmentID),
				newUpdateParagraphStyleRequest(item.Paragraph.ParagraphStyle, "alignment,lineSpacing,direction,spaceAbove,spaceBelow", index, endIndex, segmentID),
			)
			index = endIndex
		}
	}
	return requests
}

func metadataFromHeader(doc *docs.Document, headerID string) SectionMetadata {
	header, ok := doc.Headers[headerID]
	if !ok {
		return SectionMetadata{}
	}

	lines := make([]string, 0)
	for _, el := range header.Content {
		if el.Paragraph == nil || el.Paragraph.Elements == nil {
			continue
		}
		line := paragraphToPlainText(el)
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}

	md := SectionMetadata{}
	if len(lines) > 0 {
		md.Title = lines[0]
	}
	for _, line := range lines {
		key, bpm, time, ok := parseMetadataLine(line)
		if ok {
			md.Key = key
			md.BPM = bpm
			md.Time = time
			break
		}
	}
	return md
}

func (s *DriveFileService) extractSectionMetadata(doc *docs.Document, section docs.StructuralElement) SectionMetadata {
	md := SectionMetadata{}
	sectionStart := section.StartIndex + 1
	sections := getSections(doc)
	sectionIndex := 0
	for i := range sections {
		if sections[i].StartIndex == section.StartIndex {
			sectionIndex = i
			break
		}
	}
	sectionEnd := contentEndForSection(doc, sections, sectionIndex)

	continuous := findFirstContinuousBreakInRange(doc, sectionStart, sectionEnd)
	if continuous != nil {
		paragraphs := paragraphElementsInRange(doc, sectionStart, continuous.StartIndex)
		nonEmptyBeforeMetadata := make([]string, 0)
		for _, para := range paragraphs {
			text := paragraphToPlainText(para)
			if strings.TrimSpace(text) == "" {
				continue
			}
			if key, bpm, time, ok := parseMetadataLine(text); ok {
				md.Key = key
				md.BPM = bpm
				md.Time = time
				break
			}
			nonEmptyBeforeMetadata = append(nonEmptyBeforeMetadata, text)
		}
		if len(nonEmptyBeforeMetadata) > 0 {
			docTitleNormalized := normalizeComparableText(doc.Title)
			for _, candidate := range nonEmptyBeforeMetadata {
				if normalizeComparableText(candidate) == docTitleNormalized {
					md.Title = candidate
					break
				}
			}
			if md.Title == "" {
				md.Title = nonEmptyBeforeMetadata[0]
			}
		}
	}

	if md.Title == "" || md.Key == "" || md.BPM == "" || md.Time == "" {
		headerID := ""
		if section.SectionBreak != nil && section.SectionBreak.SectionStyle != nil {
			headerID = section.SectionBreak.SectionStyle.DefaultHeaderId
		}
		if headerID == "" && doc.DocumentStyle != nil {
			headerID = doc.DocumentStyle.DefaultHeaderId
		}
		if headerID != "" {
			hdrMD := metadataFromHeader(doc, headerID)
			if md.Title == "" {
				md.Title = hdrMD.Title
			}
			if md.Key == "" {
				md.Key = hdrMD.Key
			}
			if md.BPM == "" {
				md.BPM = hdrMD.BPM
			}
			if md.Time == "" {
				md.Time = hdrMD.Time
			}
		}
	}

	if md.Title == "" {
		md.Title = normalizeTextValue(doc.Title)
	}
	return normalizeMetadata(md)
}

func singleColumnSectionStyle() (*docs.SectionStyle, string) {
	return &docs.SectionStyle{
		ColumnProperties: []*docs.SectionColumnProperties{{}},
	}, "columnProperties"
}

func copyDimension(d *docs.Dimension) *docs.Dimension {
	if d == nil {
		return nil
	}
	return &docs.Dimension{
		Magnitude: d.Magnitude,
		Unit:      d.Unit,
	}
}

func copyBodySectionStyle(style *docs.SectionStyle) (*docs.SectionStyle, string) {
	if style == nil {
		return nil, ""
	}

	copied := &docs.SectionStyle{}
	fields := make([]string, 0, 12)

	if len(style.ColumnProperties) > 0 {
		columns := make([]*docs.SectionColumnProperties, 0, len(style.ColumnProperties))
		for _, col := range style.ColumnProperties {
			if col == nil {
				columns = append(columns, &docs.SectionColumnProperties{})
				continue
			}
			columns = append(columns, &docs.SectionColumnProperties{PaddingEnd: copyDimension(col.PaddingEnd)})
		}
		copied.ColumnProperties = columns
		fields = append(fields, "columnProperties")
	}

	if strings.TrimSpace(style.ColumnSeparatorStyle) != "" {
		copied.ColumnSeparatorStyle = style.ColumnSeparatorStyle
		fields = append(fields, "columnSeparatorStyle")
	}

	if strings.TrimSpace(style.ContentDirection) != "" {
		copied.ContentDirection = style.ContentDirection
		fields = append(fields, "contentDirection")
	}
	if style.MarginBottom != nil {
		copied.MarginBottom = copyDimension(style.MarginBottom)
		fields = append(fields, "marginBottom")
	}
	if style.MarginFooter != nil {
		copied.MarginFooter = copyDimension(style.MarginFooter)
		fields = append(fields, "marginFooter")
	}
	if style.MarginHeader != nil {
		copied.MarginHeader = copyDimension(style.MarginHeader)
		fields = append(fields, "marginHeader")
	}
	if style.MarginLeft != nil {
		copied.MarginLeft = copyDimension(style.MarginLeft)
		fields = append(fields, "marginLeft")
	}
	if style.MarginRight != nil {
		copied.MarginRight = copyDimension(style.MarginRight)
		fields = append(fields, "marginRight")
	}
	if style.MarginTop != nil {
		copied.MarginTop = copyDimension(style.MarginTop)
		fields = append(fields, "marginTop")
	}
	if style.FlipPageOrientation {
		copied.FlipPageOrientation = true
		fields = append(fields, "flipPageOrientation")
	}
	if style.PageNumberStart > 0 {
		copied.PageNumberStart = style.PageNumberStart
		fields = append(fields, "pageNumberStart")
	}
	if style.UseFirstPageHeaderFooter {
		copied.UseFirstPageHeaderFooter = true
		fields = append(fields, "useFirstPageHeaderFooter")
	}

	if len(fields) == 0 {
		return nil, ""
	}
	return copied, strings.Join(fields, ",")
}

func sectionStyleUpdateRequest(start int64, style *docs.SectionStyle, fields string) *docs.Request {
	if style == nil || fields == "" {
		return nil
	}
	return &docs.Request{
		UpdateSectionStyle: &docs.UpdateSectionStyleRequest{
			Range: &docs.Range{
				StartIndex:      start,
				EndIndex:        start + 1,
				ForceSendFields: []string{"StartIndex"},
			},
			SectionStyle: style,
			Fields:       fields,
		},
	}
}

func applyMetadataPatch(md SectionMetadata, patch MetadataPatch) SectionMetadata {
	if patch.Title != nil {
		md.Title = *patch.Title
	}
	if patch.Key != nil {
		md.Key = *patch.Key
	}
	if patch.BPM != nil {
		md.BPM = *patch.BPM
	}
	if patch.Time != nil {
		md.Time = *patch.Time
	}
	return normalizeMetadata(md)
}

func newMetadataParagraphStyle(alignment string) *docs.ParagraphStyle {
	return &docs.ParagraphStyle{
		Alignment: alignment,
		SpaceAbove: &docs.Dimension{
			Magnitude: paraSpacingMagnitude,
			Unit:      unitPoints,
		},
		SpaceBelow: &docs.Dimension{
			Magnitude: paraSpacingMagnitude,
			Unit:      unitPoints,
		},
		LineSpacing: paraLineSpacing,
	}
}

func newMetadataTextStyle(fontSize float64) *docs.TextStyle {
	return &docs.TextStyle{
		WeightedFontFamily: &docs.WeightedFontFamily{FontFamily: fontFamilyRobotoMono},
		FontSize: &docs.Dimension{
			Magnitude: fontSize,
			Unit:      unitPoints,
		},
		Bold:            true,
		Italic:          false,
		Underline:       false,
		Strikethrough:   false,
		ForegroundColor: newOptionalColor(rgbColorBlack),
	}
}

func newMetadataKeyAccentStyle(chordColor *docs.RgbColor) *docs.TextStyle {
	if chordColor == nil {
		chordColor = rgbColorChord
	}
	return &docs.TextStyle{
		ForegroundColor: newOptionalColor(chordColor),
		Bold:            true,
	}
}

func composeCanonicalMetadataStyleRequests(sectionStart int64, md SectionMetadata, chordColor *docs.RgbColor) []*docs.Request {
	md = normalizeMetadata(md)

	titleText := md.Title + "\n"
	metaLineText := fmt.Sprintf("KEY: %s; BPM: %s; TIME: %s;\n", md.Key, md.BPM, md.Time)
	lastLineText := "\n"

	titleStart := sectionStart
	titleEnd := titleStart + int64(len([]rune(titleText)))
	metaStart := titleEnd
	metaEnd := metaStart + int64(len([]rune(metaLineText)))
	lastStart := metaEnd
	lastEnd := lastStart + int64(len([]rune(lastLineText)))

	requests := make([]*docs.Request, 0, 7)
	requests = append(requests,
		newUpdateParagraphStyleRequest(
			newMetadataParagraphStyle(metadataAlignmentCenter),
			"alignment,lineSpacing,spaceAbove,spaceBelow",
			titleStart,
			titleEnd,
			"",
		),
		newUpdateParagraphStyleRequest(
			newMetadataParagraphStyle(metadataAlignmentEnd),
			"alignment,lineSpacing,spaceAbove,spaceBelow",
			metaStart,
			metaEnd,
			"",
		),
		newUpdateParagraphStyleRequest(
			newMetadataParagraphStyle(metadataAlignmentCenter),
			"alignment,lineSpacing,spaceAbove,spaceBelow",
			lastStart,
			lastEnd,
			"",
		),
		newUpdateTextStyleRequest(
			newMetadataTextStyle(metadataFontSizeTitle),
			"weightedFontFamily,fontSize,bold,italic,underline,strikethrough,foregroundColor",
			titleStart,
			titleEnd,
			"",
		),
		newUpdateTextStyleRequest(
			newMetadataTextStyle(metadataFontSizeLine),
			"weightedFontFamily,fontSize,bold,italic,underline,strikethrough,foregroundColor",
			metaStart,
			metaEnd,
			"",
		),
		newUpdateTextStyleRequest(
			newMetadataTextStyle(metadataFontSizeLastLine),
			"weightedFontFamily,fontSize,bold,italic,underline,strikethrough,foregroundColor",
			lastStart,
			lastEnd,
			"",
		),
	)
	keyStart := metaStart + int64(len([]rune("KEY: ")))
	keyEnd := keyStart + int64(len([]rune(string(md.Key))))
	if keyEnd > keyStart {
		requests = append(requests, newUpdateTextStyleRequest(
			newMetadataKeyAccentStyle(chordColor),
			"foregroundColor,bold",
			keyStart,
			keyEnd,
			"",
		))
	}

	return requests
}

func isCanonicalMetadataSubsection(paragraphs []*docs.StructuralElement, md SectionMetadata, titleIdx, metadataIdx int) bool {
	if len(paragraphs) != 3 {
		return false
	}
	if titleIdx != 0 || metadataIdx != 1 {
		return false
	}

	titleText := paragraphToExactLineText(paragraphs[0])
	if titleText != md.Title {
		return false
	}

	metaLineText := paragraphToExactLineText(paragraphs[1])
	expectedMetaLine := fmt.Sprintf("KEY: %s; BPM: %s; TIME: %s;", md.Key, md.BPM, md.Time)
	if metaLineText != expectedMetaLine {
		return false
	}

	if paragraphToExactLineText(paragraphs[2]) != "" {
		return false
	}

	return true
}

func leadingEmptyBodyParagraphRange(doc *docs.Document, bodyStart, bodyEnd int64) (int64, int64, bool) {
	hasLeadingEmpty := false
	var deleteStart, deleteEnd int64

	for _, item := range doc.Body.Content {
		if item.StartIndex < bodyStart || item.EndIndex > bodyEnd {
			continue
		}
		if item.SectionBreak != nil {
			continue
		}
		if item.Paragraph == nil {
			if hasLeadingEmpty {
				return deleteStart, deleteEnd, true
			}
			return 0, 0, false
		}

		if strings.TrimSpace(paragraphToPlainText(item)) == "" {
			if !hasLeadingEmpty {
				hasLeadingEmpty = true
				deleteStart = item.StartIndex
			}
			deleteEnd = item.EndIndex
			continue
		}

		if hasLeadingEmpty {
			return deleteStart, deleteEnd, true
		}
		return 0, 0, false
	}

	return 0, 0, false
}

func findMetadataTitleAndLine(paragraphs []*docs.StructuralElement, docTitle string) (int, int) {
	titleIdx := -1
	metadataIdx := -1
	type titleCandidate struct {
		index int
		text  string
	}
	nonEmptyBeforeMetadata := make([]titleCandidate, 0)
	docTitleNormalized := normalizeComparableText(docTitle)

	for i, para := range paragraphs {
		text := paragraphToPlainText(para)
		if strings.TrimSpace(text) == "" {
			continue
		}
		if _, _, _, ok := parseMetadataLine(text); ok {
			metadataIdx = i
			break
		}
		nonEmptyBeforeMetadata = append(nonEmptyBeforeMetadata, titleCandidate{
			index: i,
			text:  text,
		})
	}
	if len(nonEmptyBeforeMetadata) > 0 {
		for _, candidate := range nonEmptyBeforeMetadata {
			if normalizeComparableText(candidate.text) == docTitleNormalized {
				titleIdx = candidate.index
				break
			}
		}
		if titleIdx < 0 {
			titleIdx = nonEmptyBeforeMetadata[0].index
		}
	}
	return titleIdx, metadataIdx
}

func (s *DriveFileService) normalizeMetadataLayoutWithOptions(ID string, options normalizeMetadataLayoutOptions) (*docs.Document, *MetadataNormalizeResult, error) {
	doc, err := s.getDoc(ID)
	if err != nil {
		return nil, nil, err
	}

	sections := getSections(doc)
	result := &MetadataNormalizeResult{}
	requests := make([]*docs.Request, 0)

	for i := len(sections) - 1; i >= 0; i-- {
		section := sections[i]
		sectionStart := section.StartIndex + 1
		sectionEnd := contentEndForSection(doc, sections, i)

		continuous := findFirstContinuousBreakInRange(doc, sectionStart, sectionEnd)
		metadataEnd := sectionStart
		if continuous != nil {
			metadataEnd = continuous.StartIndex
		}

		metadataParagraphs := paragraphElementsInRange(doc, sectionStart, metadataEnd)
		titleIdx, metadataIdx := findMetadataTitleAndLine(metadataParagraphs, doc.Title)

		md := s.extractSectionMetadata(doc, section)
		if titleIdx >= 0 {
			titleText := paragraphToPlainText(metadataParagraphs[titleIdx])
			if strings.TrimSpace(titleText) != "" {
				md.Title = titleText
			}
		}
		if metadataIdx >= 0 {
			key, bpm, time, ok := parseMetadataLine(paragraphToPlainText(metadataParagraphs[metadataIdx]))
			if ok {
				md.Key = key
				md.BPM = bpm
				md.Time = time
			}
		}
		md = normalizeMetadata(md)
		// Title is always synced with the document title.
		md.Title = normalizeTextValue(doc.Title)

		canonicalText := composeCanonicalMetadataText(md)
		metadataLen := int64(len([]rune(canonicalText)))
		canonicalAlready := continuous != nil && isCanonicalMetadataSubsection(metadataParagraphs, md, titleIdx, metadataIdx)
		needsMetadataRestyle := !canonicalAlready

		bodyStart := sectionStart + metadataLen + 1
		tailParagraphs := make([]*docs.StructuralElement, 0)
		if !canonicalAlready {
			if metadataIdx >= 0 && metadataIdx+1 < len(metadataParagraphs) {
				tailCandidates := metadataParagraphs[metadataIdx+1:]
				// Preserve the canonical empty paragraph after KEY/BPM/TIME inside metadata.
				if len(tailCandidates) > 0 && strings.TrimSpace(paragraphToPlainText(tailCandidates[0])) == "" {
					tailCandidates = tailCandidates[1:]
				}
				tailParagraphs = keepTailParagraphsForBody(tailCandidates)
			}
			// If body starts with empty paragraphs and we're not moving non-empty tail content,
			// remove those leading empties so section body starts with real content.
			if continuous != nil && len(tailParagraphs) == 0 {
				bodyStartExisting := continuous.StartIndex + 1
				if deleteStart, deleteEnd, ok := leadingEmptyBodyParagraphRange(doc, bodyStartExisting, sectionEnd); ok {
					requests = append(requests, newDeleteContentRangeRequest(deleteStart, deleteEnd, ""))
				}
			}

			deleteEnd := metadataEnd
			if continuous != nil && deleteEnd > sectionStart {
				deleteEnd--
			}
			if deleteEnd > sectionStart {
				requests = append(requests, newDeleteContentRangeRequest(sectionStart, deleteEnd, ""))
			}

			insertText := canonicalText
			insertedMetadataLen := metadataLen
			if continuous == nil || continuous.StartIndex > sectionStart {
				insertText = composeMetadataTextWithoutTrailingEmpty(md)
				insertedMetadataLen = int64(len([]rune(insertText)))
			}
			requests = append(requests, newInsertTextRequest(insertText, sectionStart, ""))

			if continuous == nil {
				breakIndex := sectionStart + insertedMetadataLen
				if sectionEnd <= sectionStart {
					breakIndex = sectionStart + insertedMetadataLen - 1
				}
				if breakIndex < sectionStart {
					breakIndex = sectionStart
				}
				requests = append(requests, &docs.Request{
					InsertSectionBreak: &docs.InsertSectionBreakRequest{
						Location: &docs.Location{
							Index:           breakIndex,
							ForceSendFields: []string{"Index"},
						},
						SectionType: "CONTINUOUS",
					},
				})

				if section.SectionBreak != nil {
					if bodyStyle, fields := copyBodySectionStyle(section.SectionBreak.SectionStyle); bodyStyle != nil && fields != "" {
						if req := sectionStyleUpdateRequest(bodyStart, bodyStyle, fields); req != nil {
							requests = append(requests, req)
						}
					}
				}
			}
		}
		if canonicalAlready && continuous != nil {
			bodyStartExisting := continuous.StartIndex + 1
			if deleteStart, deleteEnd, ok := leadingEmptyBodyParagraphRange(doc, bodyStartExisting, sectionEnd); ok {
				requests = append(requests, newDeleteContentRangeRequest(deleteStart, deleteEnd, ""))
			}
		}

		if needsMetadataRestyle && options.applyMetadataStyles {
			requests = append(requests, composeCanonicalMetadataStyleRequests(sectionStart, md, chordColorForSectionIndex(i))...)
		}

		if metaStyle, fields := singleColumnSectionStyle(); metaStyle != nil {
			if req := sectionStyleUpdateRequest(sectionStart, metaStyle, fields); req != nil {
				requests = append(requests, req)
			}
		}

		if len(tailParagraphs) > 0 {
			requests = append(requests, cloneParagraphElementsRequests(tailParagraphs, bodyStart, "")...)
		}

		result.SectionsNormalized++
	}

	headerIDs := make([]string, 0)
	for headerID := range doc.Headers {
		headerIDs = append(headerIDs, headerID)
	}
	sort.Strings(headerIDs)
	for _, headerID := range headerIDs {
		requests = append(requests, &docs.Request{
			DeleteHeader: &docs.DeleteHeaderRequest{HeaderId: headerID},
		})
		result.HeadersDeleted++
	}

	if len(requests) > 0 {
		_, err = s.batchUpdate(ID, requests)
		if err != nil {
			return nil, nil, err
		}
	}

	doc, err = s.getDoc(ID)
	if err != nil {
		return nil, nil, err
	}

	return doc, result, nil
}

func (s *DriveFileService) normalizeMetadataLayout(ID string) (*docs.Document, *MetadataNormalizeResult, error) {
	return s.normalizeMetadataLayoutWithOptions(ID, normalizeMetadataLayoutOptions{
		applyMetadataStyles: true,
	})
}

func (s *DriveFileService) ensureBodyMetadataLayout(ID string) (*docs.Document, error) {
	doc, _, err := s.normalizeMetadataLayout(ID)
	return doc, err
}

func metadataRewriteRequestsForSection(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, md SectionMetadata) ([]*docs.Request, error) {
	if sectionIndex < 0 || sectionIndex >= len(sections) {
		return nil, fmt.Errorf("section index %d is out of bounds", sectionIndex)
	}

	section := sections[sectionIndex]
	sectionStart := section.StartIndex + 1
	sectionEnd := contentEndForSection(doc, sections, sectionIndex)
	continuous := findFirstContinuousBreakInRange(doc, sectionStart, sectionEnd)
	if continuous == nil {
		return nil, fmt.Errorf("section %d has no continuous break after metadata", sectionIndex)
	}

	md = normalizeMetadata(md)
	insertText := composeMetadataTextWithoutTrailingEmpty(md)
	deleteEnd := continuous.StartIndex
	if deleteEnd > sectionStart {
		deleteEnd--
	}

	requests := make([]*docs.Request, 0, 4)
	if deleteEnd > sectionStart {
		requests = append(requests, newDeleteContentRangeRequest(sectionStart, deleteEnd, ""))
	}
	requests = append(requests, newInsertTextRequest(insertText, sectionStart, ""))
	requests = append(requests, composeCanonicalMetadataStyleRequests(sectionStart, md, chordColorForSectionIndex(sectionIndex))...)

	if metaStyle, fields := singleColumnSectionStyle(); metaStyle != nil {
		if req := sectionStyleUpdateRequest(sectionStart, metaStyle, fields); req != nil {
			requests = append(requests, req)
		}
	}

	return requests, nil
}

func (s *DriveFileService) updateSectionMetadataByIndex(ID string, sectionIndex int, patch MetadataPatch) error {
	doc, _, err := s.normalizeMetadataLayout(ID)
	if err != nil {
		return err
	}

	sections := getSections(doc)
	if sectionIndex < 0 || sectionIndex >= len(sections) {
		return fmt.Errorf("section index %d is out of bounds", sectionIndex)
	}
	md := s.extractSectionMetadata(doc, sections[sectionIndex])
	md = applyMetadataPatch(md, patch)
	md.Title = normalizeTextValue(doc.Title)
	requests, err := metadataRewriteRequestsForSection(doc, sections, sectionIndex, md)
	if err != nil {
		return err
	}

	_, err = s.batchUpdate(ID, requests)
	return err
}

func (s *DriveFileService) updateMetadataAcrossSections(ID string, patch MetadataPatch) error {
	doc, _, err := s.normalizeMetadataLayout(ID)
	if err != nil {
		return err
	}

	sections := getSections(doc)
	requests := make([]*docs.Request, 0)

	for i := len(sections) - 1; i >= 0; i-- {
		section := sections[i]
		sectionStart := section.StartIndex + 1
		sectionEnd := contentEndForSection(doc, sections, i)
		continuous := findFirstContinuousBreakInRange(doc, sectionStart, sectionEnd)
		if continuous == nil {
			continue
		}

		md := s.extractSectionMetadata(doc, section)
		md = applyMetadataPatch(md, patch)
		md.Title = normalizeTextValue(doc.Title)
		sectionReqs, reqErr := metadataRewriteRequestsForSection(doc, sections, i, md)
		if reqErr != nil {
			return reqErr
		}
		requests = append(requests, sectionReqs...)
	}

	_, err = s.batchUpdate(ID, requests)
	return err
}

func (s *DriveFileService) NormalizeMetadataLayout(ID string) error {
	_, _, err := s.normalizeMetadataLayout(ID)
	return err
}

func (s *DriveFileService) EnsureBodyMetadataLayout(ID string) (*docs.Document, error) {
	return s.ensureBodyMetadataLayout(ID)
}

func (s *DriveFileService) UpdateMetadataAcrossSections(ID string, patch MetadataPatch) error {
	return s.updateMetadataAcrossSections(ID, patch)
}
