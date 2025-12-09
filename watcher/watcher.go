package watcher

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

// Watcher: Returns the updates for a particular file
// 			Watches for changes too


const (
	CHUNKSIZE = 4096
)

// Returns the number of lines requested
func GetLastNLines(filepath string, n int) ([]string, int64, error) {
	// open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return []string{}, 0, nil
	}

	offset := findOffsetForLastNLines(file, fileSize, n)

	_, err = file.Seek(offset, 0)
	if err != nil {
		return nil, 0, err
	}
	var lines []string

	// TODO: linear scan very unoptimised (fail for GB)
	// probably do a seek to the EOF and read from the last?
	scanner := bufio.NewScanner(file)
    for scanner.Scan() {
		lines = append(lines, scanner.Text())
    }
	return lines, int64(len(lines)), nil
}

func Watch(filepath string, initialOffset int64, updatesChan chan<- string) {
	// watcher
	watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

	// Add file to watcher
	err = watcher.Add(filepath)
	if err != nil {
		log.Printf("ERROR: Could not add file %s: %v", filepath, err)
		return
	}

	lastOffset := initialOffset

	// Event Loop
	go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }

				// Write Modify
				if event.Has(fsnotify.Write) {
                    lastOffset = readNewLines(filepath, lastOffset, updatesChan)

                }
                
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Printf("watcher error: %v\n", err)
            }
        }
    }()
}

// reads from the last known offset
func readNewLines(filepath string, lastOffset int64, updatesChan chan<- string) int64 {
	// open the file
	file, err := os.Open(filepath)
	if err != nil {
		return lastOffset
	}
	defer file.Close()

	_, err = file.Seek(lastOffset, 0)
	if err != nil {
		return lastOffset
	}

	reader := bufio.NewReader(file)
	newOffset := lastOffset

	for {
		line, err := reader.ReadString('\n')

		if len(line) > 0 {
			updatesChan <- line
			newOffset += int64(len(line))
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
	}
	return newOffset
}

// this func should crawl backwards to find where the last N line begins
func findOffsetForLastNLines(file *os.File, filesize int64, n int) int64 {
	if n == 0 {
		return filesize
	}

	var newlinesfound int
	var offset int64 = filesize
	buf := make([]byte, CHUNKSIZE)

	for offset > 0 {
		readSize := int64(CHUNKSIZE)
		if offset < CHUNKSIZE {
			readSize = offset
		}
		
		offset -= readSize
		_, err := file.Seek(offset, 0)
		if err != nil {
			return 0
		}

		_, err = file.Read(buf[:readSize])
		if err != nil {
			return 0
		}

		for i := int(readSize) - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				newlinesfound++
			}

			if newlinesfound > n {
				return offset + int64(i) + 1
			}
		}
	}

	return 0
}