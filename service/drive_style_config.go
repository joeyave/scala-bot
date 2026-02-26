package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/api/docs/v1"
	"gopkg.in/yaml.v3"
)

const (
	driveStyleConfigPathEnv     = "BOT_DRIVE_STYLE_CONFIG_PATH"
	defaultDriveStyleConfigPath = "config/drive_style.yml"
)

type DriveStyleConfig struct {
	NewDocument DriveStyleNewDocumentConfig `yaml:"new_document"`
	Document    DriveStyleDocumentConfig    `yaml:"document"`
	Text        DriveStyleTextConfig        `yaml:"text"`
	Paragraph   DriveStyleParagraphConfig   `yaml:"paragraph"`
	Metadata    DriveStyleMetadataConfig    `yaml:"metadata"`
	Chords      DriveStyleChordsConfig      `yaml:"chords"`
	Tokens      DriveStyleTokensConfig      `yaml:"tokens"`
}

type DriveStyleNewDocumentConfig struct {
	DefaultFontSizePt float64 `yaml:"default_font_size_pt"`
}

type DriveStyleDocumentConfig struct {
	Unit      string              `yaml:"unit"`
	MarginsPt DriveStyleMarginsPt `yaml:"margins_pt"`
}

type DriveStyleMarginsPt struct {
	Top    float64 `yaml:"top"`
	Right  float64 `yaml:"right"`
	Bottom float64 `yaml:"bottom"`
	Left   float64 `yaml:"left"`
	Header float64 `yaml:"header"`
}

type DriveStyleTextConfig struct {
	FontFamily string `yaml:"font_family"`
}

type DriveStyleParagraphConfig struct {
	LineSpacing  float64 `yaml:"line_spacing"`
	SpaceAbovePt float64 `yaml:"space_above_pt"`
	SpaceBelowPt float64 `yaml:"space_below_pt"`
}

type DriveStyleMetadataConfig struct {
	TitleAlignment     string  `yaml:"title_alignment"`
	LineAlignment      string  `yaml:"line_alignment"`
	LastLineAlignment  string  `yaml:"last_line_alignment"`
	FontSizeTitlePt    float64 `yaml:"font_size_title_pt"`
	FontSizeLinePt     float64 `yaml:"font_size_line_pt"`
	FontSizeLastLinePt float64 `yaml:"font_size_last_line_pt"`
	BaselineOffset     string  `yaml:"baseline_offset"`
}

type DriveStyleChordsConfig struct {
	PrimaryColor         string  `yaml:"primary_color"`
	AlternateColor       string  `yaml:"alternate_color"`
	PlainTextColor       string  `yaml:"plain_text_color"`
	SuffixStylingEnabled bool    `yaml:"suffix_styling_enabled"`
	ChordRatioThreshold  float64 `yaml:"chord_ratio_threshold"`
}

type DriveStyleTokensConfig struct {
	Enabled bool `yaml:"enabled"`
}

var (
	driveStyleConfigMu     sync.RWMutex
	activeDriveStyleConfig = DefaultDriveStyleConfig()
)

func DefaultDriveStyleConfig() DriveStyleConfig {
	return DriveStyleConfig{
		NewDocument: DriveStyleNewDocumentConfig{
			DefaultFontSizePt: 14,
		},
		Document: DriveStyleDocumentConfig{
			Unit: "PT",
			MarginsPt: DriveStyleMarginsPt{
				Top:    14,
				Right:  30,
				Bottom: 14,
				Left:   30,
				Header: 18,
			},
		},
		Text: DriveStyleTextConfig{
			FontFamily: "Roboto Mono",
		},
		Paragraph: DriveStyleParagraphConfig{
			LineSpacing:  90,
			SpaceAbovePt: 0,
			SpaceBelowPt: 0,
		},
		Metadata: DriveStyleMetadataConfig{
			TitleAlignment:     "CENTER",
			LineAlignment:      "END",
			LastLineAlignment:  "CENTER",
			FontSizeTitlePt:    20,
			FontSizeLinePt:     14,
			FontSizeLastLinePt: 11,
			BaselineOffset:     "NONE",
		},
		Chords: DriveStyleChordsConfig{
			PrimaryColor:         "#CC0000",
			AlternateColor:       "#9900FF",
			PlainTextColor:       "#000000",
			SuffixStylingEnabled: true,
			ChordRatioThreshold:  0,
		},
		Tokens: DriveStyleTokensConfig{
			Enabled: true,
		},
	}
}

func LoadAndApplyDriveStyleConfigFromEnv() error {
	path := strings.TrimSpace(os.Getenv(driveStyleConfigPathEnv))
	if path == "" {
		path = defaultDriveStyleConfigPath
	}
	return LoadAndApplyDriveStyleConfig(path)
}

func LoadAndApplyDriveStyleConfig(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("drive style config path is empty")
	}

	cfg, err := loadDriveStyleConfig(path)
	if err != nil {
		return err
	}

	setDriveStyleConfig(cfg)
	return nil
}

func loadDriveStyleConfig(path string) (DriveStyleConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return DriveStyleConfig{}, fmt.Errorf("read drive style config %q: %w", path, err)
	}

	cfg := DefaultDriveStyleConfig()
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return DriveStyleConfig{}, fmt.Errorf("parse drive style config %q: %w", path, err)
	}

	normalizeDriveStyleConfig(&cfg)
	if err := validateDriveStyleConfig(cfg); err != nil {
		return DriveStyleConfig{}, fmt.Errorf("validate drive style config %q: %w", path, err)
	}

	return cfg, nil
}

func normalizeDriveStyleConfig(cfg *DriveStyleConfig) {
	cfg.Document.Unit = strings.ToUpper(strings.TrimSpace(cfg.Document.Unit))
	cfg.Text.FontFamily = strings.TrimSpace(cfg.Text.FontFamily)
	cfg.Metadata.TitleAlignment = strings.ToUpper(strings.TrimSpace(cfg.Metadata.TitleAlignment))
	cfg.Metadata.LineAlignment = strings.ToUpper(strings.TrimSpace(cfg.Metadata.LineAlignment))
	cfg.Metadata.LastLineAlignment = strings.ToUpper(strings.TrimSpace(cfg.Metadata.LastLineAlignment))
	cfg.Metadata.BaselineOffset = strings.ToUpper(strings.TrimSpace(cfg.Metadata.BaselineOffset))
	cfg.Chords.PrimaryColor = strings.TrimSpace(cfg.Chords.PrimaryColor)
	cfg.Chords.AlternateColor = strings.TrimSpace(cfg.Chords.AlternateColor)
	cfg.Chords.PlainTextColor = strings.TrimSpace(cfg.Chords.PlainTextColor)
}

func validateDriveStyleConfig(cfg DriveStyleConfig) error {
	allowedAlignments := map[string]struct{}{
		"START":     {},
		"CENTER":    {},
		"END":       {},
		"JUSTIFIED": {},
	}
	allowedBaselineOffsets := map[string]struct{}{
		"NONE":        {},
		"SUPERSCRIPT": {},
		"SUBSCRIPT":   {},
	}

	if cfg.Document.Unit == "" {
		return fmt.Errorf("document.unit must not be empty")
	}
	if cfg.Text.FontFamily == "" {
		return fmt.Errorf("text.font_family must not be empty")
	}
	if err := validatePositive("new_document.default_font_size_pt", cfg.NewDocument.DefaultFontSizePt); err != nil {
		return err
	}
	if err := validateNonNegative("document.margins_pt.top", cfg.Document.MarginsPt.Top); err != nil {
		return err
	}
	if err := validateNonNegative("document.margins_pt.right", cfg.Document.MarginsPt.Right); err != nil {
		return err
	}
	if err := validateNonNegative("document.margins_pt.bottom", cfg.Document.MarginsPt.Bottom); err != nil {
		return err
	}
	if err := validateNonNegative("document.margins_pt.left", cfg.Document.MarginsPt.Left); err != nil {
		return err
	}
	if err := validateNonNegative("document.margins_pt.header", cfg.Document.MarginsPt.Header); err != nil {
		return err
	}
	if err := validatePositive("paragraph.line_spacing", cfg.Paragraph.LineSpacing); err != nil {
		return err
	}
	if err := validateNonNegative("paragraph.space_above_pt", cfg.Paragraph.SpaceAbovePt); err != nil {
		return err
	}
	if err := validateNonNegative("paragraph.space_below_pt", cfg.Paragraph.SpaceBelowPt); err != nil {
		return err
	}
	if _, ok := allowedAlignments[cfg.Metadata.TitleAlignment]; !ok {
		return fmt.Errorf("metadata.title_alignment must be one of START,CENTER,END,JUSTIFIED")
	}
	if _, ok := allowedAlignments[cfg.Metadata.LineAlignment]; !ok {
		return fmt.Errorf("metadata.line_alignment must be one of START,CENTER,END,JUSTIFIED")
	}
	if _, ok := allowedAlignments[cfg.Metadata.LastLineAlignment]; !ok {
		return fmt.Errorf("metadata.last_line_alignment must be one of START,CENTER,END,JUSTIFIED")
	}
	if err := validatePositive("metadata.font_size_title_pt", cfg.Metadata.FontSizeTitlePt); err != nil {
		return err
	}
	if err := validatePositive("metadata.font_size_line_pt", cfg.Metadata.FontSizeLinePt); err != nil {
		return err
	}
	if err := validatePositive("metadata.font_size_last_line_pt", cfg.Metadata.FontSizeLastLinePt); err != nil {
		return err
	}
	if _, ok := allowedBaselineOffsets[cfg.Metadata.BaselineOffset]; !ok {
		return fmt.Errorf("metadata.baseline_offset must be one of NONE,SUPERSCRIPT,SUBSCRIPT")
	}
	if err := validateHexColor("chords.primary_color", cfg.Chords.PrimaryColor); err != nil {
		return err
	}
	if err := validateHexColor("chords.alternate_color", cfg.Chords.AlternateColor); err != nil {
		return err
	}
	if err := validateHexColor("chords.plain_text_color", cfg.Chords.PlainTextColor); err != nil {
		return err
	}
	if cfg.Chords.ChordRatioThreshold < 0 || cfg.Chords.ChordRatioThreshold > 1 {
		return fmt.Errorf("chords.chord_ratio_threshold must be between 0 and 1")
	}

	return nil
}

func validatePositive(name string, value float64) error {
	if value <= 0 {
		return fmt.Errorf("%s must be > 0", name)
	}
	return nil
}

func validateNonNegative(name string, value float64) error {
	if value < 0 {
		return fmt.Errorf("%s must be >= 0", name)
	}
	return nil
}

func validateHexColor(name, color string) error {
	if _, _, _, err := parseHexColor(color); err != nil {
		return fmt.Errorf("%s must be in #RRGGBB format: %w", name, err)
	}
	return nil
}

func parseHexColor(hexColor string) (float64, float64, float64, error) {
	trimmed := strings.TrimSpace(hexColor)
	if len(trimmed) != 7 || !strings.HasPrefix(trimmed, "#") {
		return 0, 0, 0, fmt.Errorf("invalid color %q", hexColor)
	}

	r, err := strconv.ParseUint(trimmed[1:3], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid red channel in %q", hexColor)
	}
	g, err := strconv.ParseUint(trimmed[3:5], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid green channel in %q", hexColor)
	}
	b, err := strconv.ParseUint(trimmed[5:7], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid blue channel in %q", hexColor)
	}

	return float64(r) / 255, float64(g) / 255, float64(b) / 255, nil
}

func driveStyleColorFromHex(hexColor string) *docs.RgbColor {
	r, g, b, err := parseHexColor(hexColor)
	if err != nil {
		fallback := DefaultDriveStyleConfig().Chords.PlainTextColor
		r, g, b, _ = parseHexColor(fallback)
	}
	return newRgbColor(r, g, b)
}

func driveStylePlainTextColor() *docs.RgbColor {
	cfg := getDriveStyleConfig()
	return driveStyleColorFromHex(cfg.Chords.PlainTextColor)
}

func driveStyleChordPrimaryColor() *docs.RgbColor {
	cfg := getDriveStyleConfig()
	return driveStyleColorFromHex(cfg.Chords.PrimaryColor)
}

func driveStyleChordAlternateColor() *docs.RgbColor {
	cfg := getDriveStyleConfig()
	return driveStyleColorFromHex(cfg.Chords.AlternateColor)
}

func newDriveDocumentStyleFromConfig() (*docs.DocumentStyle, string) {
	cfg := getDriveStyleConfig()
	unit := cfg.Document.Unit

	style := &docs.DocumentStyle{
		MarginBottom: &docs.Dimension{Magnitude: cfg.Document.MarginsPt.Bottom, Unit: unit},
		MarginLeft:   &docs.Dimension{Magnitude: cfg.Document.MarginsPt.Left, Unit: unit},
		MarginRight:  &docs.Dimension{Magnitude: cfg.Document.MarginsPt.Right, Unit: unit},
		MarginTop:    &docs.Dimension{Magnitude: cfg.Document.MarginsPt.Top, Unit: unit},
		MarginHeader: &docs.Dimension{Magnitude: cfg.Document.MarginsPt.Header, Unit: unit},
	}

	return style, "marginBottom,marginLeft,marginRight,marginTop,marginHeader"
}

func getDriveStyleConfig() DriveStyleConfig {
	driveStyleConfigMu.RLock()
	defer driveStyleConfigMu.RUnlock()
	return activeDriveStyleConfig
}

func setDriveStyleConfig(cfg DriveStyleConfig) {
	driveStyleConfigMu.Lock()
	activeDriveStyleConfig = cfg
	driveStyleConfigMu.Unlock()
}
