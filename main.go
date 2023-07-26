package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	root := "/papka/"

	checksums := make(map[string][]string)

	minSize := int64(1024)

	minAge := int64(60 * 60 * 24)

	numWorkers := 4

	startTime := time.Now()
	paths := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(paths, checksums, &wg)
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if info.Size() < minSize || time.Since(info.ModTime()) < time.Duration(minAge)*time.Second {
			return nil
		}

		paths <- path

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	close(paths)
	wg.Wait()

	endTime := time.Now()

	for _, filenames := range checksums {
		if len(filenames) > 1 {
			fmt.Printf("Duplicate files:\n")
			for _, filename := range filenames {
				fmt.Printf("%s\n", filename)
				err := os.Remove(filename)
				if err != nil {
					fmt.Printf("Failed to remove %s: %s\n", filename, err)
				} else {
					fmt.Printf("Removed %s\n", filename)
				}
			}
		}
	}

	fmt.Printf("Elapsed time: %v\n", endTime.Sub(startTime))
}

func worker(paths chan string, checksums map[string][]string, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range paths {
		checksum, err := fileSHA256(path)
		if err != nil {
			fmt.Println(err)
			continue
		}

		checksums[checksum] = append(checksums[checksum], path)
	}
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
