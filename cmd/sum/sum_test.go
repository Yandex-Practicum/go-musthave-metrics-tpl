// sum_test.go
package sum

import "testing"

func TestSum(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   int
	}{
		{
			name:   "simple tets #1",
			values: []int{1, 2, 3},
			want:   6,
		},
		{
			name:   "one",
			values: []int{1},
			want:   1,
		},
		{
			name:   "with negative values",
			values: []int{-1, -2, -3},
			want:   -6,
		},
		{
			name:   "with negative zero",
			values: []int{-0, 3},
			want:   3,
		},
		{
			name:   "a lot of values",
			values: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 18},
			want:   189,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if sum := Sum(test.values...); sum != test.want {
				t.Errorf("Sum() = %d, want %d", sum, test.want)
			}
		})

	}
}
