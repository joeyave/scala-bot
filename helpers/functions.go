package helpers

import (
	"bytes"
	"encoding/json"
)

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
