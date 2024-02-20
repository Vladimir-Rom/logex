package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingBuffer(t *testing.T) {
	checkRB(t, []int{}, []int{})
	checkRB(t, []int{1}, []int{1})
	checkRB(t, []int{1, 2}, []int{1, 2})
	checkRB(t, []int{1, 2, 3}, []int{1, 2, 3})
	checkRB(t, []int{1, 2, 3, 4}, []int{2, 3, 4})
	checkRB(t, []int{1, 2, 3, 4, 5}, []int{3, 4, 5})
	checkRB(t, []int{1, 2, 3, 4, 5, 6}, []int{4, 5, 6})
	checkRB(t, []int{1, 2, 3, 4, 5, 6, 7}, []int{5, 6, 7})
}

func checkRB(t *testing.T, in []int, expected []int) {
	t.Helper()
	rb := newRingBuffer[int](3)
	for _, i := range in {
		rb.Add(i)
	}

	res := make([]int, 0, len(expected))
	for i := range rb.All {
		res = append(res, i)
	}

	assert.Equal(t, expected, res)
}
