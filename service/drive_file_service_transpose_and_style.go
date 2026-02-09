package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/joeyave/chords-transposer/transposer"
	"github.com/joeyave/scala-bot/entity"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// =========================================================================
// Constants & Package-Level Variables
// =========================================================================

var (
	newLineRe    = regexp.MustCompile(`\s*[\r\n]$`)
	barlineRe    = regexp.MustCompile(`[|]`)
	bracketedRe  = regexp.MustCompile(`\[[^]]*]`)
	repetitionRe = regexp.MustCompile(`([xх])\d+`)
)

const (
	unitPoints           = "PT"
	keyNashville         = "Numbers"
	fontFamilyRobotoMono = "Roboto Mono"
	maxBatchRequests     = 900

	// Margins and Spacing.
	docMarginVertical   float64 = 14
	docMarginHorizontal float64 = 30
	docMarginHeader     float64 = 18

	paraLineSpacing      float64 = 90
	paraSpacingMagnitude float64 = 0

	// Font Sizes.
	chordRatioThresholdBody float64 = 0 // todo: find a better value.
)

var (
	rgbColorChord = newRgbColor(0.8, 0, 0)
	rgbColorBlack = newRgbColor(0, 0, 0)
)

// =========================================================================
// Public Service Methods (DriveFileService)
// =========================================================================

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

func (s *DriveFileService) TransposeOne(ID string, toKey entity.Key, sectionIndex int) (*drive.File, error) {
	doc, _, err := s.normalizeMetadataLayout(ID)
	if err != nil {
		return nil, err
	}

	sections := getSections(doc)
	createdNewSection := false

	if len(sections) <= sectionIndex || sectionIndex < 0 {
		createdNewSection = true
		sections, err = s.appendSectionByID(ID)
		if err != nil {
			return nil, err
		}

		doc, err = s.ensureBodyMetadataLayout(ID)
		if err != nil {
			return nil, err
		}

		sections = getSections(doc)
		sectionIndex = len(sections) - 1
	}

	sourceMetadata := s.extractSectionMetadata(doc, sections[0])
	requests := transposeBody(doc, sections, sectionIndex, sourceMetadata.Key, toKey)
	if createdNewSection {
		targetBodyStart := getSectionBodyStartIndex(doc, sections, sectionIndex)
		sourceBodyStyle := getSectionBodyStyle(doc, sections, 0)
		if bodyStyle, fields := copyBodySectionStyle(sourceBodyStyle); bodyStyle != nil && fields != "" {
			if req := sectionStyleUpdateRequest(targetBodyStart, bodyStyle, fields); req != nil {
				requests = append(requests, req)
			}
		}
	}

	targetMetadata := sourceMetadata
	targetMetadata.Title = normalizeTextValue(doc.Title)
	targetMetadata.Key = toKey
	metadataReqs, err := metadataRewriteRequestsForSection(doc, sections, sectionIndex, targetMetadata)
	if err != nil {
		return nil, err
	}
	requests = append(requests, metadataReqs...)

	_, err = s.batchUpdate(doc.DocumentId, requests)
	if err != nil {
		return nil, err
	}

	return s.FindOneByID(ID)
}

// CopyAndTransposeFirstSection copies a file to the temp folder and transposes section 0.
// Returns the new file's Drive file. On transpose error, the copied file is deleted.
func (s *DriveFileService) CopyAndTransposeFirstSection(sourceID string, sourceName string, toKey entity.Key, tempFolderID string) (*drive.File, error) {
	copyTarget := &drive.File{
		Parents: []string{tempFolderID},
	}
	if strings.TrimSpace(sourceName) != "" {
		// Prevent Drive default "Copy of ..." naming to keep metadata title stable.
		copyTarget.Name = sourceName
	}

	copiedFile, err := s.CloneOne(sourceID, copyTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	_, err = s.TransposeOne(copiedFile.Id, toKey, 0)
	if err != nil {
		// Clean up copied file on error
		_ = s.driveClient.Files.Delete(copiedFile.Id).Do()
		return nil, fmt.Errorf("failed to transpose: %w", err)
	}

	return copiedFile, nil
}

func (s *DriveFileService) StyleOne(ID, lang string) (*drive.File, error) {
	requests := make([]*docs.Request, 0)

	doc, _, err := s.normalizeMetadataLayoutWithOptions(ID, normalizeMetadataLayoutOptions{
		applyMetadataStyles: false,
	})
	if err != nil {
		return nil, err
	}
	_ = lang
	sections := getSections(doc)

	// Hard mode: always restore canonical metadata styles on StyleOne.
	for _, section := range sections {
		sectionStart := section.StartIndex + 1
		md := s.extractSectionMetadata(doc, section)
		md.Title = normalizeTextValue(doc.Title)
		requests = append(requests, composeCanonicalMetadataStyleRequests(sectionStart, md)...)
	}

	requests = append(requests, composeStyleRequests(getContentForSectionBody(doc, sections, 0), "", false, chordRatioThresholdBody)...)

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
	}
	requests = append(requests, newUpdateDocumentStyleRequest(docStyle, "marginBottom,marginLeft,marginRight,marginTop,marginHeader"))

	_, err = s.batchUpdate(ID, requests)
	if err != nil {
		return nil, err
	}

	// Keep non-primary sections in sync with the latest first-section content.
	if len(sections) > 1 {
		doc, err = s.getDoc(ID)
		if err != nil {
			return nil, err
		}
		sections = getSections(doc)
		if len(sections) > 1 {
			sourceMetadata := s.extractSectionMetadata(doc, sections[0])

			type retransposeTarget struct {
				sectionIndex int
				toKey        entity.Key
			}
			targets := make([]retransposeTarget, 0, len(sections)-1)
			// Process from the last section backwards so index shifts in later
			// sections never affect pending operations for earlier sections.
			for i := len(sections) - 1; i >= 1; i-- {
				targetMetadata := s.extractSectionMetadata(doc, sections[i])
				if targetMetadata.Key == sourceMetadata.Key {
					continue
				}
				if !isTranspositionTargetKeySupported(targetMetadata.Key) {
					continue
				}
				targets = append(targets, retransposeTarget{
					sectionIndex: i,
					toKey:        targetMetadata.Key,
				})
			}

			pendingRequests := make([]*docs.Request, 0)
			flushPending := func() error {
				if len(pendingRequests) == 0 {
					return nil
				}
				_, batchErr := s.batchUpdate(ID, pendingRequests)
				pendingRequests = pendingRequests[:0]
				return batchErr
			}

			for _, target := range targets {
				targetRequests := transposeBody(doc, sections, target.sectionIndex, sourceMetadata.Key, target.toKey)
				targetMetadata := sourceMetadata
				targetMetadata.Title = normalizeTextValue(doc.Title)
				targetMetadata.Key = target.toKey
				metadataReqs, reqErr := metadataRewriteRequestsForSection(doc, sections, target.sectionIndex, targetMetadata)
				if reqErr != nil {
					return nil, reqErr
				}
				targetRequests = append(targetRequests, metadataReqs...)

				if len(targetRequests) >= maxBatchRequests {
					if err := flushPending(); err != nil {
						return nil, err
					}
					if _, err := s.batchUpdate(ID, targetRequests); err != nil {
						return nil, err
					}
					continue
				}

				if len(pendingRequests)+len(targetRequests) > maxBatchRequests {
					if err := flushPending(); err != nil {
						return nil, err
					}
				}

				pendingRequests = append(pendingRequests, targetRequests...)
			}

			if err := flushPending(); err != nil {
				return nil, err
			}
		}
	}

	return s.FindOneByID(ID)
}

// =========================================================================
// Core Logic - Transposition
// =========================================================================

func transposeBody(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int, key, toKey entity.Key) []*docs.Request {
	requests := make([]*docs.Request, 0)

	sectionToInsertStartIndex := getSectionBodyStartIndex(doc, sections, sectionIndex)
	var sectionToInsertEndIndex int64

	if len(sections) > sectionIndex+1 {
		sectionToInsertEndIndex = sections[sectionIndex+1].StartIndex - 1
	} else {
		lastContentElement := doc.Body.Content[len(doc.Body.Content)-1]
		sectionToInsertEndIndex = lastContentElement.EndIndex - 1
	}

	content := getContentForSectionBody(doc, sections, 0)

	if sectionToInsertEndIndex-sectionToInsertStartIndex > 0 {
		requests = append(requests, newDeleteContentRangeRequest(sectionToInsertStartIndex, sectionToInsertEndIndex, ""))
	}

	transposeRequests, _ := composeTransposeRequests(
		content,
		sectionToInsertStartIndex,
		key,
		toKey,
		"",
		chordRatioThresholdBody,
	)
	requests = append(requests, transposeRequests...)

	return requests
}

func composeTransposeRequests(content []*docs.StructuralElement, index int64, key, toKey entity.Key, segmentId string, chordRatioThreshold float64) ([]*docs.Request, entity.Key) {
	allRequests := make([]*docs.Request, 0)
	paragraphs, idxs := getParagraphs(content)

	for i, paragraph := range paragraphs {
		fullText := idxs[i].fullText

		// Decide if this paragraph should be treated as chords
		shouldTranspose := shouldTransposeParagraph(fullText, chordRatioThreshold)

		// Determine a paragraph-level key once (fall back to curKey)
		key = guessKeyIfNeeded(key, fullText)

		// Generate requests for all elements in this paragraph
		isLastParagraph := i == len(paragraphs)-1
		paraRequests, newIndex := newTransposeRequestsForParagraph(
			paragraph, isLastParagraph, shouldTranspose,
			key, toKey, segmentId, index,
		)

		allRequests = append(allRequests, paraRequests...)
		index = newIndex // Update the index for the next paragraph
	}

	return allRequests, key
}

// =========================================================================
// Core Logic - Styling
// =========================================================================

func composeStyleRequests(content []*docs.StructuralElement, segmentID string, isHeader bool, chordRatioThreshold float64) []*docs.Request {
	requests := make([]*docs.Request, 0)
	_ = isHeader

	for _, paragraph := range content {
		if paragraph.Paragraph == nil {
			continue
		}

		// 1) Paragraph-level spacing
		paragraphStyle := docs.ParagraphStyle{
			SpaceAbove:  &docs.Dimension{Magnitude: paraSpacingMagnitude, Unit: unitPoints, ForceSendFields: []string{"Magnitude"}},
			SpaceBelow:  &docs.Dimension{Magnitude: paraSpacingMagnitude, Unit: unitPoints, ForceSendFields: []string{"Magnitude"}},
			LineSpacing: paraLineSpacing,
		}

		paragraphStyleFields := "lineSpacing,spaceAbove,spaceBelow"
		requests = append(requests, newUpdateParagraphStyleRequest(&paragraphStyle, paragraphStyleFields, paragraph.StartIndex, paragraph.EndIndex, segmentID))

		// 2) Ensure all runs use Roboto Mono
		requests = append(requests, newBaseTextStyleRequests(paragraph.Paragraph, segmentID)...)

		// Build the index ONCE for this paragraph
		ip, ok := newIndexedParagraph(paragraph.Paragraph)
		if !ok {
			continue
		}

		// 3) Style chords across the whole paragraph (paragraph-level heuristic)
		requests = append(requests, changeStyleForChordsAcross(ip, segmentID, chordRatioThreshold)...)

		// [|] -> bold, black
		textStyle := docs.TextStyle{
			Bold:            true,
			ForegroundColor: newOptionalColor(rgbColorBlack),
		}
		requests = append(requests, changeStyleByRegexAcross(ip, barlineRe, textStyle, "bold,foregroundColor", nil, segmentID)...)

		// [ ... ] -> bold + uppercase (preserving repetition markers)
		textStyle = docs.TextStyle{
			Bold: true,
		}
		requests = append(requests, changeStyleByRegexAcross(ip, bracketedRe, textStyle, "bold", uppercasePreservingRepetition, segmentID)...)

		// (x|х)\d+ -> bold, red-ish
		textStyle = docs.TextStyle{
			Bold:            true,
			ForegroundColor: newOptionalColor(rgbColorChord),
		}
		requests = append(requests, changeStyleByRegexAcross(ip, repetitionRe, textStyle, "bold,foregroundColor", nil, segmentID)...)
	}

	return requests
}

// =========================================================================
// Paragraph / TextRun Utilities
// =========================================================================

type paraSlice struct {
	el *docs.ParagraphElement
	// paragraph-relative rune offsets for this element's content
	start int64
	end   int64
}

// indexedParagraph holds a "flat" view of a paragraph for easy manipulation.
type indexedParagraph struct {
	para     *docs.Paragraph
	fullText string
	slices   []paraSlice
}

// newIndexedParagraph builds the index.
func newIndexedParagraph(paragraph *docs.Paragraph) (*indexedParagraph, bool) {
	if paragraph == nil || paragraph.Elements == nil {
		return nil, false
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
		return nil, false
	}
	return &indexedParagraph{
		para:     paragraph,
		fullText: full,
		slices:   slices,
	}, true
}

// toDocRange converts paragraph-relative rune offsets to absolute doc range.
func (ip *indexedParagraph) toDocRange(runeStart, runeEnd int64) (int64, int64, bool) {
	var first *docs.ParagraphElement
	var last *docs.ParagraphElement
	var firstElementOffset, lastElementOffset int64
	for _, slice := range ip.slices {
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

// byteToRune converts a byte index (from Go regex) into a rune count prefix.
func (ip *indexedParagraph) byteToRune(byteIdx int) int64 {
	return int64(len([]rune(ip.fullText[:byteIdx])))
}

// uppercasePreservingRepetition uppercases a string while preserving
// substrings matching repetitionRe in their original case.
func uppercasePreservingRepetition(s string) string {
	matches := repetitionRe.FindAllStringIndex(s, -1)
	if matches == nil {
		return strings.ToUpper(s)
	}

	var result strings.Builder
	lastEnd := 0
	for _, match := range matches {
		// Uppercase text before this repetition marker
		result.WriteString(strings.ToUpper(s[lastEnd:match[0]]))
		// Keep repetition marker as-is
		result.WriteString(s[match[0]:match[1]])
		lastEnd = match[1]
	}
	// Uppercase remaining text after last match
	result.WriteString(strings.ToUpper(s[lastEnd:]))
	return result.String()
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

// changeStyleByRegexAcross applies style (and optional textFunc transforms) for matches
// that may span multiple ParagraphElements inside the given paragraph.
func changeStyleByRegexAcross(ip *indexedParagraph, regex *regexp.Regexp, style docs.TextStyle, fields string, textFunc func(string) string, segmentID string) []*docs.Request {
	requests := make([]*docs.Request, 0)

	// Find matches on concatenated text (regex gives byte offsets)
	matches := regex.FindAllStringIndex(ip.fullText, -1)
	if matches == nil {
		return requests
	}

	for _, match := range matches {
		runeStart := ip.byteToRune(match[0])
		runeEnd := ip.byteToRune(match[1])
		if runeStart == runeEnd {
			continue
		}

		docStart, docEnd, ok := ip.toDocRange(runeStart, runeEnd)
		if !ok {
			continue
		}

		// Optional replacement before styling
		if textFunc != nil {
			originalText := ip.fullText[match[0]:match[1]]
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
func changeStyleForChordsAcross(ip *indexedParagraph, segmentID string, chordRatioThreshold float64) []*docs.Request {
	requests := make([]*docs.Request, 0)

	// Tokenize the full paragraph (heuristic is inside Tokenize via ChordRatioThreshold)
	lines := transposer.Tokenize(ip.fullText, true, false, &transposer.TransposeOpts{
		ChordRatioThreshold: chordRatioThreshold,
	})

	// Style for chords
	chordStyle := docs.TextStyle{
		Bold:            true,
		ForegroundColor: newOptionalColor(rgbColorChord),
	}

	chordSuffixStyle := docs.TextStyle{
		BaselineOffset: "SUPERSCRIPT",
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
			docStart, docEnd, ok := ip.toDocRange(runeStart, runeEnd)
			if !ok {
				continue
			}
			requests = append(requests, styleRange(docStart, docEnd, chordStyle, "bold,foregroundColor", segmentID))

			runeStart = token.Offset + int64(len([]rune(token.Chord.Root)))
			runeEnd = runeStart + int64(len([]rune(token.Chord.Suffix)))
			if mSuffix := token.Chord.MinorSuffix(); mSuffix != "" {
				runeStart += int64(len([]rune(mSuffix)))
			}
			if runeStart == runeEnd {
				continue
			}
			docStart, docEnd, ok = ip.toDocRange(runeStart, runeEnd)
			if !ok {
				continue
			}
			requests = append(requests, styleRange(docStart, docEnd, chordSuffixStyle, "baselineOffset", segmentID))
		}
	}

	return requests
}

// --- Helper Functions ---

func getSectionBodyStartIndex(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int) int64 {
	sectionStart := sections[sectionIndex].StartIndex + 1
	sectionEnd := contentEndForSection(doc, sections, sectionIndex)
	continuous := findFirstContinuousBreakInRange(doc, sectionStart, sectionEnd)
	if continuous == nil {
		return sectionStart
	}
	return continuous.StartIndex + 1
}

func getContentForSectionBody(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int) []*docs.StructuralElement {
	bodyStart := getSectionBodyStartIndex(doc, sections, sectionIndex)
	bodyEnd := contentEndForSection(doc, sections, sectionIndex)
	bodyEndExclusive := bodyEnd + 1

	content := make([]*docs.StructuralElement, 0)
	for _, item := range doc.Body.Content {
		// Use start-based boundaries to avoid dropping the last paragraph that
		// ends exactly at the section boundary (EndIndex semantics are exclusive).
		if item.StartIndex < bodyStart || item.StartIndex >= bodyEndExclusive {
			continue
		}
		if item.SectionBreak != nil {
			continue
		}
		content = append(content, item)
	}
	return content
}

func getSectionBodyStyle(doc *docs.Document, sections []docs.StructuralElement, sectionIndex int) *docs.SectionStyle {
	sectionStart := sections[sectionIndex].StartIndex + 1
	sectionEnd := contentEndForSection(doc, sections, sectionIndex)
	continuous := findFirstContinuousBreakInRange(doc, sectionStart, sectionEnd)
	if continuous != nil && continuous.SectionBreak != nil {
		return continuous.SectionBreak.SectionStyle
	}
	if sections[sectionIndex].SectionBreak != nil {
		return sections[sectionIndex].SectionBreak.SectionStyle
	}
	return nil
}

// getParagraphs filters content for valid paragraphs and builds an indexedParagraph for each.
func getParagraphs(content []*docs.StructuralElement) ([]*docs.Paragraph, []*indexedParagraph) {
	var paragraphs []*docs.Paragraph
	var ips []*indexedParagraph
	for _, item := range content {
		if item.Paragraph == nil || item.Paragraph.Elements == nil {
			continue
		}
		ip, ok := newIndexedParagraph(item.Paragraph)
		if !ok {
			continue
		}
		paragraphs = append(paragraphs, item.Paragraph)
		ips = append(ips, ip)
	}
	return paragraphs, ips
}

// newBaseTextStyleRequests applies the base font (Roboto Mono), weight, and header-specific font sizes.
func newBaseTextStyleRequests(paragraph *docs.Paragraph, segmentID string) []*docs.Request {
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
			textStyle.Bold = existingTextStyle.Bold
		}

		requests = append(requests, newUpdateTextStyleRequest(textStyle, textStyleFields, element.StartIndex, element.EndIndex, segmentID))
	}
	return requests
}

// shouldTransposeParagraph decides if a paragraph should be transposed based on its chord ratio.
func shouldTransposeParagraph(fullText string, chordRatioThreshold float64) bool {
	if chordRatioThreshold <= 0 {
		return true // No heuristic, default to transposing
	}

	lines := transposer.Tokenize(fullText, true, false, &transposer.TransposeOpts{
		ChordRatioThreshold: chordRatioThreshold,
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
func guessKeyIfNeeded(currentKey entity.Key, fullText string) entity.Key {
	currentKey = entity.Key(strings.TrimSpace(string(currentKey)))
	if currentKey != "" {
		if _, err := transposer.ParseKey(string(currentKey)); err == nil {
			return currentKey // Valid key is already set, do nothing
		}
		// Non-single-key values (e.g. "C->D") should not block auto-guessing.
		currentKey = ""
	}
	if guessedKey, err := transposer.GuessKeyFromText(fullText); err == nil {
		return entity.Key(guessedKey.String()) // Guessed a new key
	}
	return currentKey // Guessing failed, return original (empty) key
}

func isTranspositionTargetKeySupported(key entity.Key) bool {
	normalized := strings.TrimSpace(string(key))
	if normalized == "" || normalized == "?" {
		return false
	}
	if normalized == keyNashville {
		return true
	}
	_, err := transposer.ParseKey(normalized)
	return err == nil
}

// newTransposeRequestsForParagraph generates all the requests for a single paragraph's elements.
func newTransposeRequestsForParagraph(paragraph *docs.Paragraph, isLastParagraph, shouldTranspose bool, key, toKey entity.Key, segmentId string, index int64) ([]*docs.Request, int64) {
	requests := make([]*docs.Request, 0)
	paragraphRangeStart := int64(-1)
	paragraphRangeEnd := int64(-1)

	for j, element := range paragraph.Elements {
		if element.TextRun == nil || element.TextRun.Content == "" {
			continue
		}

		runText := element.TextRun.Content
		textStyle := &docs.TextStyle{}
		if element.TextRun.TextStyle != nil {
			copiedStyle := *element.TextRun.TextStyle
			textStyle = &copiedStyle
		}

		// Clean newline from the very last element of the content
		isLastElement := j == len(paragraph.Elements)-1

		if isLastParagraph && isLastElement {
			runText = newLineRe.ReplaceAllString(runText, " ")
		}

		if shouldTranspose && key != "" {
			var transposedText string
			var err error
			if toKey == keyNashville {
				transposedText, err = transposer.TransposeToNashville(runText, string(key))
			} else {
				transposedText, err = transposer.TransposeToKey(runText, string(key), string(toKey))
			}
			if err == nil {
				runText = transposedText
			}
		}

		if textStyle.ForegroundColor == nil {
			textStyle.ForegroundColor = newOptionalColor(rgbColorBlack)
		}

		runStart := index
		runTextLen := int64(len([]rune(runText)))
		endIndex := index + runTextLen

		// Insert the (possibly transposed) text and reapply element style for that span.
		requests = append(requests,
			newInsertTextRequest(runText, index, segmentId),
			newUpdateTextStyleRequest(textStyle, "*", index, endIndex, segmentId),
		)

		index += runTextLen
		if paragraphRangeStart < 0 {
			paragraphRangeStart = runStart
		}
		paragraphRangeEnd = endIndex
	}

	// Paragraph style is identical across the full paragraph; apply once.
	if paragraph.ParagraphStyle != nil && paragraphRangeStart >= 0 && paragraphRangeEnd > paragraphRangeStart {
		requests = append(requests, newUpdateParagraphStyleRequest(
			paragraph.ParagraphStyle,
			"alignment, lineSpacing, direction, spaceAbove, spaceBelow",
			paragraphRangeStart,
			paragraphRangeEnd,
			segmentId,
		))
	}

	return requests, index
}

// =========================================================================
// Google Docs Request Builders
// =========================================================================

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

// newUpdateDocumentStyleRequest creates a new UpdateDocumentStyle request.
func newUpdateDocumentStyleRequest(style *docs.DocumentStyle, fields string) *docs.Request {
	return &docs.Request{
		UpdateDocumentStyle: &docs.UpdateDocumentStyleRequest{
			DocumentStyle: style,
			Fields:        fields,
		},
	}
}
