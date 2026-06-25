package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const googleAPIsKeyEnv = "BOT_GOOGLEAPIS_KEY"

type googleAPIsKey struct {
	ClientEmail string `json:"client_email"`
}

func GoogleAPIsClientEmailFromJSON(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%s is empty", googleAPIsKeyEnv)
	}

	var key googleAPIsKey
	if err := json.Unmarshal([]byte(raw), &key); err != nil {
		return "", fmt.Errorf("parse %s: %w", googleAPIsKeyEnv, err)
	}

	email := strings.TrimSpace(key.ClientEmail)
	if email == "" {
		return "", fmt.Errorf("%s does not contain client_email", googleAPIsKeyEnv)
	}

	return email, nil
}

func GoogleAPIsClientEmailFromEnv() (string, error) {
	return GoogleAPIsClientEmailFromJSON(os.Getenv(googleAPIsKeyEnv))
}
