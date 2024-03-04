package lib

import (
	"math/rand"
)

// uunifastDiscard generates utilization values using the UUniFast algorithm and discard
// tasks that exceed the utilization limit.
func uunifastDiscard(numTask int, utilization float64, taskUtilizationLimit float64) []float64 {
	utilizationValues := make([]float64, numTask)
	// an infinite loop that will break when the utilization values are within the limit
	for {
		utilizationValues = make([]float64, numTask)
		sumU := utilization
		for i := 0; i < numTask-1; i++ {
			nextSumU := sumU * rand.Float64() * (1.0 / float64(numTask-i))
			utilizationValues[i] = sumU - nextSumU
			sumU = nextSumU
		}
		utilizationValues[numTask-1] = sumU
		// check if the utilization values are within the limit
		flag := true
		for _, u := range utilizationValues {
			if u > taskUtilizationLimit {
				flag = false
				break
			}
		}
		if flag {
			break
		}
	}
	return utilizationValues

}
