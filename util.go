package dmmclust

// distro represents the sparse distribution of words. The key is the corpus ID, while the value is the frequency
type distro map[int]float64

// kvs is a simple linear associative array - it allocates far less memory and requires far less book keeping than distro
type kvs []struct {
	key int
	val float64
}

func (a kvs) has(key int) bool {
	for _, n := range a {
		if n.key == key {
			return true
		}
	}
	return false
}

func (a kvs) val(key int) float64 {
	for _, n := range a {
		if n.key == key {
			return n.val
		}
	}
	return 0.0
}

func (a kvs) add(key int) kvs {
	if a.has(key) {
		return a
	}
	return append(a, struct {
		key int
		val float64
	}{key: key, val: 0.0})
}

func (a kvs) incr(key int) {
	for i := range a {
		if a[i].key == key {
			a[i].val++
		}
	}
}

func sum(a []float64) (retVal float64) {
	for _, v := range a {
		retVal += v
	}
	return
}
