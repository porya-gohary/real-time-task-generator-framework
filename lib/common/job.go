package common

import (
	"encoding/csv"
	"os"
	"strconv"
)

type Job struct {
	Task                *Task
	TaskID              int
	JobID               int
	EarliestArrivalTime int
	LatestArrivalTime   int
	Priority            int
	AbsoluteDeadline    int
}

type JobSet []*Job

// WriteJobSet writes a job set to a file
func (js JobSet) WriteJobSet(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	// write the job set to a file
	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Task ID", "Job ID", "Arrival min", "Arrival max", "Cost min", "Cost max", "Deadline", "Priority"}
	writer.Write(headers)

	for _, job := range js {
		row := []string{
			strconv.Itoa(job.TaskID),
			strconv.Itoa(job.JobID),
			strconv.Itoa(job.EarliestArrivalTime),
			strconv.Itoa(job.LatestArrivalTime),
			strconv.Itoa(job.Task.BCET),
			strconv.Itoa(job.Task.WCET),
			strconv.Itoa(job.AbsoluteDeadline),
			strconv.Itoa(job.Priority),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}
