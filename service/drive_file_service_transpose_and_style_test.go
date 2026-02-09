package service

import (
	"regexp"
	"strings"
	"testing"

	"github.com/joeyave/scala-bot/entity"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/docs/v1"
)

func TestUppercasePreservingRepetition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text no markers",
			input:    "[text]",
			expected: "[TEXT]",
		},
		{
			name:     "text with single marker at end",
			input:    "[verse x2]",
			expected: "[VERSE x2]",
		},
		{
			name:     "text with multiple markers",
			input:    "[verse x2 and chorus x3]",
			expected: "[VERSE x2 AND CHORUS x3]",
		},
		{
			name:     "marker only",
			input:    "[x2]",
			expected: "[x2]",
		},
		{
			name:     "cyrillic x marker",
			input:    "[verse х2]",
			expected: "[VERSE х2]",
		},
		{
			name:     "marker at start",
			input:    "[x2 intro]",
			expected: "[x2 INTRO]",
		},
		{
			name:     "marker in middle",
			input:    "[verse x2 bridge]",
			expected: "[VERSE x2 BRIDGE]",
		},
		{
			name:     "multiple consecutive markers",
			input:    "[x2x3]",
			expected: "[x2x3]",
		},
		{
			name:     "empty brackets",
			input:    "[]",
			expected: "[]",
		},
		{
			name:     "mixed case input",
			input:    "[Verse X2]",
			expected: "[VERSE X2]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uppercasePreservingRepetition(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper to create an indexedParagraph from text for testing.
func createTestIndexedParagraph(text string, startIndex int64) *indexedParagraph {
	para := &docs.Paragraph{
		Elements: []*docs.ParagraphElement{
			{
				StartIndex: startIndex,
				EndIndex:   startIndex + int64(len([]rune(text))),
				TextRun: &docs.TextRun{
					Content: text,
				},
			},
		},
	}
	ip, _ := newIndexedParagraph(para)
	return ip
}

func createTestStructuralParagraph(text string, startIndex, endIndex int64) *docs.StructuralElement {
	return &docs.StructuralElement{
		StartIndex: startIndex,
		EndIndex:   endIndex,
		Paragraph: &docs.Paragraph{
			Elements: []*docs.ParagraphElement{
				{
					StartIndex: startIndex,
					EndIndex:   endIndex,
					TextRun: &docs.TextRun{
						Content: text,
					},
				},
			},
		},
	}
}

func TestChangeStyleByRegexAcross(t *testing.T) {
	testRegex := regexp.MustCompile(`\[[^\]]*\]`) // matches [...]
	boldStyle := docs.TextStyle{Bold: true}

	t.Run("no matches returns empty requests", func(t *testing.T) {
		ip := createTestIndexedParagraph("no brackets here", 0)
		requests := changeStyleByRegexAcross(ip, testRegex, boldStyle, "bold", nil, "")
		assert.Empty(t, requests)
	})

	t.Run("single match generates one style request", func(t *testing.T) {
		ip := createTestIndexedParagraph("[verse]", 0)
		requests := changeStyleByRegexAcross(ip, testRegex, boldStyle, "bold", nil, "")

		assert.Len(t, requests, 1)
		assert.NotNil(t, requests[0].UpdateTextStyle)
		assert.Equal(t, int64(0), requests[0].UpdateTextStyle.Range.StartIndex)
		assert.Equal(t, int64(7), requests[0].UpdateTextStyle.Range.EndIndex)
		assert.True(t, requests[0].UpdateTextStyle.TextStyle.Bold)
	})

	t.Run("multiple matches generate multiple requests", func(t *testing.T) {
		ip := createTestIndexedParagraph("[verse] text [chorus]", 0)
		requests := changeStyleByRegexAcross(ip, testRegex, boldStyle, "bold", nil, "")

		assert.Len(t, requests, 2)
		// First match: [verse] at positions 0-7
		assert.Equal(t, int64(0), requests[0].UpdateTextStyle.Range.StartIndex)
		assert.Equal(t, int64(7), requests[0].UpdateTextStyle.Range.EndIndex)
		// Second match: [chorus] at positions 13-21
		assert.Equal(t, int64(13), requests[1].UpdateTextStyle.Range.StartIndex)
		assert.Equal(t, int64(21), requests[1].UpdateTextStyle.Range.EndIndex)
	})

	t.Run("textFunc generates replace and style requests", func(t *testing.T) {
		ip := createTestIndexedParagraph("[text]", 0)
		toUpper := func(s string) string { return strings.ToUpper(s) }
		requests := changeStyleByRegexAcross(ip, testRegex, boldStyle, "bold", toUpper, "")

		// Should have: DeleteContentRange, InsertText, UpdateTextStyle
		assert.Len(t, requests, 3)
		assert.NotNil(t, requests[0].DeleteContentRange)
		assert.NotNil(t, requests[1].InsertText)
		assert.Equal(t, "[TEXT]", requests[1].InsertText.Text)
		assert.NotNil(t, requests[2].UpdateTextStyle)
	})

	t.Run("respects start index offset", func(t *testing.T) {
		ip := createTestIndexedParagraph("[verse]", 100)
		requests := changeStyleByRegexAcross(ip, testRegex, boldStyle, "bold", nil, "")

		assert.Len(t, requests, 1)
		assert.Equal(t, int64(100), requests[0].UpdateTextStyle.Range.StartIndex)
		assert.Equal(t, int64(107), requests[0].UpdateTextStyle.Range.EndIndex)
	})

	t.Run("includes segment ID in requests", func(t *testing.T) {
		ip := createTestIndexedParagraph("[verse]", 0)
		requests := changeStyleByRegexAcross(ip, testRegex, boldStyle, "bold", nil, "header-123")

		assert.Len(t, requests, 1)
		assert.Equal(t, "header-123", requests[0].UpdateTextStyle.Range.SegmentId)
	})
}

func TestGuessKeyIfNeeded(t *testing.T) {
	t.Run("keeps valid metadata key", func(t *testing.T) {
		key := guessKeyIfNeeded(entity.Key("Dm"), "C F G C")
		assert.Equal(t, entity.Key("Dm"), key)
	})

	t.Run("invalid metadata key falls back to guess", func(t *testing.T) {
		key := guessKeyIfNeeded(entity.Key("C->D"), "C F G C")
		assert.NotEqual(t, entity.Key("C->D"), key)
		assert.NotEmpty(t, string(key))
	})
}

func TestNewTransposeRequestsForParagraphAppliesParagraphStyleOnce(t *testing.T) {
	paragraph := &docs.Paragraph{
		ParagraphStyle: &docs.ParagraphStyle{Alignment: "CENTER"},
		Elements: []*docs.ParagraphElement{
			{
				TextRun: &docs.TextRun{
					Content:   "A ",
					TextStyle: &docs.TextStyle{Bold: true},
				},
			},
			{
				TextRun: &docs.TextRun{
					Content:   "B\n",
					TextStyle: &docs.TextStyle{Italic: true},
				},
			},
		},
	}

	requests, _ := newTransposeRequestsForParagraph(paragraph, false, false, "", "", "", 1, nil)

	paragraphStyleReqs := 0
	for _, req := range requests {
		if req.UpdateParagraphStyle != nil {
			paragraphStyleReqs++
		}
	}

	assert.Equal(t, 1, paragraphStyleReqs)
}

func TestGetContentForSectionBodyKeepsBoundaryParagraph(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				createTestStructuralParagraph("Line 1\n", 1, 10),
				createTestStructuralParagraph("Last line in section\n", 10, 20),
				{
					StartIndex: 20,
					EndIndex:   21,
					SectionBreak: &docs.SectionBreak{
						SectionStyle: &docs.SectionStyle{SectionType: "NEXT_PAGE"},
					},
				},
				createTestStructuralParagraph("Next section line\n", 21, 35),
			},
		},
	}

	sections := []docs.StructuralElement{
		{StartIndex: 0},
		{
			StartIndex: 20,
			SectionBreak: &docs.SectionBreak{
				SectionStyle: &docs.SectionStyle{SectionType: "NEXT_PAGE"},
			},
		},
	}

	firstBody := getContentForSectionBody(doc, sections, 0)
	assert.Len(t, firstBody, 2)
	assert.Equal(t, int64(10), firstBody[1].StartIndex)
	assert.Equal(t, "Last line in section\n", firstBody[1].Paragraph.Elements[0].TextRun.Content)
}
