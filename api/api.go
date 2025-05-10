package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

const (
	host           = "http://192.168.0.1"
	propertiesPath = "/v1/props"
	photosPath     = "/v1/photos"
)

func readURI(uri string) ([]byte, error) {
	resp, err := http.Get(host + uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read URI %s: %v", uri, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func GetDeviceInfo() (props *Properties, err error) {
	data, err := readURI(propertiesPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &props)
	if err != nil {
		return nil, err
	}

	return props, nil
}

func GetPhotos() (Photos, error) {
	data, err := readURI(photosPath)
	if err != nil {
		return Photos{}, err
	}
	var photos Photos
	err = json.Unmarshal(data, &photos)
	if err != nil {
		return Photos{}, err
	}
	return photos, nil
}

func DownloadPhoto(photoPath string, destDir string) (destPath string, err error) {
	// Determine destination file and folder
	destPath = path.Join(destDir, photoPath)
	destPhotoDir := path.Dir(destPath)

	// Check if destination file already exists
	if _, err := os.Stat(destPath); os.IsExist((err)) {
		return "", fmt.Errorf("file already exists: %s", destPath)
	}

	// Check if destination folder exists
	if _, err := os.Stat(destPhotoDir); os.IsNotExist(err) {
		err = os.MkdirAll(destPhotoDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create destination folder: %v", err)
		}
	}

	// Download photo
	photoUri := photosPath + photoPath
	photoData, err := readURI(photoUri)
	if err != nil {
		return "", fmt.Errorf("failed to download photo %s: %v", photoPath, err)
	}
	err = os.WriteFile(destPath, photoData, 0750)
	if err != nil {
		return "", fmt.Errorf("failed to write photo file: %v", err)
	}
	return destPath, nil
}
