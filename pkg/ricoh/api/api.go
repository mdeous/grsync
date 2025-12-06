package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

var (
	baseURL        = "http://192.168.0.1"
	propertiesPath = "/v1/props"
	photosPath     = "/v1/photos"
)

// SetBaseURL overrides the base URL for testing purposes
func SetBaseURL(url string) {
	baseURL = url
}

func readURI(uri string) ([]byte, error) {
	resp, err := http.Get(baseURL + uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to camera API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read URI %s: status %s", uri, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func GetDeviceInfo() (props *Properties, err error) {
	data, err := readURI(propertiesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read device properties: %w", err)
	}
	err = json.Unmarshal(data, &props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device properties: %w", err)
	}

	return props, nil
}

func GetPhotos() (Photos, error) {
	data, err := readURI(photosPath)
	if err != nil {
		return Photos{}, fmt.Errorf("failed to read photos list: %w", err)
	}
	var photos Photos
	err = json.Unmarshal(data, &photos)
	if err != nil {
		return Photos{}, fmt.Errorf("failed to parse photos list: %w", err)
	}
	return photos, nil
}

func DownloadPhoto(photoPath string, destDir string) (destPath string, err error) {
	destPath = path.Join(destDir, photoPath)
	destPhotoDir := path.Dir(destPath)

	if _, err := os.Stat(destPath); err == nil {
		return "", fmt.Errorf("file already exists: %s", destPath)
	}

	if _, err := os.Stat(destPhotoDir); os.IsNotExist(err) {
		err = os.MkdirAll(destPhotoDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create destination folder: %w", err)
		}
	}

	photoUri := path.Join(photosPath, photoPath)
	photoData, err := readURI(photoUri)
	if err != nil {
		return "", fmt.Errorf("failed to download photo %s: %w", photoPath, err)
	}
	err = os.WriteFile(destPath, photoData, 0750)
	if err != nil {
		return "", fmt.Errorf("failed to write photo file: %w", err)
	}
	return destPath, nil
}
