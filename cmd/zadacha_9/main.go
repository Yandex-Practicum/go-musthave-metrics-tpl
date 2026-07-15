package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "0.0.1"

/*var Usage = func() {
	fmt.Fprintf(CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	PrintDefaults()
}*/

func main() {

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %v\nUsage of %s:\n", version, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}
