package storage

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

type ImageStorage struct {
	mu          sync.Mutex
	imageFolder string
	images      map[string]*ImageInfo
}

type ImageInfo struct {
	LaptopID string
	Type     string
	Path     string
}

func NewImageStorage(imageFolder string) *ImageStorage {
	return &ImageStorage{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo),
	}
}

func (storage *ImageStorage) Save(
	laptopID string,
	imageType string,
	imageData bytes.Buffer,
) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate id: (%w)", err)
	}
	imagePath := fmt.Sprintf("%s/%s%s", storage.imageFolder, imageID, imageType)
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: (%w)", err)
	}

	if _, err = imageData.WriteTo(file); err != nil {
		return "", fmt.Errorf("cannot write image data to file: (%w)", err)
	}
	storage.mu.Lock()
	defer storage.mu.Unlock()
	storage.images[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type:     imageType,
		Path:     imagePath,
	}
	return imageID.String(), nil
}
