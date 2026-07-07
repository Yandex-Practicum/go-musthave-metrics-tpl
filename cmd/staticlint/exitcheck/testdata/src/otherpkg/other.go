package otherpkg

import "os"

// В пакете, отличном от main, прямой вызов os.Exit допустим и не должен
// вызывать диагностику анализатора.
func Quit() {
	os.Exit(1)
}
