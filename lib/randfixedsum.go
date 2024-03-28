package lib

import (
	"math"
	"math/rand"
)

// StaffordRandFixedSum generates an n by m array x, each of whose m columns
// contains n random values lying in the interval [a,b], but
// subject to the condition that their sum be equal to s.
func StaffordRandFixedSum(n int, u, a, b float64) []float64 {

	// Deal with n=1 case
	if n == 1 {
		return []float64{u}
	}

	// Rescale to a unit cube: 0 <= x(i) <= 1
	u = (u - float64(n)*a) / (b - a)

	k := int(math.Min(u, float64(n-1)))
	s := u
	s1 := s - float64(k)
	s2 := float64(k+n) - s

	tiny := math.SmallestNonzeroFloat64
	huge := math.MaxFloat64

	w := make([][]float64, n)
	for i := range w {
		w[i] = make([]float64, n+1)
		// initialize w to 0
		for j := range w[i] {
			w[i][j] = 0.0
		}
	}
	w[0][1] = huge

	t := make([][]float64, n)
	for i := range t {
		t[i] = make([]float64, n)
		// initialize t to 0
		for j := range t[i] {
			t[i][j] = 0.0
		}
	}

	for i := 2; i <= n; i++ {
		for j := 1; j <= i; j++ {
			tmp1 := w[i-2][j] * s1 / float64(i)
			tmp2 := w[i-2][j-1] * s2 / float64(i)
			w[i-1][j] = tmp1 + tmp2
			tmp3 := w[i-1][j] + tiny
			tmp4 := 0.0
			if s2 > s1 {
				tmp4 = 1.0
			}
			t[i-2][j-1] = (tmp2/tmp3)*tmp4 + (1.0-tmp1/tmp3)*(1.0-tmp4)
		}
	}

	x := make([]float64, n)
	// initialize x to 0
	for i := range x {
		x[i] = 0.0
	}
	rt := rand.Float64() // rand simplex type
	rs := rand.Float64() // rand position in simplex
	s = u
	j := k + 1
	sm := 0.0
	pr := 1.0

	for i := n - 1; i >= 0; i-- {
		e := rt <= t[i][j-1]               // decide which direction to move in this dimension (1 or 0)
		sx := math.Pow(rs, 1.0/float64(i)) // next simplex coord
		sm += (1.0 - sx) * pr * s / float64(i+1)
		pr *= sx
		x[n-i-1] = sm + pr*boolToFloat64(e)
		s -= boolToFloat64(e)
		j -= int(boolToFloat64(e)) // change transition table column if required
	}

	x[n-1] = sm + pr*s

	// rescale from the unit cube to the desired interval
	for i := range x {
		x[i] = a + x[i]*(b-a)
	}

	return x
}

// boolToFloat64 converts a boolean value to a float64 (1.0 for true, 0.0 for false).
func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
