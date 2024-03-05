package lib

import (
	"math"
	"math/rand"
)

// generatePeriodsLogUniform generates log-uniformly distributed periods to create tasks.
func generatePeriodsLogUniform(numTasks int, minPeriod, maxPeriod float64) []int {

	periods := make([]int, numTasks)
	for i := 0; i < numTasks; i++ {
		periods[i] = int(math.Round(math.Exp(rand.Float64()*(math.Log(maxPeriod)-math.Log(minPeriod)) + math.Log(minPeriod))))
	}

	return periods
}

// generatePeriodsLogUniformDiscrete generates log-uniformly distributed periods and
// rounds them down to the nearest predefined periods.
func generatePeriodsLogUniformDiscrete(numTasks int, minPeriod, maxPeriod float64, roundDownSet []int) []int {
	periodSet := generatePeriodsLogUniform(numTasks, minPeriod, maxPeriod)

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
