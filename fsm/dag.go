// Copyright 2018-2020 Celer Network

package fsm

// if next state is allowed from cur state as immediate next step. NO graph traversal!
// note if cur == next, result depends on allowed[cur] value
func isTransitionValid(allowed map[int][]int, cur, next int) bool {
	for _, s := range allowed[cur] {
		if s == next {
			return true
		}
	}
	return false
}

// traverse DAG depth first from start, return all states reachable including start
// uses recursion and NO cycle detection!
// could be further optimized by remember/pass visited state
// due to golang random map iteratation, return has no order guarantee
func reachableFrom(allowed map[int][]int, start int) map[int]bool {
	ret := map[int]bool{
		start: true,
	}
	for _, c := range allowed[start] {
		if ret[c] {
			continue
		}
		for k := range reachableFrom(allowed, c) {
			ret[k] = true
		}
	}
	return ret
}

func isReachable(allowed map[int][]int, old, new int) bool {
	return reachableFrom(allowed, old)[new]
}
