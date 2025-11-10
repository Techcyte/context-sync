package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"tcs/internal/model"
)

func FormatJson(input []byte) (string, error) {
	var formattedJson bytes.Buffer
	err := json.Indent(&formattedJson, input, "\t", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("\t%v", formattedJson.String()), nil
}

func PrettyPrintMessage(message model.Message) (string, error) {
	bytes, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	return FormatJson(bytes)
}
