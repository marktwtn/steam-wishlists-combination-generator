package main

import (
	"testing"
)

func Test_get_combination_count(t *testing.T) {
	var result = get_combination_count(3, 50)
	if result != 19600 {
		t.Error("The result should be 19600, but got", result)
	}
	result = get_combination_count(50, 3)
	if result != 0 {
		t.Error("The result should be 0, but got", result)
	}
	result = get_combination_count(0, 50)
	if result != 0 {
		t.Error("The result should be 0, but got", result)
	}
	result = get_combination_count(3, 0)
	if result != 0 {
		t.Error("The result should be 0, but got", result)
	}
	result = get_combination_count(-3, 50)
	if result != 0 {
		t.Error("The result should be 0, but got", result)
	}
	result = get_combination_count(3, -50)
	if result != 0 {
		t.Error("The result should be 0, but got", result)
	}
	result = get_combination_count(-3, -50)
	if result != 0 {
		t.Error("The result should be 0, but got", result)
	}
}
