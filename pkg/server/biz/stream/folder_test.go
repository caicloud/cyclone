package stream

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/caicloud/nirvana/log"
)

func TestFolderReader(t *testing.T) {
	reader := NewFolderReader("/tmp/logs", "cyclone_", []string{"cyclone_c.log"}, time.Second*30)

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				log.Errorf("Read error: %v", err)
				return
			}

			fmt.Print(string(line))
		}
	}

}

func TestRegex(t *testing.T) {
	fmt.Println(InputContainerRegex.MatchString("i1"))
	fmt.Println(InputContainerRegex.MatchString("i11"))
	fmt.Println(InputContainerRegex.MatchString("i1a"))
	fmt.Println(InputContainerRegex.MatchString("ai1"))
	fmt.Println(OutputContainerRegex.MatchString("o1"))
	fmt.Println(OutputContainerRegex.MatchString("o2"))
}
