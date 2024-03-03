package common

type Task struct {
	Name     string
	Jitter   float64
	BCET     float64
	WCET     float64
	Period   float64
	Deadline float64
	PE       int
}

type TaskSet []*Task
