package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/joeyave/chords-transposer/transposer"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

var (
	newLineRe    = regexp.MustCompile(`\s*[\r\n]$`)
	barlineRe    = regexp.MustCompile(`[|]`)
	bracketedRe  = regexp.MustCompile(`\[[^]]*]`)
	repetitionRe = regexp.MustCompile(`([xх])\d+`)
)

const (
	headerTypeDefault    = "DEFAULT"
	unitPoints           = "PT"
	keyNashville         = "NNS"
	fontFamilyRobotoMono = "Roboto Mono"
	alignmentCenter      = "CENTER"
	alignmentEnd         = "END"

	// Margins and Spacing
	docMarginVertical   float64 = 14
	docMarginHorizontal float64 = 30
	docMarginHeader     float64 = 18

	paraLineSpacing      float64 = 90
	paraSpacingMagnitude float64 = 0

	// Font Sizes
	fontSizeHeaderTitle    float64 = 20
	fontSizeHeaderMetadata float64 = 14
	fontSizeHeaderLastPara float64 = 11
)

var rgbColorChord = newRgbColor(0.8, 0, 0)
var rgbColorBlack = newRgbColor(0, 0, 0)

// getDoc is a helper to fetch a document by its ID.
func (s *DriveFileService) getDoc(ID string) (*docs.Document, error) {
	return s.docsRepository.Documents.Get(ID).Do()
}

// batchUpdate is a helper to send a batch of requests for a document.
func (s *DriveFileService) batchUpdate(docID string, requests []*docs.Request) (*docs.BatchUpdateDocumentResponse, error) {
	if len(requests) == 0 {
		return nil, nil // No requests to send
	}
	return s.docsRepository.Documents.BatchUpdate(docID,
		&docs.BatchUpdateDocumentRequest{Requests: requests}).Do()
}

func (s *DriveFileService) TransposeOne(ID string, toKey string, sectionIndex int) (*drive.File, error) {
	doc, err := s.getDoc(ID)
	if err != nil {
		return nil, err
	}

	sections := getSections(doc)

	if len(sections) <= sectionIndex || sectionIndex < 0 {
		sections, err = s.appendSectionByID(ID)
		if err != nil {
			return nil, err
		}

		doc, err = s.getDoc(ID)
		if err != nil {
			return nil, err
		}

		sectionIndex = len(sections) - 1
	}

	requests, key := transposeHeader(doc, sections, sectionIndex, toKey)
	requests = append(requests, transposeBody(doc, sections, sectionIndex, key, toKey)...)

	_, err = s.batchUpdate(doc.DocumentId, requests)
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

func (s *DriveFileService) TransposeHeader(ID string, toKey string, sectionIndex int) (*drive.File, error) {
	doc, err := s.getDoc(ID)
	if err != nil {
		return nil, err
	}

	sections := getSections(doc)

	if len(sections) <= sectionIndex || sectionIndex < 0 {
		return nil, fmt.Errorf("section index %d is out of bounds", sectionIndex)
	}

	requests, _ := transposeHeader(doc, sections, sectionIndex, toKey)

	_, err = s.batchUpdate(doc.DocumentId, requests)
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

func (s *DriveFileService) StyleOne(ID, lang string) (*drive.File, error) {
	requests := make([]*docs.Request, 0)

	doc, err := s.getDoc(ID)
	if err != nil {
		return nil, err
	}

	if doc.DocumentStyle.DefaultHeaderId == "" {
		res, err := s.batchUpdate(ID, []*docs.Request{
			newCreateHeaderRequest(headerTypeDefault, nil),
		})

		if err == nil && res != nil && len(res.Replies) > 0 {
			createHeaderRes := res.Replies[0].CreateHeader
			if createHeaderRes != nil && createHeaderRes.HeaderId != "" {
				_, _ = s.batchUpdate(ID, []*docs.Request{
					getDefaultHeaderRequest(createHeaderRes.HeaderId, doc.Title, "", "", "", lang),
				})
			}
		}
	}

	doc, err = s.getDoc(ID)
	if err != nil {
		return nil, err
	}

	for _, header := range doc.Headers {
		requests = append(requests, composeStyleRequests(header.Content, header.HeaderId, true, 0)...)
	}

	requests = append(requests, composeStyleRequests(doc.Body.Content, "", false, 0)...)

	docStyle := &docs.DocumentStyle{
		MarginBottom: &docs.Dimension{
			Magnitude: docMarginVertical,
			Unit:      unitPoints,
		},
		MarginLeft: &docs.Dimension{
			Magnitude: docMarginHorizontal,
			Unit:      unitPoints,
		},
		MarginRight: &docs.Dimension{
			Magnitude: docMarginHorizontal,
			Unit:      unitPoints,
		},
		MarginTop: &docs.Dimension{
			Magnitude: docMarginVertical,
			Unit:      unitPoints,
		},
		MarginHeader: &docs.Dimension{
			Magnitude: docMarginHeader,
			Unit:      unitPoints,
		},
		UseFirstPageHeaderFooter: false,
		ForceSendFields:          []string{"UseFirstPageHeaderFooter"},
	}
	requests = append(requests, newUpdateDocumentStyleRequest(docStyle, "marginBottom,marginLeft,marginRight,marginTop,marginHeader,useFirstPageHeaderFooter"))

	_, err = s.batchUpdate(ID, requests)
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

func transposeHeader(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, toKey string) ([]*docs.Request, string) {
	if doc.DocumentStyle.DefaultHeaderId == "" {
		return nil, ""
	}

	requests := make([]*docs.Request, 0)

	section := sections[sectionIndex]
	defaultHeaderID := section.SectionBreak.SectionStyle.DefaultHeaderId

	var targetSegmentID string // Will hold the ID of the header to write to

	// Create header if section doesn't have it.
	if defaultHeaderID == "" {
		location := newLocation(section.StartIndex, "")
		requests = append(requests, newCreateHeaderRequest(headerTypeDefault, location))
		// NOTE: targetSegmentID will be "", which may cause transpose to write to the body.
	} else {
		header := doc.Headers[defaultHeaderID]
		targetSegmentID = header.HeaderId // Set the segment ID for transpose

		// Clear existing content from the section header
		lastHeaderContent := header.Content[len(header.Content)-1]
		if lastHeaderContent.EndIndex-1 > 0 {
			requests = append(requests, newDeleteContentRangeRequest(0, lastHeaderContent.EndIndex-1, header.HeaderId))
		}
	}
	transposeRequests, key := composeTransposeRequests(
		doc.Headers[doc.DocumentStyle.DefaultHeaderId].Content,
		0,
		"",
		toKey,
		targetSegmentID, // Use the extracted variable
		0,
	)
	requests = append(requests, transposeRequests...)

	return requests, key
}

func transposeBody(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, key string, toKey string) []*docs.Request {
	requests := make([]*docs.Request, 0)

	sectionToInsertStartIndex := sections[sectionIndex].StartIndex + 1
	var sectionToInsertEndIndex int64

	if len(sections) > sectionIndex+1 {
		sectionToInsertEndIndex = sections[sectionIndex+1].StartIndex - 1
	} else {
		lastContentElement := doc.Body.Content[len(doc.Body.Content)-1]
		sectionToInsertEndIndex = lastContentElement.EndIndex - 1
	}

	content := getContentForFirstSection(doc, sections)

	if sectionToInsertEndIndex-sectionToInsertStartIndex > 0 {
		requests = append(requests, newDeleteContentRangeRequest(sectionToInsertStartIndex, sectionToInsertEndIndex, ""))
	}

	transposeRequests, _ := composeTransposeRequests(
		content,
		sectionToInsertStartIndex,
		key,
		toKey,
		"",
		0,
	)
	requests = append(requests, transposeRequests...)

	return requests
}

func composeTransposeRequests(content []*docs.StructuralElement, index int64, key string, toKey string, segmentId string, chordRatioThreshold float64) ([]*docs.Request, string) {
	allRequests := make([]*docs.Request, 0)
	paragraphs, idxs := getParagraphs(content)

	for i, paragraph := range paragraphs {
		fullText := idxs[i].full

		// Decide if this paragraph should be treated as chords
		shouldTranspose := shouldTransposeParagraph(fullText, chordRatioThreshold)

		// Determine a paragraph-level key once (fall back to curKey)
		key = guessKeyIfNeeded(key, fullText)

		// Generate requests for all elements in this paragraph
		isLastParagraph := (i == len(paragraphs)-1)
		paraRequests, newIndex := newTransposeRequestsForParagraph(
			paragraph, isLastParagraph, shouldTranspose,
			key, toKey, segmentId, index,
		)

		allRequests = append(allRequests, paraRequests...)
		index = newIndex // Update the index for the next paragraph
	}

	return allRequests, key
}

func composeStyleRequests(content []*docs.StructuralElement, segmentID string, isHeader bool, chordRatioThreshold float64) []*docs.Request {
	requests := make([]*docs.Request, 0)

	for i, paragraph := range content {
		if paragraph.Paragraph == nil {
			continue
		}

		// 1) Paragraph-level spacing
		paragraphStyle := docs.ParagraphStyle{
			SpaceAbove:  &docs.Dimension{Magnitude: paraSpacingMagnitude, Unit: unitPoints, ForceSendFields: []string{"Magnitude"}},
			SpaceBelow:  &docs.Dimension{Magnitude: paraSpacingMagnitude, Unit: unitPoints, ForceSendFields: []string{"Magnitude"}},
			LineSpacing: paraLineSpacing,
		}
		if isHeader {
			paragraphStyle.Alignment = alignmentCenter
			if i == 1 {
				paragraphStyle.Alignment = alignmentEnd
			}
		}

		paragraphStyleFields := "lineSpacing,spaceAbove,spaceBelow"
		if isHeader {
			paragraphStyleFields = "alignment," + paragraphStyleFields
		}
		requests = append(requests, newUpdateParagraphStyleRequest(&paragraphStyle, paragraphStyleFields, paragraph.StartIndex, paragraph.EndIndex, segmentID))

		// 2) Ensure all runs use Roboto Mono
		requests = append(requests, newBaseTextStyleRequests(paragraph.Paragraph, isHeader, i, segmentID)...)

		// 3) Style chords across the whole paragraph (paragraph-level heuristic)
		requests = append(requests, changeStyleForChordsAcross(paragraph.Paragraph, segmentID, chordRatioThreshold)...)

		// [|] -> bold, black
		textStyle := docs.TextStyle{
			Bold:            true,
			ForegroundColor: newOptionalColor(rgbColorBlack),
		}
		requests = append(requests, changeStyleByRegexAcross(barlineRe, paragraph.Paragraph, textStyle, "bold,foregroundColor", nil, segmentID)...)

		// [ ... ] -> bold
		textStyle = docs.TextStyle{
			Bold: true,
		}
		requests = append(requests, changeStyleByRegexAcross(bracketedRe, paragraph.Paragraph, textStyle, "bold", nil, segmentID)...)

		// (x|х)\d+ -> bold, red-ish
		textStyle = docs.TextStyle{
			Bold:            true,
			ForegroundColor: newOptionalColor(rgbColorChord),
		}
		requests = append(requests, changeStyleByRegexAcross(repetitionRe, paragraph.Paragraph, textStyle, "bold,foregroundColor", nil, segmentID)...)
	}

	return requests
}

type paraSlice struct {
	el *docs.ParagraphElement
	// paragraph-relative rune offsets for this element's content
	start int64
	end   int64
}

type paraIndex struct {
	full   string
	slices []paraSlice
}

// buildParaIndex concatenates all TextRun contents and builds rune-based element slices.
func buildParaIndex(paragraph *docs.Paragraph) (paraIndex, bool) {
	if paragraph == nil || paragraph.Elements == nil {
		return paraIndex{}, false
	}
	var builder strings.Builder
	slices := make([]paraSlice, 0, len(paragraph.Elements))
	var runeOffset int64
	for i := range paragraph.Elements {
		el := paragraph.Elements[i]
		if el.TextRun == nil || el.TextRun.Content == "" {
			continue
		}
		start := runeOffset
		runes := []rune(el.TextRun.Content)
		runeOffset += int64(len(runes))
		end := runeOffset
		builder.WriteString(el.TextRun.Content)
		slices = append(slices, paraSlice{el: el, start: start, end: end})
	}
	full := builder.String()
	if full == "" {
		return paraIndex{}, false
	}
	return paraIndex{full: full, slices: slices}, true
}

// toDocRange converts paragraph-relative rune offsets to absolute doc range.
func toDocRange(paraIndex paraIndex, runeStart, runeEnd int64) (int64, int64, bool) {
	var first *docs.ParagraphElement
	var last *docs.ParagraphElement
	var firstElementOffset, lastElementOffset int64
	for _, slice := range paraIndex.slices {
		if runeStart < slice.end && runeEnd > slice.start {
			if first == nil {
				first = slice.el
				firstElementOffset = runeStart - slice.start
			}
			last = slice.el
			lastElementOffset = runeEnd - slice.start
		}
	}
	if first == nil || last == nil {
		return 0, 0, false
	}
	return first.StartIndex + firstElementOffset, last.StartIndex + lastElementOffset, true
}

// styleRange builds a single UpdateTextStyle request for [docStart, docEnd).
func styleRange(docStart, docEnd int64, style docs.TextStyle, fields, segmentID string) *docs.Request {
	styleCopy := style // copy to avoid mutation surprises
	return newUpdateTextStyleRequest(&styleCopy, fields, docStart, docEnd, segmentID)
}

// replaceRange deletes [docStart, docEnd) and inserts text at docStart.
func replaceRange(docStart, docEnd int64, text, segmentID string) []*docs.Request {
	return []*docs.Request{
		newDeleteContentRangeRequest(docStart, docEnd, segmentID),
		newInsertTextRequest(text, docStart, segmentID),
	}
}

// byteToRune converts a byte index (from Go regex) into a rune count prefix.
func byteToRune(full string, byteIdx int) int64 {
	return int64(len([]rune(full[:byteIdx])))
}

// changeStyleByRegexAcross applies style (and optional textFunc transforms) for matches
// that may span multiple ParagraphElements inside the given paragraph.
func changeStyleByRegexAcross(regex *regexp.Regexp, paragraph *docs.Paragraph, style docs.TextStyle, fields string, textFunc func(string) string, segmentID string) []*docs.Request {

	requests := make([]*docs.Request, 0)
	if paragraph == nil || paragraph.Elements == nil {
		return requests
	}

	// Build paragraph index once
	idx, ok := buildParaIndex(paragraph)
	if !ok {
		return requests
	}

	// Find matches on concatenated text (regex gives byte offsets)
	matches := regex.FindAllStringIndex(idx.full, -1)
	if matches == nil {
		return requests
	}

	for _, match := range matches {
		runeStart := byteToRune(idx.full, match[0])
		runeEnd := byteToRune(idx.full, match[1])
		if runeStart == runeEnd {
			continue
		}

		docStart, docEnd, ok := toDocRange(idx, runeStart, runeEnd)
		if !ok {
			continue
		}

		// Optional replacement before styling
		if textFunc != nil {
			originalText := idx.full[match[0]:match[1]]
			replacementText := textFunc(originalText)
			requests = append(requests, replaceRange(docStart, docEnd, replacementText, segmentID)...)
			// Adjust end to new length
			docEnd = docStart + int64(len([]rune(replacementText)))
		}

		requests = append(requests, styleRange(docStart, docEnd, style, fields, segmentID))
	}

	return requests
}

// changeStyleForChordsAcross applies chord styling across an entire paragraph,
// using a paragraph-level heuristic to avoid false positives (e.g., verse numbers).
// If the ratio of chord tokens to total tokens is below chordRatioThreshold,
// no styling is applied for this paragraph.
func changeStyleForChordsAcross(paragraph *docs.Paragraph, segmentID string, chordRatioThreshold float64) []*docs.Request {
	requests := make([]*docs.Request, 0)
	if paragraph == nil || paragraph.Elements == nil {
		return requests
	}

	idx, ok := buildParaIndex(paragraph)
	if !ok {
		return requests
	}

	// Tokenize the full paragraph (heuristic is inside Tokenize via ChordRatioThreshold)
	lines := transposer.Tokenize(idx.full, true, false, &transposer.TransposeOpts{
		ChordRatioThreshold: chordRatioThreshold,
	})

	// Style for chords
	chordStyle := docs.TextStyle{
		Bold:            true,
		ForegroundColor: newOptionalColor(rgbColorChord),
	}

	for _, line := range lines {
		for _, token := range line {
			if token.Chord == nil {
				continue
			}
			runeStart := token.Offset
			runeEnd := token.Offset + int64(len([]rune(token.Chord.String())))
			if runeStart == runeEnd {
				continue
			}
			docStart, docEnd, ok := toDocRange(idx, runeStart, runeEnd)
			if !ok {
				continue
			}
			requests = append(requests, styleRange(docStart, docEnd, chordStyle, "bold,foregroundColor", segmentID))
		}
	}

	return requests
}

// --- Helper Functions ---

// getContentForFirstSection finds the body content that belongs to the first section.
func getContentForFirstSection(doc *docs.Document, sections []docs.StructuralElement) []*docs.StructuralElement {
	if len(sections) > 1 {
		index := len(doc.Body.Content)
		for i := range doc.Body.Content {
			if doc.Body.Content[i].StartIndex == sections[1].StartIndex {
				index = i
				break
			}
		}
		return doc.Body.Content[:index]
	}
	return doc.Body.Content
}

// getParagraphs filters content for valid paragraphs and builds a paraIndex for each.
func getParagraphs(content []*docs.StructuralElement) ([]*docs.Paragraph, []paraIndex) {
	var paragraphs []*docs.Paragraph
	var idxs []paraIndex
	for _, item := range content {
		if item.Paragraph == nil || item.Paragraph.Elements == nil {
			continue
		}
		idx, ok := buildParaIndex(item.Paragraph)
		if !ok {
			continue
		}
		paragraphs = append(paragraphs, item.Paragraph)
		idxs = append(idxs, idx)
	}
	return paragraphs, idxs
}

// newBaseTextStyleRequests applies the base font (Roboto Mono), weight, and header-specific font sizes.
func newBaseTextStyleRequests(paragraph *docs.Paragraph, isHeader bool, paragraphIndex int, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0, len(paragraph.Elements))

	for _, element := range paragraph.Elements {
		if element.TextRun == nil || element.TextRun.Content == "" {
			continue
		}

		textStyle := &docs.TextStyle{
			WeightedFontFamily: &docs.WeightedFontFamily{FontFamily: fontFamilyRobotoMono},
			Bold:               false,
		}
		textStyleFields := "weightedFontFamily,bold"

		existingTextStyle := element.TextRun.TextStyle
		if existingTextStyle != nil {
			if existingTextStyle.WeightedFontFamily != nil {
				textStyle.WeightedFontFamily.Weight = existingTextStyle.WeightedFontFamily.Weight
			}
			// Always include Bold in update: headers => true, body => preserve existing
			textStyle.Bold = isHeader || existingTextStyle.Bold
		}

		if isHeader {
			switch paragraphIndex {
			case 0:
				textStyle.FontSize = &docs.Dimension{Magnitude: fontSizeHeaderTitle, Unit: unitPoints}
			case 1:
				textStyle.FontSize = &docs.Dimension{Magnitude: fontSizeHeaderMetadata, Unit: unitPoints}
			case 2:
				textStyle.FontSize = &docs.Dimension{Magnitude: fontSizeHeaderLastPara, Unit: unitPoints}
			}
			textStyleFields += ",fontSize"
		}

		requests = append(requests, newUpdateTextStyleRequest(textStyle, textStyleFields, element.StartIndex, element.EndIndex, segmentID))
	}
	return requests
}

// shouldTransposeParagraph decides if a paragraph should be transposed based on its chord ratio.
func shouldTransposeParagraph(fullText string, threshold float64) bool {
	if threshold <= 0 {
		return true // No heuristic, default to transposing
	}

	lines := transposer.Tokenize(fullText, true, false, &transposer.TransposeOpts{
		ChordRatioThreshold: threshold,
	})
	for _, line := range lines {
		for _, token := range line {
			if token.Chord != nil {
				return true // Heuristic passed, has chords
			}
		}
	}

	// Heuristic was run, but no chords were found
	return false
}

// guessKeyIfNeeded attempts to guess the key from text if no key is currently set.
func guessKeyIfNeeded(currentKey, fullText string) string {
	if currentKey != "" {
		return currentKey // Key is already set, do nothing
	}
	if guessedKey, err := transposer.GuessKeyFromText(fullText); err == nil {
		return guessedKey.String() // Guessed a new key
	}
	return currentKey // Guessing failed, return original (empty) key
}

// newTransposeRequestsForParagraph generates all the requests for a single paragraph's elements.
func newTransposeRequestsForParagraph(paragraph *docs.Paragraph, isLastParagraph bool, shouldTranspose bool, key, toKey, segmentId string, index int64) ([]*docs.Request, int64) {
	requests := make([]*docs.Request, 0)

	for j, element := range paragraph.Elements {
		if element.TextRun == nil || element.TextRun.Content == "" {
			continue
		}

		runText := element.TextRun.Content
		textStyle := element.TextRun.TextStyle // Extracted variable

		// Clean newline from the very last element of the content
		isLastElement := j == len(paragraph.Elements)-1

		if isLastParagraph && isLastElement {
			runText = newLineRe.ReplaceAllString(runText, " ")
		}

		if shouldTranspose && key != "" {
			var transposedText string
			var err error
			if toKey == keyNashville {
				transposedText, err = transposer.TransposeToNashville(runText, key)
			} else {
				transposedText, err = transposer.TransposeToKey(runText, key, toKey)
			}
			if err == nil {
				runText = transposedText
			}
		}

		if textStyle.ForegroundColor == nil {
			textStyle.ForegroundColor = newOptionalColor(rgbColorBlack)
		}

		runTextLen := int64(len([]rune(runText)))
		endIndex := index + runTextLen

		// Insert the (possibly transposed) text and reapply element + paragraph styles for that span
		requests = append(requests,
			newInsertTextRequest(runText, index, segmentId),
			newUpdateTextStyleRequest(textStyle, "*", index, endIndex, segmentId), // Use extracted variable
			newUpdateParagraphStyleRequest(paragraph.ParagraphStyle, "alignment, lineSpacing, direction, spaceAbove, spaceBelow", index, endIndex, segmentId),
		)

		index += runTextLen
	}

	return requests, index
}

// newRgbColor creates a new RgbColor struct with specified force fields.
func newRgbColor(r, g, b float64) *docs.RgbColor {
	return &docs.RgbColor{Red: r, Green: g, Blue: b, ForceSendFields: []string{"blue", "green", "red"}}
}

// newOptionalColor wraps an RgbColor in an OptionalColor struct.
func newOptionalColor(rgb *docs.RgbColor) *docs.OptionalColor {
	return &docs.OptionalColor{Color: &docs.Color{RgbColor: rgb}}
}

// newLocation creates a new Location struct.
func newLocation(index int64, segmentId string) *docs.Location {
	return &docs.Location{
		Index:           index,
		SegmentId:       segmentId,
		ForceSendFields: []string{"Index"},
	}
}

// newRange creates a new Range struct.
func newRange(start, end int64, segmentId string) *docs.Range {
	return &docs.Range{
		StartIndex:      start,
		EndIndex:        end,
		SegmentId:       segmentId,
		ForceSendFields: []string{"StartIndex"},
	}
}

// newInsertTextRequest creates a new InsertText request.
func newInsertTextRequest(text string, index int64, segmentId string) *docs.Request {
	return &docs.Request{
		InsertText: &docs.InsertTextRequest{
			Location: newLocation(index, segmentId),
			Text:     text,
		},
	}
}

// newDeleteContentRangeRequest creates a new DeleteContentRange request.
func newDeleteContentRangeRequest(start, end int64, segmentId string) *docs.Request {
	return &docs.Request{
		DeleteContentRange: &docs.DeleteContentRangeRequest{
			Range: newRange(start, end, segmentId),
		},
	}
}

// newUpdateTextStyleRequest creates a new UpdateTextStyle request.
func newUpdateTextStyleRequest(style *docs.TextStyle, fields string, start, end int64, segmentId string) *docs.Request {
	return &docs.Request{
		UpdateTextStyle: &docs.UpdateTextStyleRequest{
			Fields:    fields,
			TextStyle: style,
			Range:     newRange(start, end, segmentId),
		},
	}
}

// newUpdateParagraphStyleRequest creates a new UpdateParagraphStyle request.
func newUpdateParagraphStyleRequest(style *docs.ParagraphStyle, fields string, start, end int64, segmentId string) *docs.Request {
	return &docs.Request{
		UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
			Fields:         fields,
			ParagraphStyle: style,
			Range:          newRange(start, end, segmentId),
		},
	}
}

// newCreateHeaderRequest creates a new CreateHeader request.
func newCreateHeaderRequest(headerType string, location *docs.Location) *docs.Request {
	return &docs.Request{
		CreateHeader: &docs.CreateHeaderRequest{
			Type:                 headerType,
			SectionBreakLocation: location,
		},
	}
}

// newUpdateDocumentStyleRequest creates a new UpdateDocumentStyle request.
func newUpdateDocumentStyleRequest(style *docs.DocumentStyle, fields string) *docs.Request {
	return &docs.Request{
		UpdateDocumentStyle: &docs.UpdateDocumentStyleRequest{
			DocumentStyle: style,
			Fields:        fields,
		},
	}
}
