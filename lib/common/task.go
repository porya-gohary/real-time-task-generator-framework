package common

import "sort"

type Task struct {
	Name     string
	Jitter   int
	BCET     int
	WCET     int
	Period   int
	Deadline int
	PE       int
}

func (t *Task) String() string {
	return "{ " + t.Name + " " + string(t.Jitter) + " " + string(t.BCET) + " " + string(t.WCET) +
		" " + string(t.Period) + " " + string(t.Deadline) + " " + string(t.PE) + " }"
}

type TaskSet []*Task

// SortByPeriod function to sort tasks by period
func (ts *TaskSet) SortByPeriod() {
	// sort the tasks in ts by period
	sort.Slice(*ts, func(i, j int) bool {
		return (*ts)[i].Period < (*ts)[j].Period
	})
}
