package main

import myos "os"

func main() {
	// os импортирован под псевдонимом — вызов всё равно должен
	// распознаваться по информации о типах.
	myos.Exit(1) // want "прямой вызов os.Exit в функции main запрещён"
}
