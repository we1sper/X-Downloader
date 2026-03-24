package util

import (
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
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

func SaveFromBytes(filePath string, bytes []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(bytes)
	return err
}

func LoadFile[V any](filePath string) (*V, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var result V
	if err = json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type File struct {
	Path       string
	ModifyTime time.Time
}

func FindJsonFiles(directory string) ([]*File, error) {
	files := make([]*File, 0)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			files = append(files, &File{Path: path, ModifyTime: info.ModTime()})
		}

		return nil
	})
	if err != nil {
		return files, err
	}

	// Sort files in descending order by modification time.
	slices.SortFunc(files, func(a, b *File) int {
		if a.ModifyTime.Before(b.ModifyTime) {
			return 1
		}
		return -1
	})

	return files, nil
}

func SaveToJsonFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	return SaveFromBytes(filePath, bytes)
}

func CreateDirectory(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return os.MkdirAll(directory, os.ModePerm)
	}
	return nil
}
