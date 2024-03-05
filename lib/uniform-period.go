package lib

import (
	"math/rand"
)

// generatePeriodsUniform generates uniformly distributed periods.
func generatePeriodsUniform(numTasks int, minPeriod, maxPeriod float64) []int {
	periods := make([]int, numTasks)
	for i := 0; i < numTasks; i++ {
		periods[i] = int(rand.Float64()*(maxPeriod-minPeriod) + minPeriod)
	}

	return periods
}

// generatePeriodsUniformDiscrete generates uniformly distributed periods and
// rounds them down to the nearest predefined periods.
func generatePeriodsUniformDiscrete(numTasks int, minPeriod, maxPeriod float64, roundDownSet []int) []int {
	periodSet := generatePeriodsUniform(numTasks, minPeriod, maxPeriod)

	roundedPeriodSets := make([]int, len(periodSet))
	for i, p := range periodSet {
		for _, r := range roundDownSet {
			if p >= r {
				roundedPeriodSets[i] = r
				break
			}
		}
	}

	return roundedPeriodSets
}
