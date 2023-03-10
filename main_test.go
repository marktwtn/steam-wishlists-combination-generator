package main

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/marktwtn/steam-wishlists-combination-generator/crawler"
)

func Test_is_budget_valid(t *testing.T) {
	test_data := []struct {
		lower_bound int
		upper_bound int
		expected    bool
	}{
		{0, 500, true},
		{500, 0, false},
		{500, 500, true},
		{-10, 500, false},
		{0, -500, false},
	}
	for _, data := range test_data {
		var result = is_budget_valid(data.lower_bound, data.upper_bound)
		if result != data.expected {
			assert.Equalf(t, data.expected, result, "The result should be %t instead of %t.", result, data.expected)
		}
	}
}

func Test_generate_filtered_combination(t *testing.T) {
	test_data := []crawler.Wishitem{}
	for idx := 0; idx < 10; idx++ {
		test_data = append(test_data, crawler.New_wishitem(uint(idx), strconv.Itoa(idx), 50, 80))
	}
	generate_filtered_combination(7, 1000, test_data, make(chan int, 100))
}
