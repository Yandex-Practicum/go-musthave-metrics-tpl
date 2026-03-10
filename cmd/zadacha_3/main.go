package main

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// MyApiError — описание ошибки при неверном запросе.
type MyApiError struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Post — модель, описание основного объекта.
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

func main() {
	client := resty.New()

	resp, err := client.R().SetPathParams(map[string]string{
		"postID": "1",
	}).Get("https://jsonplaceholder.typicode.com/posts/{postID}")

	if err != nil {
		panic(err)
	}

	fmt.Println(resp)
}

/*client := resty.New()

	var responseErr MyApiError
	var post Post

	_, err := client.R().
		SetError(&responseErr).
		SetResult(&post).
		Get("https://jsonplaceholder.typicode.com/posts/1")

	if err != nil {
		fmt.Println(responseErr)
		panic(err)
		return
	}

	fmt.Println(post)
}*/
