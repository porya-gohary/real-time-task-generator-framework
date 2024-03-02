package lib

import (
	"math"
	"math/rand"
)

// Statistical distribution for task set generation from table 3
// of WATERS free benchmark paper.
var (
	acetByPeriod = map[int]float64{1000: 5, 2000: 4.20, 5000: 11.04, 10000: 10.09, 20000: 8.74, 50000: 17.56, 100000: 10.53, 200000: 2.56, 1000000: 0.43}
	wcetFmin     = map[int]float64{1000: 1.30, 2000: 1.54, 5000: 1.13, 10000: 1.06, 20000: 1.06, 50000: 1.13, 100000: 1.02, 200000: 1.03, 1000000: 1.84}
	wcetFmax     = map[int]float64{1000: 29.11, 2000: 19.04, 5000: 18.44, 10000: 30.03, 20000: 15.61, 50000: 7.76, 100000: 8.88, 200000: 4.90, 1000000: 4.75}
)

func generateAutomotiveWCET(period int) int {
	min := acetByPeriod[period] * wcetFmin[period]
	max := acetByPeriod[period] * wcetFmax[period]
	return int(min + rand.Float64()*(max-min))
}

func generateAutomotivePeriod() int {
	options := []int{1000, 2000, 5000, 10000, 20000, 50000, 100000, 200000, 1000000}
	weights := []float64{0.04, 0.02, 0.02, 0.29, 0.29, 0.04, 0.24, 0.01, 0.05}
	var totalWeight float64
	for _, weight := range weights {
		totalWeight += weight
	}
	for i, weight := range weights {
		weights[i] = weight / totalWeight
	}
	return options[weightedRandom(weights)]
}

func weightedRandom(weights []float64) int {
	r := rand.Float64()
	for i, weight := range weights {
		r -= weight
		if r <= 0 {
			return i
		}
	}
	return len(weights) - 1
}

func generateAutomotiveRunnable(targetUtilization float64) ([]int, []int) {
	var periods []int
	var wcets []int
	currentUtilization := 0.0
	for math.Abs(currentUtilization-targetUtilization) > 0.01 {
		period := generateAutomotivePeriod()
		wcet := generateAutomotiveWCET(period)
		if currentUtilization+float64(wcet)/float64(period) < targetUtilization {
			currentUtilization += float64(wcet) / float64(period)
			periods = append(periods, period)
			wcets = append(wcets, wcet)
		}
	}
	return periods, wcets
}

func generateAutomotiveTaskSet(targetUtilization float64) [][]int {
	periods, wcets := generateAutomotiveRunnable(targetUtilization)
	tasks := make([][]int, 0)
	t1 := periods[0]
	c1 := 0
	for i, period := range periods {
		if period != t1 {
			break
		}
		c1 += wcets[i]
	}
	ai := rand.Float64() * float64(2*(t1-c1))
	var currentTask []int
	for i, period := range periods {
		if period == t1 && currentTask != nil && float64(currentTask[1]+wcets[i]) <= ai {
			currentTask[1] += wcets[i]
		} else {
			if currentTask != nil {
				tasks = append(tasks, currentTask)
			}
			currentTask = []int{period, wcets[i]}
			ai = rand.Float64() * float64(2*(t1-c1))
		}
	}
	if currentTask != nil {
		tasks = append(tasks, currentTask)
	}
	return tasks
}
