package main

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"sync"
)

const (
	fileURL    = "https://sample-img.lb-product.com/wp-content/themes/hitchcock/images/1GB.png"
	outputFile = "bigfile.png"
	chunkSize  = 1024 * 1024 // 1MB
)

func main() {
	err := DownloadFile(fileURL, outputFile)
	if err != nil {
		fmt.Printf("Error downloading file: %s\n", err)
	}
}

func DownloadFile(url, filename string) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var contentLength = resp.ContentLength
	if contentLength == -1 {
		return fmt.Errorf("server doesn't return file size")
	}
	numChunks := contentLength / chunkSize
	if contentLength%chunkSize != 0 {
		numChunks++
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var eg errgroup.Group
	var m sync.Mutex

	for i := int64(0); i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize - 1

		if end >= contentLength {
			end = contentLength - 1
		}

		eg.Go(func() error {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Read the body into a byte slice.
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			m.Lock()
			defer m.Unlock()

			_, err = file.WriteAt(body, start)
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
