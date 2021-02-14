package helpers

import (
	"errors"
	"scala-chords-bot/entities"
)

func ValidateTextInput(text string, user entities.User) (entities.User, error) {
	var err error

	switch text {
	case "":
		user.State.Index--
		err = errors.New("no text in message")
	case Cancel:
		if user.State.Prev != nil {
			user.State = user.State.Prev
			user.State.Index = 0
		} else {
			user.State = &entities.State{
				Index: 0,
				Name:  MainMenuState,
			}
		}

		err = errors.New("cancel")
	}

	return user, err
}
