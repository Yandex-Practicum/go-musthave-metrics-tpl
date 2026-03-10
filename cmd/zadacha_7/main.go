package main

import (
	"flag"
	"fmt"
)

func main() {

	destDir := flag.String("dest", "./output", "destination folder")
	width := flag.Int("w", 1024, "width of the image")
	isThumb := flag.Bool("thumb", false, "create thumb")

	flag.Parse()

	for i, v := range flag.Args() {
		fmt.Printf("Image file (%d): \r\n", i, v)
	}

	fmt.Println("Destination folder:", *destDir)
	fmt.Println("Width:", *width)
	fmt.Println("Thumbs:", *isThumb)
}
