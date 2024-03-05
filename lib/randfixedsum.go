package lib

import (
	"math"
	"math/rand"
)

// StaffordRandFixedSum generates nsets sets of n numbers whose sum is approximately u.
func StaffordRandFixedSum(n int, u float64, nsets int) [][]float64 {
	// If n is less than u, return nil
	if n < int(u) {
		return nil
	}

	// Deal with n=1 case
	if n == 1 {
		result := make([][]float64, nsets)
		for i := 0; i < nsets; i++ {
			result[i] = []float64{u}
		}
		return result
	}

	k := math.Min(float64(int(u)), float64(n-1))
	s := u
	s1 := make([]float64, n)
	for i := 0; i < n; i++ {
		s1[i] = s - float64(int(k)) - float64(i)
	}
	s2 := make([]float64, n)
	for i := 0; i < n; i++ {
		s2[i] = float64(i) + float64(int(k)+1) - s
	}

	tiny := math.SmallestNonzeroFloat64
	huge := math.MaxFloat64

	w := make([][]float64, n)
	for i := range w {
		w[i] = make([]float64, n+1)
	}
	w[0][1] = huge
	t := make([][]float64, n-1)
	for i := range t {
		t[i] = make([]float64, n)
	}

	for i := 2; i <= n; i++ {
		tmp1 := make([]float64, i)
		for j := 0; j < i; j++ {
			tmp1[j] = w[i-2][j+1] * s1[j] / float64(i)
		}
		tmp2 := make([]float64, i)
		for j := 0; j < i; j++ {
			tmp2[j] = w[i-2][j] * s2[n-i+j] / float64(i)
		}
		for j := 1; j <= i; j++ {
			w[i-1][j] = tmp1[j-1] + tmp2[j-1]
		}
		tmp3 := make([]float64, i)
		tmp4 := make([]bool, i)
		for j := 0; j < i; j++ {
			tmp3[j] = w[i-1][j+1] + tiny
			tmp4[j] = s2[n-i+j] > s1[j]
		}
		for j := 0; j < i; j++ {
			if tmp4[j] {
				t[i-2][j] = (tmp2[j] / tmp3[j]) * 1.0
			} else {
				t[i-2][j] = (1 - tmp1[j]/tmp3[j]) * 1.0
			}
		}
	}

	x := make([][]float64, n)
	for i := range x {
		x[i] = make([]float64, nsets)
	}
	rt := make([][]float64, n-1)
	for i := range rt {
		rt[i] = make([]float64, nsets)
		for j := range rt[i] {
			rt[i][j] = rand.Float64()
		}
	}
	rs := make([][]float64, n-1)
	for i := range rs {
		rs[i] = make([]float64, nsets)
		for j := range rs[i] {
			rs[i][j] = rand.Float64()
		}
	}
	sArray := make([]float64, nsets)
	for i := range sArray {
		sArray[i] = s
	}
	jArray := make([]float64, nsets)
	for i := range jArray {
		jArray[i] = k + 1
	}
	sm := make([]float64, nsets)
	pr := make([]float64, nsets)
	for i := range pr {
		pr[i] = 1.0
	}

	for i := n - 1; i > 0; i-- {
		e := make([]bool, nsets)
		for j := range e {
			e[j] = rt[n-i-1][j] <= t[i-1][int(jArray[j])-1]
		}
		sx := make([]float64, nsets)
		for j := range sx {
			sx[j] = math.Pow(rs[n-i-1][j], 1/float64(i))
		}
		for j := range sm {
			sm[j] = sm[j] + (1-sx[j])*pr[j]*s/float64(i+1)
		}
		for j := range pr {
			pr[j] = sx[j] * pr[j]
		}
		for j := range x {
			x[n-i-1][j] = sm[j] + pr[j]*boolToFloat64(e[j])
		}
		for j := range sArray {
			sArray[j] = sArray[j] - boolToFloat64(e[j])
		}
		for j := range jArray {
			jArray[j] = jArray[j] - boolToFloat64(e[j])
		}
	}
	for j := range sm {
		x[n-1][j] = sm[j] + pr[j]*sArray[j]
	}

	for j := range x {
		x[j] = shuffleFloat64(x[j])
	}

	return x
}

// shuffleFloat64 shuffles a slice of float64 numbers.
func shuffleFloat64(slice []float64) []float64 {
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

// boolToFloat64 converts a boolean value to a float64 (1.0 for true, 0.0 for false).
func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
