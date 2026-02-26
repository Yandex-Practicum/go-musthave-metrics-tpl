package Sum

import "testing"

func TestSumm(t *testing.T) {
	if sum := Summ(1, 2); sum != 3 {
		t.Errorf("sum expected to be 3; got %d", sum)
	}
}
