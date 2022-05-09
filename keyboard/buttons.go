package keyboard

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/joeyave/scala-bot-v2/entity"
	"github.com/joeyave/scala-bot-v2/txt"
	"github.com/joeyave/scala-bot-v2/util"
	"github.com/klauspost/lctime"
	"google.golang.org/api/drive/v3"
	"regexp"
	"sort"
	"strings"
	"time"
)

func EventButton(event *entity.Event, user *entity.User, lang string, showMemberships bool) []gotgbot.KeyboardButton {

	t, _ := lctime.StrftimeLoc(util.IetfToIsoLangCode(lang), "%A, %d.%m.%Y", event.Time)
	text := fmt.Sprintf("%s (%s)", event.Name, t)

	if user != nil {
		var memberships []string
		for _, membership := range event.Memberships {
			if membership.UserID == user.ID {
				memberships = append(memberships, membership.Role.Name)
			}
		}

		if len(memberships) > 0 {
			if showMemberships {
				text = fmt.Sprintf("%s [%s]", text, strings.Join(memberships, ", "))
			} else {
				text = fmt.Sprintf("%s 🙋‍♂️", text)
			}
		}
	}

	return []gotgbot.KeyboardButton{{Text: text}}
}

var eventButtonRegEx = regexp.MustCompile(`(.*)\s\(.*,\s*([\d.]+)`)

func ParseEventButton(text string) (string, time.Time, error) {

	matches := eventButtonRegEx.FindStringSubmatch(text)
	if len(matches) < 3 {
		return "", time.Time{}, fmt.Errorf("error parsing event button: %v", matches)
	}

	eventName := matches[1]

	eventTime, err := time.Parse("02.01.2006", strings.TrimSpace(matches[2]))
	if err != nil {
		return "", time.Time{}, err
	}

	return eventName, eventTime, nil
}

type SongButtonOpts struct {
	ShowStats bool
	ShowLike  bool
}

func SongButton(song *entity.SongWithEvents, user *entity.User, lang string, opts *SongButtonOpts) []gotgbot.KeyboardButton {
	text := song.PDF.Name

	if opts != nil {
		if opts.ShowStats {
			text += fmt.Sprintf(" (%s)", song.Stats(lang))
		}
		if opts.ShowLike {
			for _, userID := range song.Song.Likes {
				if user.ID == userID {
					text += fmt.Sprintf(" %s", txt.Get("button.like", ""))
					break
				}
			}
		}
	}

	return []gotgbot.KeyboardButton{{Text: text}}
}

// todo: refactor
var songButtonRegEx = regexp.MustCompile(`\s*\(.*\)\s*(` + txt.Get("button.like", "") + `)?\s*`)

func ParseSongButton(text string) string {
	return songButtonRegEx.ReplaceAllString(text, "")
}

type DriveFileButtonOpts struct {
	ShowLike bool
	ShowBand bool
}

func DriveFileButton(driveFile *drive.File, likedSongs []*entity.Song, opts *DriveFileButtonOpts) []gotgbot.KeyboardButton {
	text := driveFile.Name

	if opts != nil {
		if opts.ShowLike {
			for _, likedSong := range likedSongs {
				if likedSong.DriveFileID == driveFile.Id {
					text += fmt.Sprintf(" %s", txt.Get("button.like", ""))
					break
				}
			}
		}
	}

	return []gotgbot.KeyboardButton{{Text: text}}
}

var driveFileButtonRegEx = regexp.MustCompile(`(\s` + txt.Get("button.like", "") + `)?`)

func ParseDriveFileButton(text string) string {
	return driveFileButtonRegEx.ReplaceAllString(text, "")
}

func IsWeekdayButton(text string) bool {
	switch strings.ToLower(text) {
	case "пн.", "вт.", "ср.", "чт.", "пт.", "сб.", "вс.":
		return true
	}
	return false
}

func ParseWeekdayButton(text string) time.Weekday {
	switch strings.ToLower(text) {
	case "пн.":
		return time.Monday
	case "вт.":
		return time.Tuesday
	case "ср.":
		return time.Wednesday
	case "чт.":
		return time.Thursday
	case "пт.":
		return time.Friday
	case "сб.":
		return time.Saturday
	case "вс.":
		return time.Sunday
	}
	return time.Sunday
}

func SelectedButton(text string) gotgbot.KeyboardButton {
	selected := fmt.Sprintf("〔%s〕", text)
	button := gotgbot.KeyboardButton{Text: selected}
	return button
}

func ParseSelectedButton(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "〔", ""), "〕", "")
}

func IsSelectedButton(text string) bool {
	if strings.HasPrefix(text, "〔") && strings.HasSuffix(text, "〕") {
		return true
	}
	return false
}

func GetEventsStateFilterButtons(events []*entity.Event, lang string) []gotgbot.KeyboardButton {

	weekdaysMap := make(map[time.Weekday]time.Time, 0)
	for _, event := range events {
		weekdaysMap[event.Time.Weekday()] = event.Time
	}

	var times []time.Time
	for _, t := range weekdaysMap {
		times = append(times, t)
	}

	sort.Slice(times, func(i, j int) bool {
		timeI := times[i]
		timeJ := times[j]

		weekdayI := timeI.Weekday()
		weekdayJ := timeJ.Weekday()

		if timeI.Weekday() == 0 {
			weekdayI = 7
		}
		if timeJ.Weekday() == 0 {
			weekdayJ = 7
		}

		return weekdayI < weekdayJ
	})

	var buttons []gotgbot.KeyboardButton
	buttons = append(buttons, gotgbot.KeyboardButton{Text: txt.Get("button.eventsWithMe", lang)})
	for _, t := range times {
		text, _ := lctime.StrftimeLoc(util.IetfToIsoLangCode(lang), "%a", t)
		buttons = append(buttons, gotgbot.KeyboardButton{Text: text})
	}
	buttons = append(buttons, gotgbot.KeyboardButton{Text: txt.Get("button.archive", lang)})

	return buttons
}

func GetSongsStateFilterButtons(lang string) []gotgbot.KeyboardButton {
	return []gotgbot.KeyboardButton{
		{Text: txt.Get("button.like", lang)}, {Text: txt.Get("button.calendar", lang)}, {Text: txt.Get("button.numbers", lang)}, {Text: txt.Get("button.tag", lang)},
	}
}
