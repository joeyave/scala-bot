package keyboard

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	"github.com/klauspost/lctime"
	"google.golang.org/api/drive/v3"
)

func EventButton(event *entity.Event, user *entity.User, lang string, showMemberships bool) []gotgbot.KeyboardButton {

	t, _ := lctime.StrftimeLoc(util.IetfToIsoLangCode(lang), "%A, %d.%m.%Y", event.TimeUTC)
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
			for _, like := range song.Likes {
				if user.ID == like.UserID {
					text += fmt.Sprintf(" %s", txt.Get("button.like", ""))
					break
				}
			}
		}
	}

	return []gotgbot.KeyboardButton{{Text: text}}
}

// todo: refactor
var songButtonRegEx = regexp.MustCompile(`\s*\([^()]*\)\s*(` + txt.Get("button.like", "") + `)?\s*$`)

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
	case "пн", "вт", "ср", "чт", "пт", "сб", "нд":
		return true
	}
	return false
}

func ParseWeekdayButton(text string) time.Weekday {
	switch strings.ToLower(text) {
	case "пн.", "пн":
		return time.Monday
	case "вт.", "вт":
		return time.Tuesday
	case "ср.", "ср":
		return time.Wednesday
	case "чт.", "чт":
		return time.Thursday
	case "пт.", "пт":
		return time.Friday
	case "сб.", "сб":
		return time.Saturday
	case "вс.", "нд":
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
		weekdaysMap[event.TimeUTC.Weekday()] = event.TimeUTC
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

func GetStatsPeriodButtonText(period entity.StatsPeriod, lang string, noPeriodWord bool) string {
	periodStr := ""
	switch period {
	case entity.StatsPeriodLastYear:
		periodStr = txt.Get("text.period.lastYear", lang)
	case entity.StatsPeriodLastThreeMonths:
		periodStr = txt.Get("text.period.lastThreeMonths", lang)
	case entity.StatsPeriodAllTime:
		periodStr = txt.Get("text.period.allTime", lang)
	default:
		periodStr = txt.Get("text.period.lastHalfYear", lang)
	}

	if noPeriodWord {
		return periodStr
	}

	return txt.Get("text.period", lang, periodStr)
}

func GetStatsPeriodByButtonText(text string, lang string) entity.StatsPeriod {
	switch text {
	case txt.Get("text.period.lastYear", lang):
		return entity.StatsPeriodLastYear
	case txt.Get("text.period.lastThreeMonths", lang):
		return entity.StatsPeriodLastThreeMonths
	case txt.Get("text.period.allTime", lang):
		return entity.StatsPeriodAllTime
	default:
		return entity.StatsPeriodLastHalfYear
	}
}

func GetStatsPeriodButton(period entity.StatsPeriod, lang string) []gotgbot.KeyboardButton {

	text := GetStatsPeriodButtonText(period, lang, false)
	return []gotgbot.KeyboardButton{
		{Text: text},
	}
}

func GetStatsSortingButtonText(sorting entity.StatsSorting, lang string, noSortingWord bool) string {
	str := ""
	switch sorting {
	case entity.StatsSortingAscending:
		str = txt.Get("text.sorting.ascending", lang)
	case entity.StatsSortingDescending:
		str = txt.Get("text.sorting.descending", lang)
	}

	if noSortingWord {
		return str
	}

	return txt.Get("text.sorting", lang, str)
}

func GetStatsSortingByButtonText(text string, lang string) entity.StatsSorting {
	switch text {
	case txt.Get("text.sorting.ascending", lang):
		return entity.StatsSortingAscending
	case txt.Get("text.sorting.descending", lang):
		return entity.StatsSortingDescending
	default:
		return entity.StatsSortingDescending
	}
}

func GetStatsSortingButton(sorting entity.StatsSorting, lang string) []gotgbot.KeyboardButton {

	text := GetStatsSortingButtonText(sorting, lang, false)
	return []gotgbot.KeyboardButton{
		{Text: text},
	}
}
