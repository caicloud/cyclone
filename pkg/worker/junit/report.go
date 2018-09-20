package junit

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/caicloud/nirvana/log"
)

const (
	xmlExt                = ".xml"
	testsuiteElementStart = "<testsuite"
	testsuiteElementEnd   = "testsuite>"
)

type ReportInterface interface {
	FindReportFiles() []string
}

type Report struct {
	defaultBasePaths []string

	rootPath string
}

func NewReport(base string) ReportInterface {
	report := &Report{
		defaultBasePaths: []string{
			path.Join(base, "target/surefire-reports"), // maven
			path.Join(base, "build/test-results"),      // gradle
			path.Join(base, "build/test-results/test"), // gradle
		},
		rootPath: base,
	}

	return report

}

// FindReportFiles finds junit xml files in default base paths firstly, if no results have bend found, then find in root file.
func (r *Report) FindReportFiles() []string {
	paths := []string{}

	for _, p := range r.defaultBasePaths {
		fs := findAllJUnitXmlFiles(p)
		paths = append(paths, fs...)
	}

	if len(paths) > 0 {
		return paths
	}

	return findAllJUnitXmlFiles(r.rootPath)
}

func findAllJUnitXmlFiles(root string) []string {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}
	log.Error("--- root:", root)
	var files []string
	filepath.Walk(root, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			// JUnit xml file end with '.xml'
			if filepath.Ext(path) != xmlExt {
				return nil
			}

			// Junit xml contain 'testsuite' element
			if !containsElement(path) {
				return nil
			}

			files = append(files, path)
		}
		return nil
	})
	return files
}

// containsElement judge whether the path file contains JUnit xml element, e.g. 'testsuite'
func containsElement(path string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("read xsd file  error: %v", err)
		return false
	}

	s := string(data)
	return strings.Contains(s, testsuiteElementStart) && strings.Contains(s, testsuiteElementEnd)
}
