package fsm

import (
	"reflect"
	"testing"
)

func TestDAGReachable(t *testing.T) {
	edges := map[int][]int{
		1: []int{2, 3, 5},
		2: []int{4, 7},
		3: []int{4, 7},
		4: []int{5, 6},
		7: []int{6},
	}
	r := reachableFrom(edges, 1)
	exp := []int{1, 2, 3, 4, 5, 6, 7}
	if !eqCheck(r, exp) {
		t.Errorf("mismatch exp: %v, got %+v", exp, r)
	}
	r = reachableFrom(edges, 3)
	exp = []int{3, 4, 5, 6, 7}
	if !eqCheck(r, exp) {
		t.Errorf("mismatch exp: %v, got %+v", exp, r)
	}
	r = reachableFrom(edges, 5)
	exp = []int{5}
	if !eqCheck(r, exp) {
		t.Errorf("mismatch exp: %v, got %+v", exp, r)
	}
}

// check without order
func eqCheck(m map[int]bool, exp []int) bool {
	expm := make(map[int]bool)
	for _, i := range exp {
		expm[i] = true
	}
	return reflect.DeepEqual(m, expm)
}
