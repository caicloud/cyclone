package utils

import (
	"io/ioutil"
	"os"
)

// ListAllFiles lists all files from a directory and its sub-directory recursively.
func ListAllFiles(directory string) ([]string, error) {
	var results []string

	list, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	separator := string(os.PathSeparator)
	for _, f := range list {
		path := directory + separator + f.Name()
		if f.IsDir() {
			subFiles, err := ListAllFiles(path)
			if err != nil {
				return nil, err
			}
			results = append(results, subFiles...)
		} else {
			results = append(results, path)
		}
	}

	return results, nil
}
