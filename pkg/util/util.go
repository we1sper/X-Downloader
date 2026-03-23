package util

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func ExtractFromCookie(cookie, key string) string {
	segments := strings.Split(cookie, ";")
	for _, segment := range segments {
		pair := strings.Split(segment, "=")
		if len(pair) > 1 && strings.TrimSpace(pair[0]) == key {
			return strings.TrimSpace(pair[1])
		}
	}
	return ""
}

func ExtractFileNameFromURL(givenUrl string) string {
	parsedURL, err := url.Parse(givenUrl)
	if err != nil {
		return ""
	}
	return filepath.Base(parsedURL.Path)
}

func IsFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func SaveFromStream(filePath string, stream io.ReadCloser) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	return err
}
