package main

import "os"

func helper() {
	os.Exit(1) // вызов вне main — не считается нарушением
}

func main() {
	if len(os.Args) > 1 {
		os.Exit(1) // want "прямой вызов os.Exit в функции main запрещён"
	}
	helper()
	os.Exit(0) // want "прямой вызов os.Exit в функции main запрещён"
}
