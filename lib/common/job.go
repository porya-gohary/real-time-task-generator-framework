package common

import (
	"encoding/csv"
	"os"
	"strconv"
)

type Job struct {
	Task                *Task
	Vertex              *Vertex
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
		}

		// we need to check if the task is a vertex
		if job.Vertex != nil {
			row = append(row, []string{
				strconv.Itoa(job.Vertex.BCET),
				strconv.Itoa(job.Vertex.WCET),
			}...)
		} else {
			row = append(row, []string{
				strconv.Itoa(job.Task.BCET),
				strconv.Itoa(job.Task.WCET),
			}...)
		}
		row = append(row, []string{
			strconv.Itoa(job.AbsoluteDeadline),
			strconv.Itoa(job.Priority),
		}...)

		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// WriteDependencyJobSet writes a job set dependency to a file
func (js JobSet) WriteDependencyJobSet(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	// write the dependency job set to a file
	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"From TID", "From JID", "To TID", "To JID"}
	writer.Write(headers)

	for _, job := range js {
		// find the successor
		// and keep their job ID
		var successorIndex []int

		for _, successor := range job.Vertex.Successors {
			for i, tempJob := range js {
				if tempJob.JobID == job.JobID {
					continue
				}
				// we need to check if they belong to the same task, and they release at the same time
				if tempJob.TaskID == successor && tempJob.AbsoluteDeadline == job.AbsoluteDeadline {
					successorIndex = append(successorIndex, i)
				}
			}
		}
		for _, successor := range successorIndex {
			row := []string{
				strconv.Itoa(job.TaskID),
				strconv.Itoa(job.JobID),
				strconv.Itoa(js[successor].TaskID),
				strconv.Itoa(js[successor].JobID),
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}
	return nil
}
