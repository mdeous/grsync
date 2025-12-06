package downloader

import (
	"os"
	"sync"

	"github.com/mdeous/grsync/internal/logger"
	"github.com/mdeous/grsync/pkg/ricoh/api"
)

// PhotoJob represents a photo download job
type PhotoJob struct {
	PhotoPath string
	DestDir   string
}

// PhotoResult represents the result of a photo download job
type PhotoResult struct {
	PhotoPath string
	DestPath  string
	Skipped   bool
	Err       error
}

// DownloadFunc is the function signature for photo downloading
type DownloadFunc func(photoPath, destDir string) (string, error)

// Downloader manages parallel photo downloads
type Downloader struct {
	numWorkers   int
	jobs         chan PhotoJob
	results      chan PhotoResult
	wg           sync.WaitGroup
	successCount int
	skippedCount int
	failedCount  int
	resultsMutex sync.Mutex
	downloadFn   DownloadFunc
}

// NewDownloader creates a new parallel downloader with the specified number of workers
func NewDownloader(numWorkers int) *Downloader {
	// Ensure at least 1 worker
	if numWorkers < 1 {
		numWorkers = 1
	}

	return &Downloader{
		numWorkers:   numWorkers,
		jobs:         make(chan PhotoJob),
		results:      make(chan PhotoResult),
		successCount: 0,
		skippedCount: 0,
		failedCount:  0,
		downloadFn:   api.DownloadPhoto,
	}
}

// worker processes download jobs
func (d *Downloader) worker() {
	defer d.wg.Done()

	for job := range d.jobs {
		destPath, err := d.downloadFn(job.PhotoPath, job.DestDir)
		result := PhotoResult{PhotoPath: job.PhotoPath}

		if err != nil {
			if os.IsExist(err) {
				result.Skipped = true
				result.DestPath = job.DestDir
			} else {
				result.Err = err
			}
		} else {
			result.DestPath = destPath
		}

		d.results <- result
	}
}

// startWorkers launches the worker pool
func (d *Downloader) startWorkers() {
	logger.SubDetail(1, "Starting %d download workers", d.numWorkers)
	d.wg.Add(d.numWorkers)
	for i := 0; i < d.numWorkers; i++ {
		go d.worker()
	}
}

// GetStats returns the download statistics
func (d *Downloader) GetStats() (success, skipped, failed int) {
	d.resultsMutex.Lock()
	defer d.resultsMutex.Unlock()

	return d.successCount, d.skippedCount, d.failedCount
}

// DownloadAll downloads all photos in parallel
func (d *Downloader) DownloadAll(photos []PhotoJob) {
	totalJobs := len(photos)
	if totalJobs == 0 {
		return
	}

	// Start workers
	d.startWorkers()

	// Send jobs to workers in a separate goroutine to avoid blocking
	go func() {
		for _, photo := range photos {
			d.jobs <- photo
		}
		close(d.jobs) // No more jobs to send
	}()

	// Process results directly in this function
	for i := 0; i < totalJobs; i++ {
		result := <-d.results
		d.resultsMutex.Lock()

		if result.Err != nil {
			logger.SubWarn(2, "Failed to download %s: %v", result.PhotoPath, result.Err)
			d.failedCount++
		} else if result.Skipped {
			logger.SubDetail(2, "Skipping, file already exists: %s", result.DestPath)
			d.skippedCount++
		} else {
			logger.SubDetail(2, "Downloaded %s", result.PhotoPath)
			d.successCount++
		}

		d.resultsMutex.Unlock()
	}

	// Wait for all workers to finish
	d.wg.Wait()
	close(d.results)
}
