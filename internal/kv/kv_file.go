package kv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileStore[K, V any] struct {
	baseDir string
}

const pathSeparator = string(os.PathSeparator)

func NewFileStore[K, V any](baseDir string) (Store[K, V], error) {
	err := os.MkdirAll(baseDir, 0o755)
	if err != nil {
		return nil, err
	}
	return &FileStore[K, V]{baseDir: baseDir}, nil
}

func (s *FileStore[K, V]) Set(key K, value V) error {
	keyPath := strings.TrimSuffix(keyToString(key), pathSeparator)
	filePath := filepath.Join(s.baseDir, keyPath+".json")
	dirPath := filepath.Dir(filePath)

	err := os.MkdirAll(dirPath, 0o755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

func (s *FileStore[K, V]) Get(key K) (res V, err error) {
	keyPath := strings.TrimSuffix(keyToString(key), pathSeparator)
	filePath := filepath.Join(s.baseDir, keyPath+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return res, fmt.Errorf("key not found: %w", err)
		}
		return res, err
	}

	var value V
	err = json.Unmarshal(data, &value)
	if err != nil {
		return res, err
	}

	return value, nil
}

func (s *FileStore[K, V]) GetPrefix(key K) ([]V, error) {
	prefix := strings.TrimSuffix(keyToString(key), pathSeparator)
	values := []V{}

	err := filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		relPath, err := filepath.Rel(s.baseDir, path)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(relPath, prefix) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var value V
		err = json.Unmarshal(data, &value)
		if err != nil {
			return err
		}

		values = append(values, value)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return values, nil
}
