package repository

import (
	"archive/zip"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"kpo_hw3_devenv/internal/storage/services"
)

const archiveDirectory = "archives"

// ArchiveStore хранит ZIP в каталоге archives внутри каталога данных.
type ArchiveStore struct {
	dataDir string
}

func NewArchiveStore(dataDir string) (*ArchiveStore, error) {
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, archiveDirectory), 0o750); err != nil {
		return nil, fmt.Errorf("create archive directory: %w", err)
	}
	return &ArchiveStore{dataDir: dataDir}, nil
}

// Save проверяет ZIP и возвращает новый ключ и путь относительно dataDir.
func (a *ArchiveStore) Save(src io.Reader) (key, archivePath string, err error) {
	key = rand.Text()
	name := key + ".zip"
	filePath := filepath.Join(a.dataDir, archiveDirectory, name)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", "", fmt.Errorf("create archive: %w", err)
	}

	size, saveErr := io.Copy(file, src)
	if saveErr == nil {
		if err := validateZIP(file, size); err != nil {
			saveErr = err
			var readErr *os.PathError
			if !errors.As(err, &readErr) {
				saveErr = errors.Join(services.ErrInvalidArchive, err)
			}
		}
	}
	if saveErr == nil {
		saveErr = file.Sync()
	}
	if err := errors.Join(saveErr, file.Close()); err != nil {
		return "", "", fmt.Errorf("save archive: %w", errors.Join(err, os.Remove(filePath)))
	}
	return key, archiveDirectory + "/" + name, nil
}

func validateZIP(src io.ReaderAt, size int64) error {
	archive, err := zip.NewReader(src, size)
	if err != nil {
		return err
	}

	for _, entry := range archive.File {
		reader, err := entry.Open()
		if err != nil {
			return err
		}

		_, readErr := io.Copy(io.Discard, reader)
		if err := errors.Join(readErr, reader.Close()); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArchiveStore) Open(archivePath string) (io.ReadCloser, error) {
	filePath, err := a.filePath(archivePath)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	return file, nil
}

// Remove удаляет архив. Повторное удаление отсутствующего файла допустимо.
func (a *ArchiveStore) Remove(archivePath string) error {
	filePath, err := a.filePath(archivePath)
	if err != nil {
		return err
	}
	if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove archive: %w", err)
	}
	return nil
}

func (a *ArchiveStore) filePath(archivePath string) (string, error) {
	name, ok := strings.CutPrefix(archivePath, archiveDirectory+"/")
	if !ok || !filepath.IsLocal(name) || strings.ContainsAny(name, `/\`) || filepath.Ext(name) != ".zip" {
		return "", fmt.Errorf("invalid archive path %q", archivePath)
	}
	return filepath.Join(a.dataDir, archiveDirectory, name), nil
}
