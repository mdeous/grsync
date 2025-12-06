package downloader

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

// Mock the download function for testing
func mockSuccessDownload(photoPath, destDir string) (string, error) {
	return fmt.Sprintf("%s/%s", destDir, photoPath), nil
}

// Mock function that simulates different outcomes based on file name
func mockMixedDownload(photoPath, destDir string) (string, error) {
	if photoPath == "exists.jpg" {
		return "", os.ErrExist
	}
	if photoPath == "error.jpg" {
		return "", errors.New("mock download error")
	}
	return fmt.Sprintf("%s/%s", destDir, photoPath), nil
}

// TestDownloader tests the parallel downloader functionality
func TestDownloader(t *testing.T) {
	t.Run("Success with multiple workers", func(t *testing.T) {
		// Create a test downloader with 2 workers
		downloader := NewDownloader(2)
		// Override with mock function
		downloader.downloadFn = mockSuccessDownload

		// Create test jobs
		jobs := []PhotoJob{
			{PhotoPath: "test1.jpg", DestDir: "/tmp"},
			{PhotoPath: "test2.jpg", DestDir: "/tmp"},
			{PhotoPath: "test3.jpg", DestDir: "/tmp"},
			{PhotoPath: "test4.jpg", DestDir: "/tmp"},
		}

		// Start test
		downloader.DownloadAll(jobs)

		// Check results
		success, skipped, failed := downloader.GetStats()
		if success != 4 {
			t.Errorf("Expected 4 successful downloads, got %d", success)
		}
		if skipped != 0 {
			t.Errorf("Expected 0 skipped downloads, got %d", skipped)
		}
		if failed != 0 {
			t.Errorf("Expected 0 failed downloads, got %d", failed)
		}
	})

	t.Run("Handle mixed results", func(t *testing.T) {
		// Create a test downloader with 1 worker
		downloader := NewDownloader(1)
		// Override with mixed mock function
		downloader.downloadFn = mockMixedDownload

		// Create test jobs with one that will fail and one already exists
		jobs := []PhotoJob{
			{PhotoPath: "success.jpg", DestDir: "/tmp"},
			{PhotoPath: "exists.jpg", DestDir: "/tmp"}, // Will return os.ErrExist
			{PhotoPath: "error.jpg", DestDir: "/tmp"},  // Will return an error
		}

		// Start test
		downloader.DownloadAll(jobs)

		// Check results
		success, skipped, failed := downloader.GetStats()
		if success != 1 {
			t.Errorf("Expected 1 successful download, got %d", success)
		}
		if skipped != 1 {
			t.Errorf("Expected 1 skipped download, got %d", skipped)
		}
		if failed != 1 {
			t.Errorf("Expected 1 failed download, got %d", failed)
		}
	})

	t.Run("Handle empty job list", func(t *testing.T) {
		// Create a test downloader with 2 workers
		downloader := NewDownloader(2)

		// Start test with empty jobs list
		downloader.DownloadAll([]PhotoJob{})

		// Check results
		success, skipped, failed := downloader.GetStats()
		if success != 0 || skipped != 0 || failed != 0 {
			t.Errorf("Expected all zeros, got success=%d, skipped=%d, failed=%d",
				success, skipped, failed)
		}
	})
}

func TestNewDownloader(t *testing.T) {
	// Test with valid worker count
	d1 := NewDownloader(5)
	if d1.numWorkers != 5 {
		t.Errorf("Expected 5 workers, got %d", d1.numWorkers)
	}

	// Test with zero worker count (should default to 1)
	d2 := NewDownloader(0)
	if d2.numWorkers != 1 {
		t.Errorf("Expected 1 worker when initialized with 0, got %d", d2.numWorkers)
	}

	// Test with negative worker count (should default to 1)
	d3 := NewDownloader(-1)
	if d3.numWorkers != 1 {
		t.Errorf("Expected 1 worker when initialized with negative value, got %d", d3.numWorkers)
	}
}
