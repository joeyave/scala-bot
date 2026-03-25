package helpers

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const dotEnvFileName = ".env"

func LoadDotEnv() error {
	return loadDotEnv(dotEnvFileName)
}

func loadDotEnv(path string) error {
	err := godotenv.Load(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
