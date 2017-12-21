package bots

// IntSet is an int64 set type.
type IntSet map[int64]struct{}

var v = struct{}{}

// Add add an int64 number to the set. Return false if the number already exists in the set.
func (set IntSet) Add(i int64) bool {
	if _, ok := set[i]; ok {
		return false
	}
	set[i] = v
	return true
}

// AddAll add a slice of ints to the set.
func (set IntSet) AddAll(xs []int64) {
	for _, i := range xs {
		set[i] = v
	}
}

// Max returns the max number in the set. O(n).
func (set IntSet) Max() int64 {
	var ret *int64
	for i := range set {
		if ret == nil || i > *ret {
			ret = &i
		}
	}
	return *ret
}

// Min returns the min number in the set. O(n).
func (set IntSet) Min() int64 {
	var ret *int64
	for i := range set {
		if ret == nil || i < *ret {
			ret = &i
		}
	}
	return *ret
}
