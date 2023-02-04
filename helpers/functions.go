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

func Chunk[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
