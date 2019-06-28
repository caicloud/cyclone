package stream

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caicloud/nirvana/log"
)

var InputContainerRegex = regexp.MustCompile(`^i\d+$`)
var OutputContainerRegex = regexp.MustCompile(`^o\d+$`)

// FolderReader reads file content from a folder, it simulates single file reader and will
// read content from all files. Sub-folder will be ignored.
type FolderReader interface {
	ReadBytes(delim byte) ([]byte, error)
	io.ReadCloser
	Watch(interval time.Duration)
}

type folderReader struct {
	folder      string
	prefix      string
	exclusions  []string
	readers     []*fileReader
	filesMap    map[string]struct{}
	lock        *sync.Mutex
	watcherStop chan struct{}
}

type fileReader struct {
	name     string
	prefix   string
	filePath string
	file     io.ReadCloser
	reader   *bufio.Reader
}

type fileReadersSorter []*fileReader

// Len is the number of elements in the collection.
func (s fileReadersSorter) Len() int {
	return len(s)
}

// containerWeight gives a weight to each container. Three kinds of containers
// considered here:
// - input containers, with container names like 'i1', 'i2'
// - output containers, with container names like 'o1', 'o2'
// - workload containers
// Input containers should have largest weight, output containers should have the
// smallest weight. Considering there won't be too many containers in a pod, we
// can adjust the container index by fix weights 100, 200 to make the trick.
// If sort by the weight, result will be: i1, i2, ... main ... o1, o2, o3, ...
func containerWeight(c string) int {
	if InputContainerRegex.MatchString(c) {
		i, _ := strconv.Atoi(strings.TrimPrefix(c, "i"))
		return -i + 200
	} else if OutputContainerRegex.MatchString(c) {
		o, _ := strconv.Atoi(strings.TrimPrefix(c, "o"))
		return -o - 200
	}

	return 100
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s fileReadersSorter) Less(i, j int) bool {
	iWeight := containerWeight(strings.TrimPrefix(s[i].name, s[i].prefix))
	jWeight := containerWeight(strings.TrimPrefix(s[j].name, s[j].prefix))
	return iWeight > jWeight
}

// Swap swaps the elements with indexes i and j.
func (s fileReadersSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// NewFolderReader creates a folder content reader.
// - folder Path of the foler
// - prefix Only file names with given prefix would be read
// - exclusions Exclude given files
func NewFolderReader(folder string, prefix string, exclusions []string, duration time.Duration) FolderReader {
	folderRdr := &folderReader{
		folder:      folder,
		prefix:      prefix,
		exclusions:  exclusions,
		readers:     nil,
		filesMap:    make(map[string]struct{}),
		lock:        &sync.Mutex{},
		watcherStop: make(chan struct{}),
	}

	folderRdr.watch()
	if duration > 0 {
		go folderRdr.Watch(duration)
	}

	return folderRdr
}

// ReadBytes reads content from each file in the folder
func (r *folderReader) ReadBytes(delim byte) ([]byte, error) {
	for _, reader := range r.readers {
		line, err := reader.reader.ReadBytes(delim)

		// Some non io.EOF error, exit directly
		if err != nil && err != io.EOF {
			return line, err
		}

		// If end of current file, continue to next file
		if err == io.EOF {
			continue
		}

		// If no content read, continue to next file
		if len(line) <= 0 {
			continue
		}

		return line, err
	}

	// No file content found, all files have reach the end.
	return nil, io.EOF
}

// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered. Even if Read
// returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) bytes, Read conventionally
// returns what is available instead of waiting for more.
func (r *folderReader) Read(p []byte) (int, error) {
	for _, reader := range r.readers {
		n, err := reader.reader.Read(p)
		if n > 0 {
			if err == io.EOF {
				return n, nil
			} else {
				return n, err
			}
		}

		if err != nil {
			if err == io.EOF {
				continue
			}

			return 0, err
		}

		return n, err
	}

	return 0, io.EOF
}

// Close closes the folder reader, close all opened file to be precisely.
func (r *folderReader) Close() error {
	close(r.watcherStop)

	var errMsgs []string
	for _, reader := range r.readers {
		if reader.file != nil {
			err := reader.file.Close()
			if err != nil {
				errMsgs = append(errMsgs, err.Error())
			}
		}
	}

	if len(errMsgs) == 0 {
		return nil
	}

	return fmt.Errorf("%d file failed to close: %v", len(errMsgs), errMsgs)
}

// Watch watches files in the folder, when new file added, start to serve new file content.
func (r *folderReader) Watch(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			r.watch()
		case <-r.watcherStop:
			log.Infof("Stopped folder watcher: %s", r.folder)
			return
		}
	}
}

// watch watches files in the folder, when new file added, start to serve new file content.
func (r *folderReader) watch() {
	var newFiles []*fileReader
	filepath.Walk(r.folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if !strings.HasPrefix(info.Name(), r.prefix) {
			return nil
		}

		for _, exclusion := range r.exclusions {
			if exclusion == info.Name() {
				return nil
			}
		}

		if _, ok := r.filesMap[info.Name()]; ok {
			return nil
		}

		newFiles = append(newFiles, &fileReader{
			name:     info.Name(),
			filePath: path,
			prefix:   r.prefix,
		})
		return nil
	})

	if len(newFiles) == 0 {
		return
	}
	sort.Sort(fileReadersSorter(newFiles))

	r.lock.Lock()
	defer r.lock.Unlock()
	for _, reader := range newFiles {
		var rdr io.Reader
		file, err := os.Open(reader.filePath)
		if err != nil {
			log.Errorf("Open log file %s error: %v", reader.filePath, err)
			rdr = bytes.NewReader([]byte(fmt.Sprintf("Failed to open log file %s, error: %v", reader.filePath, err)))
		} else {
			rdr = file
		}
		reader.file = file
		reader.reader = bufio.NewReader(rdr)
		r.readers = append(r.readers, reader)
		r.filesMap[reader.name] = struct{}{}
	}
}
