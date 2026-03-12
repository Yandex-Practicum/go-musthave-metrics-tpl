package main

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	var users []User
	url := "https://jsonplaceholder.typicode.com/users"

	client := resty.New()

	_, err := client.R().
		SetResult(&users).
		Get(url)

	if err != nil {
		panic(err)
	}

	var out []string
	for _, v := range users {
		out = append(out, v.Username)
	}
	fmt.Println(strings.Join(out, ` `))
}
