package service

import (
	"testing"

	"github.com/joeyave/scala-bot/entity"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/docs/v1"
)

func paragraphElementFromText(text string) *docs.StructuralElement {
	return &docs.StructuralElement{
		Paragraph: &docs.Paragraph{
			Elements: []*docs.ParagraphElement{{
				TextRun: &docs.TextRun{Content: text},
			}},
		},
	}
}

func paragraphElementFromTextAt(text string, start, end int64) *docs.StructuralElement {
	return &docs.StructuralElement{
		StartIndex: start,
		EndIndex:   end,
		Paragraph: &docs.Paragraph{
			Elements: []*docs.ParagraphElement{{
				StartIndex: start,
				EndIndex:   end,
				TextRun:    &docs.TextRun{Content: text},
			}},
		},
	}
}

func TestParseMetadataLine(t *testing.T) {
	key, bpm, time, ok := parseMetadataLine("KEY: Dm; BPM: 92; TIME: 6/8;")
	assert.True(t, ok)
	assert.Equal(t, entity.Key("Dm"), key)
	assert.Equal(t, "92", bpm)
	assert.Equal(t, "6/8", time)

	_, _, _, ok = parseMetadataLine("not metadata")
	assert.False(t, ok)
}

func TestComposeCanonicalMetadataText(t *testing.T) {
	text := composeCanonicalMetadataText(SectionMetadata{Title: "", Key: "", BPM: "", Time: ""})
	assert.Equal(t, "?\nKEY: ?; BPM: ?; TIME: ?;\n\n", text)
}

func TestFindMetadataTitleAndLine(t *testing.T) {
	paragraphs := []*docs.StructuralElement{
		paragraphElementFromText("intro"),
		paragraphElementFromText("Song title"),
		paragraphElementFromText("between"),
		paragraphElementFromText("KEY: C; BPM: 120; TIME: 4/4;"),
		paragraphElementFromText("tail"),
	}

	titleIdx, metadataIdx := findMetadataTitleAndLine(paragraphs, "Song title")
	assert.Equal(t, 1, titleIdx)
	assert.Equal(t, 3, metadataIdx)
}

func TestFindMetadataTitleAndLineFallbackFirstNonEmpty(t *testing.T) {
	paragraphs := []*docs.StructuralElement{
		paragraphElementFromText("first line"),
		paragraphElementFromText("another line"),
		paragraphElementFromText("KEY: C; BPM: 120; TIME: 4/4;"),
	}

	titleIdx, metadataIdx := findMetadataTitleAndLine(paragraphs, "No Match")
	assert.Equal(t, 0, titleIdx)
	assert.Equal(t, 2, metadataIdx)
}

func TestComposeCanonicalMetadataStyleRequests(t *testing.T) {
	md := SectionMetadata{
		Title: "Song",
		Key:   "Am",
		BPM:   "120",
		Time:  "4/4",
	}

	requests := composeCanonicalMetadataStyleRequests(10, md, rgbColorChord)
	assert.Len(t, requests, 7)

	assert.Equal(t, "alignment,lineSpacing,spaceAbove,spaceBelow", requests[0].UpdateParagraphStyle.Fields)
	assert.Equal(t, "CENTER", requests[0].UpdateParagraphStyle.ParagraphStyle.Alignment)
	assert.Equal(t, int64(10), requests[0].UpdateParagraphStyle.Range.StartIndex)
	assert.Equal(t, int64(15), requests[0].UpdateParagraphStyle.Range.EndIndex)

	assert.Equal(t, "alignment,lineSpacing,spaceAbove,spaceBelow", requests[1].UpdateParagraphStyle.Fields)
	assert.Equal(t, "END", requests[1].UpdateParagraphStyle.ParagraphStyle.Alignment)

	assert.Equal(t, "*", requests[3].UpdateTextStyle.Fields)
	assert.Equal(t, fontFamilyRobotoMono, requests[3].UpdateTextStyle.TextStyle.WeightedFontFamily.FontFamily)
	assert.Equal(t, metadataFontSizeTitle, requests[3].UpdateTextStyle.TextStyle.FontSize.Magnitude)
	assert.True(t, requests[3].UpdateTextStyle.TextStyle.Bold)
	assert.Equal(t, "NONE", requests[3].UpdateTextStyle.TextStyle.BaselineOffset)
	assert.NotNil(t, requests[3].UpdateTextStyle.TextStyle.ForegroundColor)

	assert.Equal(t, "foregroundColor,bold", requests[6].UpdateTextStyle.Fields)
	assert.True(t, requests[6].UpdateTextStyle.TextStyle.Bold)
	assert.NotNil(t, requests[6].UpdateTextStyle.TextStyle.ForegroundColor)
}

func TestIsCanonicalMetadataSubsection(t *testing.T) {
	md := SectionMetadata{
		Title: "Song",
		Key:   "Am",
		BPM:   "120",
		Time:  "4/4",
	}

	canonical := []*docs.StructuralElement{
		paragraphElementFromText("Song"),
		paragraphElementFromText("KEY: Am; BPM: 120; TIME: 4/4;"),
		paragraphElementFromText(""),
	}
	assert.True(t, isCanonicalMetadataSubsection(canonical, md, 0, 1))

	nonCanonical := []*docs.StructuralElement{
		paragraphElementFromText("Song"),
		paragraphElementFromText("KEY: Am; BPM: 120; TIME: 4/4;"),
		paragraphElementFromText("extra"),
	}
	assert.False(t, isCanonicalMetadataSubsection(nonCanonical, md, 0, 1))

	nonCanonicalTitleSpace := []*docs.StructuralElement{
		paragraphElementFromText("Song "),
		paragraphElementFromText("KEY: Am; BPM: 120; TIME: 4/4;"),
		paragraphElementFromText(""),
	}
	assert.False(t, isCanonicalMetadataSubsection(nonCanonicalTitleSpace, md, 0, 1))

	nonCanonicalMetadataLeadingSpace := []*docs.StructuralElement{
		paragraphElementFromText("Song"),
		paragraphElementFromText(" KEY: Am; BPM: 120; TIME: 4/4;"),
		paragraphElementFromText(""),
	}
	assert.False(t, isCanonicalMetadataSubsection(nonCanonicalMetadataLeadingSpace, md, 0, 1))
}

func TestLeadingEmptyBodyParagraphRange(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraphElementFromTextAt("\n", 20, 21),
				paragraphElementFromTextAt("\n", 21, 22),
				paragraphElementFromTextAt("Verse line\n", 22, 33),
			},
		},
	}

	start, end, ok := leadingEmptyBodyParagraphRange(doc, 20, 33)
	assert.True(t, ok)
	assert.Equal(t, int64(20), start)
	assert.Equal(t, int64(22), end)
}

func TestLeadingEmptyBodyParagraphRangeOnlyEmpty(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraphElementFromTextAt("\n", 10, 11),
				paragraphElementFromTextAt("\n", 11, 12),
			},
		},
	}

	_, _, ok := leadingEmptyBodyParagraphRange(doc, 10, 12)
	assert.False(t, ok)
}

func TestCopyBodySectionStyleIncludesExtendedFields(t *testing.T) {
	style := &docs.SectionStyle{
		ColumnProperties: []*docs.SectionColumnProperties{
			{
				PaddingEnd: &docs.Dimension{Magnitude: 12, Unit: unitPoints},
			},
			{},
		},
		ColumnSeparatorStyle:     "BETWEEN_EACH_COLUMN",
		ContentDirection:         "LEFT_TO_RIGHT",
		MarginLeft:               &docs.Dimension{Magnitude: 30, Unit: unitPoints},
		MarginRight:              &docs.Dimension{Magnitude: 31, Unit: unitPoints},
		MarginTop:                &docs.Dimension{Magnitude: 32, Unit: unitPoints},
		MarginBottom:             &docs.Dimension{Magnitude: 33, Unit: unitPoints},
		MarginHeader:             &docs.Dimension{Magnitude: 12, Unit: unitPoints},
		MarginFooter:             &docs.Dimension{Magnitude: 11, Unit: unitPoints},
		FlipPageOrientation:      true,
		PageNumberStart:          2,
		UseFirstPageHeaderFooter: true,
	}

	copied, fields := copyBodySectionStyle(style)
	assert.NotNil(t, copied)
	assert.Contains(t, fields, "columnProperties")
	assert.Contains(t, fields, "columnSeparatorStyle")
	assert.Contains(t, fields, "contentDirection")
	assert.Contains(t, fields, "marginLeft")
	assert.Contains(t, fields, "marginRight")
	assert.Contains(t, fields, "marginTop")
	assert.Contains(t, fields, "marginBottom")
	assert.Contains(t, fields, "marginHeader")
	assert.Contains(t, fields, "marginFooter")
	assert.Contains(t, fields, "flipPageOrientation")
	assert.Contains(t, fields, "pageNumberStart")
	assert.Contains(t, fields, "useFirstPageHeaderFooter")
	assert.Equal(t, "BETWEEN_EACH_COLUMN", copied.ColumnSeparatorStyle)
	assert.Equal(t, "LEFT_TO_RIGHT", copied.ContentDirection)
	assert.Len(t, copied.ColumnProperties, 2)
	assert.NotNil(t, copied.ColumnProperties[0].PaddingEnd)
	assert.Equal(t, 12.0, copied.ColumnProperties[0].PaddingEnd.Magnitude)
	assert.NotNil(t, copied.MarginLeft)
	assert.Equal(t, 30.0, copied.MarginLeft.Magnitude)
	assert.True(t, copied.FlipPageOrientation)
	assert.Equal(t, int64(2), copied.PageNumberStart)
	assert.True(t, copied.UseFirstPageHeaderFooter)
}
