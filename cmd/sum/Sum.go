package Sum

func Summ(Values ...int) int {
	var sum int

	for _, v := range Values {
		sum += v
	}
	return sum
}
