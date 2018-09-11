package main

import "testing"

func TestKrand(t *testing.T) {

	s, k := 3, 6
	ret := krand(s, k)
	expected := 3

	if len(ret) != expected {
		t.Errorf("Except the %d to be %d , but instead got %d", s, expected, len(ret))
	}

}
