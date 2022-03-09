package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joeyave/scala-bot/entities"
	"github.com/klauspost/lctime"
	"gopkg.in/telebot.v3"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func AddCallbackData(message string, url string) string {
	message = fmt.Sprintf("%s\n<a href=\"%s\">&#8203;</a>", message, url)
	return message
}

func ParseCallbackData(data string) (int, int, string) {

	parsedData := strings.Split(data, ":")
	stateStr := parsedData[0]
	indexStr := parsedData[1]
	payload := strings.Join(parsedData[2:], ":")

	state, err := strconv.Atoi(stateStr)
	if err != nil {
		state = 0 // TODO
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		index = 0 // TODO
	}

	return state, index, payload
}

func AggregateCallbackData(state int, index int, payload string) string {
	return fmt.Sprintf("%d:%d:%s", state, index, payload)
}

func JsonEscape(i string) string {

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(i)
	if err != nil {
		panic(err)
	}

	buffer.Bytes()

	b := bytes.Trim(bytes.TrimSpace(buffer.Bytes()), `"`)

	return string(b)
}

func CleanUpQuery(query string) string {
	numbersRegex := regexp.MustCompile(`\(.*?\)|[1-9.()_]*`)
	return numbersRegex.ReplaceAllString(query, "")
}

func SplitQueryByNewlines(query string) []string {
	newLinesRegex := regexp.MustCompile(`\s*[\t\r\n]+`)
	songNames := strings.Split(newLinesRegex.ReplaceAllString(query, "\n"), "\n")
	for _, songName := range songNames {
		songName = strings.TrimSpace(songName)
	}

	return songNames
}

func EventButton(event *entities.Event, user *entities.User, showMemberships bool) string {
	str := fmt.Sprintf("%s (%s)", event.Name, lctime.Strftime("%A, %d.%m.%Y", event.Time))

	if user != nil {
		var memberships []string
		for _, membership := range event.Memberships {
			if membership.UserID == user.ID {
				memberships = append(memberships, membership.Role.Name)
			}
		}

		if len(memberships) > 0 {
			if showMemberships {
				str = fmt.Sprintf("%s [%s]", str, strings.Join(memberships, ", "))
			} else {
				str = fmt.Sprintf(" %s 🙋‍♂️", str)
			}
		}
	}

	return str
}

func ParseEventButton(str string) (string, time.Time, error) {

	regex := regexp.MustCompile(`(.*)\s\(.*,\s*([\d.]+)`)

	matches := regex.FindStringSubmatch(str)
	if len(matches) < 3 {
		return "", time.Time{}, fmt.Errorf("not all subgroup matches: %v", matches)
	}

	eventName := matches[1]

	eventTime, err := time.Parse("02.01.2006", strings.TrimSpace(matches[2]))
	if err != nil {
		return "", time.Time{}, err
	}

	return eventName, eventTime, nil
}

func IsWeekdayString(str string) bool {
	switch strings.ToLower(str) {
	case "пн", "вт", "ср", "чт", "пт", "сб", "вс":
		return true
	}
	return false
}

func GetWeekdayString(t time.Time) string {
	switch t.Weekday() {
	case time.Monday:
		return "Пн"
	case time.Tuesday:
		return "Вт"
	case time.Wednesday:
		return "Ср"
	case time.Thursday:
		return "Чт"
	case time.Friday:
		return "Пт"
	case time.Saturday:
		return "Сб"
	case time.Sunday:
		return "Вс"
	}
	return ""
}

func GetWeekdayAsString(w time.Weekday) string {
	switch w {
	case time.Monday:
		return "Пн"
	case time.Tuesday:
		return "Вт"
	case time.Wednesday:
		return "Ср"
	case time.Thursday:
		return "Чт"
	case time.Friday:
		return "Пт"
	case time.Saturday:
		return "Сб"
	case time.Sunday:
		return "Вс"
	}
	return ""
}

func GetWeekdayFromString(str string) time.Weekday {
	switch strings.ToLower(str) {
	case "пн":
		return time.Monday
	case "вт":
		return time.Tuesday
	case "ср":
		return time.Wednesday
	case "чт":
		return time.Thursday
	case "пт":
		return time.Friday
	case "сб":
		return time.Saturday
	case "вс":
		return time.Sunday
	}
	return time.Sunday
}

func GetWeekdayButtons(events []*entities.Event) []telebot.ReplyButton {
	weekdaysMap := make(map[time.Weekday]int, 0)
	for _, event := range events {
		weekdaysMap[event.Time.Weekday()] = 0
	}

	var weekdays []time.Weekday
	for weekday := range weekdaysMap {
		weekdays = append(weekdays, weekday)
	}

	sort.Slice(weekdays, func(i, j int) bool {
		wi := weekdays[i]
		wj := weekdays[j]

		if wi == 0 {
			wi = 7
		}
		if wj == 0 {
			wj = 7
		}

		return wi < wj
	})

	weekdaysButtons := []telebot.ReplyButton{{Text: GetEventsWithMe}}
	for _, weekday := range weekdays {
		weekdaysButtons = append(weekdaysButtons, telebot.ReplyButton{Text: GetWeekdayAsString(weekday)})
	}
	weekdaysButtons = append(weekdaysButtons, telebot.ReplyButton{Text: Archive})

	return weekdaysButtons
}
