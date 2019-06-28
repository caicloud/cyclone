package stream

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDir = "/tmp/cyclone-ut-data-logs"
)

func prepare() error {
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		err = os.MkdirAll(testDir, 0755)
		if err != nil {
			return err
		}
	}

	if err := writeFile(fmt.Sprintf("%s/cyclone_i1", testDir), "i1"); err != nil {
		return err
	}
	if err := writeFile(fmt.Sprintf("%s/cyclone_i2", testDir), "i2"); err != nil {
		return err
	}
	if err := writeFile(fmt.Sprintf("%s/cyclone_main", testDir), "main"); err != nil {
		return err
	}
	if err := writeFile(fmt.Sprintf("%s/cyclone_skip", testDir), "skip"); err != nil {
		return err
	}
	if err := writeFile(fmt.Sprintf("%s/cyclone_o1", testDir), "o1"); err != nil {
		return err
	}
	if err := writeFile(fmt.Sprintf("%s/cyclone_o2", testDir), "o2"); err != nil {
		return err
	}

	return nil
}

func clean() {
	os.RemoveAll(testDir)
}

func writeFile(path, content string) error {
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}
		file.WriteString(content)
		defer file.Close()
	}

	return nil
}

func TestFolderReader(t *testing.T) {
	if err := prepare(); err != nil {
		t.Errorf("Prepare for test error: %v", err)
		return
	}
	defer clean()

	reader := NewFolderReader(testDir, "cyclone_", []string{"cyclone_skip"}, 0)
	defer reader.Close()

	expecteds := []string{
		"i1",
		"i2",
		"main",
		"o1",
		"o2",
	}
	for _, expected := range expecteds {
		b, err := reader.ReadBytes('\n')
		assert.Nil(t, err)
		assert.Equal(t, expected, string(b))
	}
}

func TestContainerWeight(t *testing.T) {
	cases := []struct {
		name     string
		expected int
	}{
		{
			name:     "i1",
			expected: 200 - 1,
		},
		{
			name:     "i2",
			expected: 200 - 2,
		},
		{
			name:     "i2i",
			expected: 100,
		},
		{
			name:     "o2",
			expected: -200 - 2,
		},
		{
			name:     "o1",
			expected: -200 - 1,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, containerWeight(c.name))
	}
}
