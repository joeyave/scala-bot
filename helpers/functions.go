package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/klauspost/lctime"
	"regexp"
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
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
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
