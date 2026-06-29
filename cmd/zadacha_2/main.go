package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("https://practicum.yandex.ru")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	if _, err = io.CopyN(os.Stdout, resp.Body, 512); err != nil {
		fmt.Println(err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			fmt.Println(req.URL)
			return nil
		},
	}
	response, err := client.Get("http://ya.ru")
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	fmt.Println("Final URL:", response.Request.URL)
	fmt.Println("Status:", response.Status)
}
