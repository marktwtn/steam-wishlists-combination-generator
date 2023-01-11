package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_get_combination_count(t *testing.T) {
	test_data := []struct {
		selected int
		total    int
		expected int
	}{
		{3, 50, 19600},
		{50, 3, 0},
		{0, 50, 0},
		{3, 0, 0},
		{-3, 50, 0},
		{3, -50, 0},
		{-3, -50, 0},
	}
	for _, data := range test_data {
		var result = get_combination_count(data.selected, data.total)
		if result != data.expected {
			assert.Equalf(t, data.expected, result, "The result should be %d instead of %d.", result, data.expected)
		}
	}
}
