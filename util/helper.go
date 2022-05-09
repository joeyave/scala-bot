package util

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"io"
	"net/http"
)

func File(bot *gotgbot.Bot, file *gotgbot.File) (io.ReadCloser, error) {

	url := bot.GetAPIURL() + "/file/bot" + bot.Token + "/" + file.FilePath

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := bot.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("telebot: expected status 200 but got %s", resp.Status)
	}

	return resp.Body, nil
}

func SplitInlineKeyboardToColumns(k [][]gotgbot.InlineKeyboardButton, colNum int) [][]gotgbot.InlineKeyboardButton {

	var newK [][]gotgbot.InlineKeyboardButton
	var i int

	for _, row := range k {
		for _, button := range row {
			if i == colNum {
				i = 0
			}

			if i == 0 {
				newK = append(newK, []gotgbot.InlineKeyboardButton{button})
			} else if i < colNum {
				newK[len(newK)-1] = append(newK[len(newK)-1], button)
			}
			i++
		}
	}

	return newK
}
func SplitKeyboardToColumns(k [][]gotgbot.KeyboardButton, colNum int) [][]gotgbot.KeyboardButton {

	var newK [][]gotgbot.KeyboardButton
	var i int

	for _, row := range k {
		for _, button := range row {
			if i == colNum {
				i = 0
			}

			if i == 0 {
				newK = append(newK, []gotgbot.KeyboardButton{button})
			} else if i < colNum {
				newK[len(newK)-1] = append(newK[len(newK)-1], button)
			}
			i++
		}
	}

	return newK
}
