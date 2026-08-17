package localfs

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type Storage struct {
	root string
}

func New(root string) (*Storage, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Storage{root: root}, nil
}

func (s *Storage) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	path := filepath.Join(s.root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *Storage) Get(_ context.Context, key string) (io.ReadCloser, string, error) {
	path := filepath.Join(s.root, filepath.FromSlash(key))
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	return f, "application/octet-stream", nil
}
