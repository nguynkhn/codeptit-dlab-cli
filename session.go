package main

import (
	"encoding/json"
	"errors"
	"os"
)

type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SignKey      []byte `json:"-"`
	SkewMs       int64  `json:"-"`
}

type Store interface {
	Load(*Session) error
	Save(*Session) error
}

type FileStore struct {
	path string
}

func (f *FileStore) Load(s *Session) error {
	dataBytes, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	if err := json.Unmarshal(dataBytes, s); err != nil {
		return err
	}

	return nil
}

func (f *FileStore) Save(s *Session) error {
	dataBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if err := os.WriteFile(f.path, dataBytes, 0o600); err != nil {
		return err
	}

	return nil
}
