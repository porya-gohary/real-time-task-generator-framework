package common

import (
	"sort"
	"time"
)

type Task struct {
	Name     string
	Jitter   int
	BCET     int
	WCET     int
	Period   int
	Deadline int
	PE       int
}

type TaskSet []*Task

func (t *Task) String() string {
	return "{ " + t.Name + " " + string(t.Jitter) + " " + string(t.BCET) + " " + string(t.WCET) +
		" " + string(t.Period) + " " + string(t.Deadline) + " " + string(t.PE) + " }"
}

// gcd calculates the greatest common divisor of two numbers
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// lcm calculates the least common multiple of two numbers
func lcm(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return abs(a*b) / gcd(a, b)
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// SortByPeriod function to sort tasks by period
func (ts *TaskSet) SortByPeriod() {
	// sort the tasks in ts by period
	sort.Slice(*ts, func(i, j int) bool {
		return (*ts)[i].Period < (*ts)[j].Period
	})
}

// HyperPeriod function to calculate the hyperperiod of a task set
func (ts TaskSet) HyperPeriod() int {
	// calculate the hyperperiod of the task set
	hyperperiod := 1
	start := time.Now()
	for _, t := range ts {
		if time.Since(start) < 60*time.Second {
			hyperperiod = lcm(hyperperiod, t.Period)
		} else {
			return -1
		}
	}
	return hyperperiod
}

// NumJobs function to calculate the number of jobs in the hyperperiod
func (ts TaskSet) NumJobs(hyperperiod int) int {
	// calculate the number of jobs in the hyperperiod
	numJobs := 0
	for _, t := range ts {
		numJobs += hyperperiod / t.Period
	}
	return numJobs
}
