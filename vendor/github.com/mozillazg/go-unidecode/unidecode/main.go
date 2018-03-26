package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/mozillazg/go-unidecode"
)

func main() {
	version := flag.Bool("V", false, "Output version info")
	flag.Parse()
	if *version {
		v := unidecode.Version()
		fmt.Printf("unidecode %s\n", v)
		os.Exit(0)
	}

	textSlice := flag.Args()
	stdin := []byte{}
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		stdin, _ = ioutil.ReadAll(os.Stdin)
	}
	if len(stdin) > 0 {
		textSlice = append(textSlice, string(stdin))
	}

	if len(textSlice) == 0 {
		fmt.Println("Usage: unidecode STRING")
		os.Exit(1)
	}

	s := strings.Join(textSlice, " ")
	ret := unidecode.Unidecode(s)
	fmt.Println(ret)
}
