package common

import (
	"encoding/csv"
	"fmt"
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

	headers := []string{"Task ID", "Job ID", "Arrival min", "Arrival max", "Cost min", "Cost max", "Deadline", "Priority", "Type"}
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
		if job.Vertex != nil {
			row = append(row, strconv.Itoa(job.Vertex.Type))
		}

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

// WriteJobSetYAML writes a job set to a YAML file
func (js JobSet) WriteJobSetYAML(path string) error {
	// write the job set to a YAML file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	// we need to add vertexset as the root element
	_, err = file.WriteString("jobset:\n")
	if err != nil {
		return err
	}

	// then, we add the jobs
	for _, job := range js {
		_, err = file.WriteString(fmt.Sprintf("  - TaskID: %d\n", job.TaskID))
		_, err = file.WriteString(fmt.Sprintf("    JobID: %d\n", job.JobID))
		_, err = file.WriteString(fmt.Sprintf("    Arrival min: %d\n", job.EarliestArrivalTime))
		_, err = file.WriteString(fmt.Sprintf("    Arrival max: %d\n", job.LatestArrivalTime))
		if job.Vertex != nil {
			_, err = file.WriteString(fmt.Sprintf("    Cost min: %d\n", job.Vertex.BCET))
			_, err = file.WriteString(fmt.Sprintf("    Cost max: %d\n", job.Vertex.WCET))
		} else {
			_, err = file.WriteString(fmt.Sprintf("    Cost min: %d\n", job.Task.BCET))
			_, err = file.WriteString(fmt.Sprintf("    Cost max: %d\n", job.Task.WCET))
		}
		_, err = file.WriteString(fmt.Sprintf("    Deadline: %d\n", job.AbsoluteDeadline))
		_, err = file.WriteString(fmt.Sprintf("    Priority: %d\n", job.Priority))
		if job.Vertex != nil {
			_, err = file.WriteString(fmt.Sprintf("    Type: %d\n", job.Vertex.Type))
		}

		if job.Vertex != nil {
			// now we need to check if the job has dependencies
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
			successors := "["
			for _, successor := range successorIndex {
				successors += "[" + strconv.Itoa(js[successor].TaskID) + "," + strconv.Itoa(js[successor].JobID) + "],"
			}
			if len(successors) > 1 {
				successors = successors[:len(successors)-1]
			}
			successors += "]"

			_, err = file.WriteString(fmt.Sprintf("    Successors: %s\n", successors))
		}

	}

	return nil

}
