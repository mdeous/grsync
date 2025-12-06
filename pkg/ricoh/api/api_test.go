package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// TestGetDeviceInfo tests the GetDeviceInfo function
func TestGetDeviceInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Create a mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/props" {
				t.Errorf("Expected path /v1/props, got %s", r.URL.Path)
			}

			props := Properties{
				Model:           "GR III",
				FirmwareVersion: "1.2.3",
				Battery:         85,
				SerialNumber:    "ABC123",
			}
			json.NewEncoder(w).Encode(props)
		}))
		defer server.Close()

		// Override the base URL
		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1") // Reset after test

		// Call the function
		props, err := GetDeviceInfo()
		if err != nil {
			t.Fatalf("GetDeviceInfo failed: %v", err)
		}

		// Verify the response
		if props.Model != "GR III" {
			t.Errorf("Expected model 'GR III', got '%s'", props.Model)
		}
		if props.FirmwareVersion != "1.2.3" {
			t.Errorf("Expected firmware '1.2.3', got '%s'", props.FirmwareVersion)
		}
		if props.Battery != 85 {
			t.Errorf("Expected battery 85, got %d", props.Battery)
		}
		if props.SerialNumber != "ABC123" {
			t.Errorf("Expected serial 'ABC123', got '%s'", props.SerialNumber)
		}
	})

	t.Run("HTTPError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		_, err := GetDeviceInfo()
		if err == nil {
			t.Fatal("Expected error for HTTP 500, got nil")
		}
	})

	t.Run("MalformedJSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		_, err := GetDeviceInfo()
		if err == nil {
			t.Fatal("Expected error for malformed JSON, got nil")
		}
	})

	t.Run("NetworkError", func(t *testing.T) {
		// Point to a non-existent server
		SetBaseURL("http://localhost:0")
		defer SetBaseURL("http://192.168.0.1")

		_, err := GetDeviceInfo()
		if err == nil {
			t.Fatal("Expected network error, got nil")
		}
	})
}

// TestGetPhotos tests the GetPhotos function
func TestGetPhotos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/photos" {
				t.Errorf("Expected path /v1/photos, got %s", r.URL.Path)
			}

			photos := Photos{
				Dirs: []PhotosDir{
					{
						Name:  "100RICOH",
						Files: []string{"IMG_0001.JPG", "IMG_0002.DNG"},
					},
					{
						Name:  "101RICOH",
						Files: []string{"IMG_0003.JPG"},
					},
				},
			}
			json.NewEncoder(w).Encode(photos)
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		photos, err := GetPhotos()
		if err != nil {
			t.Fatalf("GetPhotos failed: %v", err)
		}

		if len(photos.Dirs) != 2 {
			t.Errorf("Expected 2 dirs, got %d", len(photos.Dirs))
		}
		if photos.Dirs[0].Name != "100RICOH" {
			t.Errorf("Expected dir '100RICOH', got '%s'", photos.Dirs[0].Name)
		}
		if len(photos.Dirs[0].Files) != 2 {
			t.Errorf("Expected 2 files in first dir, got %d", len(photos.Dirs[0].Files))
		}
	})

	t.Run("EmptyList", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			photos := Photos{Dirs: []PhotosDir{}}
			json.NewEncoder(w).Encode(photos)
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		photos, err := GetPhotos()
		if err != nil {
			t.Fatalf("GetPhotos failed: %v", err)
		}

		if len(photos.Dirs) != 0 {
			t.Errorf("Expected 0 dirs, got %d", len(photos.Dirs))
		}
	})

	t.Run("HTTPError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		_, err := GetPhotos()
		if err == nil {
			t.Fatal("Expected error for HTTP 404, got nil")
		}
	})

	t.Run("MalformedJSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("{bad json"))
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		_, err := GetPhotos()
		if err == nil {
			t.Fatal("Expected error for malformed JSON, got nil")
		}
	})
}

// TestDownloadPhoto tests the DownloadPhoto function
func TestDownloadPhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Create temp directory for test
		tempDir := t.TempDir()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/v1/photos/100RICOH/IMG_0001.JPG"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			w.Write([]byte("fake photo data"))
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		// Download the photo
		destPath, err := DownloadPhoto("100RICOH/IMG_0001.JPG", tempDir)
		if err != nil {
			t.Fatalf("DownloadPhoto failed: %v", err)
		}

		// Verify the file was created
		expectedPath := filepath.Join(tempDir, "100RICOH", "IMG_0001.JPG")
		if destPath != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, destPath)
		}

		// Verify the file content
		data, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}
		if string(data) != "fake photo data" {
			t.Errorf("Expected 'fake photo data', got '%s'", string(data))
		}
	})

	t.Run("FileAlreadyExists", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create the file first
		testFile := filepath.Join(tempDir, "test.jpg")
		os.WriteFile(testFile, []byte("existing"), 0644)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("new data"))
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		_, err := DownloadPhoto("test.jpg", tempDir)
		if err == nil {
			t.Fatal("Expected error for existing file, got nil")
		}
		// The error is wrapped, so check the message contains "already exists"
		if !contains(err.Error(), "already exists") {
			t.Errorf("Expected error message to contain 'already exists', got: %v", err)
		}
	})

	t.Run("DirectoryCreation", func(t *testing.T) {
		tempDir := t.TempDir()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("photo data"))
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		// Download to a nested directory that doesn't exist
		destPath, err := DownloadPhoto("dir1/dir2/photo.jpg", tempDir)
		if err != nil {
			t.Fatalf("DownloadPhoto failed: %v", err)
		}

		// Verify the directory was created
		if _, err := os.Stat(filepath.Dir(destPath)); os.IsNotExist(err) {
			t.Error("Expected directory to be created")
		}
	})

	t.Run("HTTPError", func(t *testing.T) {
		tempDir := t.TempDir()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		_, err := DownloadPhoto("notfound.jpg", tempDir)
		if err == nil {
			t.Fatal("Expected error for HTTP 404, got nil")
		}
	})

	t.Run("InvalidDestination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("data"))
		}))
		defer server.Close()

		SetBaseURL(server.URL)
		defer SetBaseURL("http://192.168.0.1")

		// Try to write to a location that should fail
		_, err := DownloadPhoto("test.jpg", "/invalid/nonexistent/path")
		if err == nil {
			t.Fatal("Expected error for invalid destination, got nil")
		}
	})
}
