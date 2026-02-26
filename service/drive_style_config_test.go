package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAndApplyDriveStyleConfig(t *testing.T) {
	original := getDriveStyleConfig()
	t.Cleanup(func() {
		setDriveStyleConfig(original)
	})

	dir := t.TempDir()
	configPath := filepath.Join(dir, "drive_style.yml")
	config := `
new_document:
  default_font_size_pt: 16

document:
  unit: PT
  margins_pt:
    top: 12
    right: 28
    bottom: 12
    left: 28
    header: 16

text:
  font_family: "Courier New"

paragraph:
  line_spacing: 100
  space_above_pt: 1
  space_below_pt: 2

metadata:
  title_alignment: CENTER
  line_alignment: END
  last_line_alignment: CENTER
  font_size_title_pt: 22
  font_size_line_pt: 15
  font_size_last_line_pt: 12
  baseline_offset: NONE

chords:
  primary_color: "#112233"
  alternate_color: "#445566"
  plain_text_color: "#101010"
  suffix_styling_enabled: false
  chord_ratio_threshold: 0.2

tokens:
  enabled: false
`
	assert.NoError(t, os.WriteFile(configPath, []byte(config), 0o644))

	err := LoadAndApplyDriveStyleConfig(configPath)
	assert.NoError(t, err)

	cfg := getDriveStyleConfig()
	assert.Equal(t, 16.0, cfg.NewDocument.DefaultFontSizePt)
	assert.Equal(t, "Courier New", cfg.Text.FontFamily)
	assert.Equal(t, "#112233", cfg.Chords.PrimaryColor)
	assert.Equal(t, 0.2, cfg.Chords.ChordRatioThreshold)
	assert.False(t, cfg.Chords.SuffixStylingEnabled)
	assert.False(t, cfg.Tokens.Enabled)
}

func TestLoadAndApplyDriveStyleConfigMissingFile(t *testing.T) {
	original := getDriveStyleConfig()
	t.Cleanup(func() {
		setDriveStyleConfig(original)
	})

	err := LoadAndApplyDriveStyleConfig(filepath.Join(t.TempDir(), "missing.yml"))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "read drive style config")
}

func TestLoadAndApplyDriveStyleConfigInvalidYAML(t *testing.T) {
	original := getDriveStyleConfig()
	t.Cleanup(func() {
		setDriveStyleConfig(original)
	})

	dir := t.TempDir()
	configPath := filepath.Join(dir, "drive_style.yml")
	assert.NoError(t, os.WriteFile(configPath, []byte("new_document: ["), 0o644))

	err := LoadAndApplyDriveStyleConfig(configPath)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "parse drive style config")
}

func TestLoadAndApplyDriveStyleConfigInvalidColor(t *testing.T) {
	original := getDriveStyleConfig()
	t.Cleanup(func() {
		setDriveStyleConfig(original)
	})

	dir := t.TempDir()
	configPath := filepath.Join(dir, "drive_style.yml")
	config := "chords:\n  primary_color: \"red\"\n"
	assert.NoError(t, os.WriteFile(configPath, []byte(config), 0o644))

	err := LoadAndApplyDriveStyleConfig(configPath)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "chords.primary_color")
	assert.Equal(t, original, getDriveStyleConfig())
}
