package lib

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"task-generator/lib/common"
)

const (
	// RM is Rate Monotonic
	RM = 0
	// DM is Deadline Monotonic
	DM = 1
	// EDF is Earliest Deadline First
	EDF = 2
)

// generateJobSets generates jobs of each task set in one hyperperiod
func generateJobSet(taskPath string, priorityAssignment int) {
	// first let's see we have a prec file or not
	precPath := taskPath[:len(taskPath)-4] + ".prec.csv"
	if _, err := os.Stat(precPath); err == nil {
		// read the precedence graph
		precGraph, err := common.ReadVertexSet(precPath)
		if err != nil {
			logger.LogFatal("Error reading precedence graph: " + err.Error())
		}
		// print the number of vertices
		logger.LogInfo("Number of vertices in the precedence graph: " + strconv.Itoa(len(precGraph)))
		// print the vertices
		for i, vertex := range precGraph {
			logger.LogDebug("Vertex: " + strconv.Itoa(i) + " " + strconv.Itoa(vertex.VertexID) + " " + strconv.Itoa(vertex.Jitter) + " " + strconv.Itoa(vertex.BCET) + " " + strconv.Itoa(vertex.WCET) + " ")
		}

		// calculate the hyperperiod
		hyperperiod := precGraph.HyperPeriod()
		if hyperperiod == -1 {
			logger.LogFatal("Error calculating hyperperiod")
		}

		// now first let's create the job set
		jobSet := common.JobSet{}
		uniqueID := 0
		for _, vertex := range precGraph {
			// first we have to calculate the number of jobs
			numJobs := hyperperiod / vertex.Period
			for j := 0; j < numJobs; j++ {
				// now we have to calculate the arrival time
				earliestArrivalTime := j * vertex.Period
				latestArrivalTime := earliestArrivalTime + vertex.Jitter
				// now we have to calculate the deadline
				deadline := earliestArrivalTime + vertex.Deadline
				// now we have to calculate the priority
				priority := 0
				switch priorityAssignment {
				case RM:
					priority = vertex.Period
				case DM:
					priority = vertex.Deadline
				case EDF:
					priority = deadline
				}
				// now we have to create the job
				jobSet = append(jobSet, &common.Job{
					Vertex:              vertex,
					TaskID:              vertex.VertexID,
					JobID:               uniqueID,
					EarliestArrivalTime: earliestArrivalTime,
					LatestArrivalTime:   latestArrivalTime,
					AbsoluteDeadline:    deadline,
					Priority:            priority,
				})

				// print the job
				logger.LogDebug("Job: " + strconv.Itoa(jobSet[len(jobSet)-1].TaskID) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].JobID) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].EarliestArrivalTime) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].LatestArrivalTime) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].Vertex.BCET) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].Vertex.WCET) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].AbsoluteDeadline) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].Priority))

				uniqueID++
			}

		}
		// now we have to write the job sets to a file
		// remove folders before file name
		fileName := filepath.Base(taskPath)
		// get the taskPath without the file name
		mainPath := filepath.Dir(taskPath)
		// remove taskset from the taskPath
		mainPath = filepath.Dir(mainPath)
		// add jobsets folder to the taskPath
		mainPath = filepath.Join(mainPath, "jobsets")

		// add jobset before the file name
		mainPath = filepath.Join(mainPath, "jobset-"+fileName)

		// create the whole path
		err = os.MkdirAll(filepath.Dir(mainPath), os.ModePerm)
		err = jobSet.WriteJobSet(mainPath)
		if err != nil {
			logger.LogFatal("Error writing job set: " + err.Error())
		}

		// Now we have to create the precedence file
		// first add .prec before the file format
		precPath = mainPath[:len(mainPath)-4] + ".prec.csv"

		// create the whole path
		err = os.MkdirAll(filepath.Dir(precPath), os.ModePerm)
		err = jobSet.WriteDependencyJobSet(precPath)
		if err != nil {
			logger.LogFatal("Error writing precedence graph: " + err.Error())
		}

	} else {

		// first we have to read the task set
		tasks, err := common.ReadTaskSet(taskPath)
		if err != nil {
			logger.LogFatal("Error reading task set: " + err.Error())
		}
		// first we have to calculate the hyperperiod
		hyperperiod := tasks.HyperPeriod()
		if hyperperiod == -1 {
			logger.LogFatal("Error calculating hyperperiod")
		}
		// now we have to generate the job set
		jobSet := common.JobSet{}
		for i, task := range tasks {
			// first we have to calculate the number of jobs
			numJobs := hyperperiod / task.Period
			for j := 0; j < numJobs; j++ {
				// now we have to calculate the arrival time
				earliestArrivalTime := j * task.Period
				latestArrivalTime := earliestArrivalTime + task.Jitter
				// now we have to calculate the deadline
				deadline := earliestArrivalTime + task.Deadline
				// now we have to calculate the priority
				priority := 0
				switch priorityAssignment {
				case RM:
					priority = task.Period
				case DM:
					priority = task.Deadline
				case EDF:
					priority = deadline
				}
				// now we have to create the job
				jobSet = append(jobSet, &common.Job{
					Task:                task,
					TaskID:              i,
					JobID:               j,
					EarliestArrivalTime: earliestArrivalTime,
					LatestArrivalTime:   latestArrivalTime,
					AbsoluteDeadline:    deadline,
					Priority:            priority,
				})

				// print the job
				logger.LogDebug("Job: " + strconv.Itoa(jobSet[len(jobSet)-1].TaskID) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].JobID) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].EarliestArrivalTime) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].LatestArrivalTime) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].Task.BCET) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].Task.WCET) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].AbsoluteDeadline) + " " +
					strconv.Itoa(jobSet[len(jobSet)-1].Priority))
			}

		}
		// now we have to write the job sets to a file
		// remove folders before file name
		fileName := filepath.Base(taskPath)
		// get the taskPath without the file name
		mainPath := filepath.Dir(taskPath)
		// remove taskset from the taskPath
		mainPath = filepath.Dir(mainPath)
		// add jobsets folder to the taskPath
		mainPath = filepath.Join(mainPath, "jobsets")

		// add jobset before the file name
		mainPath = filepath.Join(mainPath, "jobset-"+fileName)

		// create the whole path
		err = os.MkdirAll(filepath.Dir(mainPath), os.ModePerm)
		err = jobSet.WriteJobSet(mainPath)
		if err != nil {
			logger.LogFatal("Error writing job set: " + err.Error())
		}
	}

}

// GenerateJobSets generates job sets for each task set in the task set folder
func GenerateJobSets(taskSetPath string, priorityAssignment int) {
	// first we have to find all the task sets with csv extension in
	var taskSetPaths []string
	err := filepath.Walk(taskSetPath, func(path string, info os.FileInfo, err error) error {
		// check folder name to be "tasksets"
		if filepath.Ext(path) == ".csv" && path[len(path)-9:] != ".prec.csv" &&
			filepath.Base(filepath.Dir(path)) == "tasksets" {
			taskSetPaths = append(taskSetPaths, path)
		}
		return nil
	})

	if err != nil {
		logger.LogFatal("Cannot find any task set in the folder: " + taskSetPath)
	} else {
		// print the number of task sets
		logger.LogInfo("Number of founded task sets: " + strconv.Itoa(len(taskSetPaths)))
	}

	if err != nil {
		logger.LogFatal("Cannot find any task set in the folder: " + taskSetPath)
	}

	// now we have to generate the job sets
	for _, taskSetPath := range taskSetPaths {
		// make sure that the task set is not generated before
		// remove folders before file name
		fileName := filepath.Base(taskSetPath)
		// get the taskPath without the file name
		mainPath := filepath.Dir(taskSetPath)
		// remove taskset from the taskPath
		mainPath = filepath.Dir(mainPath)
		// add jobsets folder to the taskPath
		mainPath = filepath.Join(mainPath, "jobsets")
		// add jobset before the file name
		mainPath = filepath.Join(mainPath, "jobset-"+fileName)
		if _, err := os.Stat(mainPath); os.IsNotExist(err) {
			logger.LogInfo("Generating job set for: " + taskSetPath)
			generateJobSet(taskSetPath, priorityAssignment)
		} else {
			logger.LogInfo("Job set for " + taskSetPath + " exists")
		}
	}
}

// GenerateJobSetsParallel generates job sets for each task set in the task set folder in parallel
func GenerateJobSetsParallel(taskSetPath string, priorityAssignment int) {
	// first we have to find all the task sets with csv extension in
	var taskSetPaths []string
	err := filepath.Walk(taskSetPath, func(path string, info os.FileInfo, err error) error {
		// check folder name to be "tasksets"
		if filepath.Ext(path) == ".csv" && path[len(path)-9:] != ".prec.csv" &&
			filepath.Base(filepath.Dir(path)) == "tasksets" {
			taskSetPaths = append(taskSetPaths, path)
		}
		return nil
	})

	if err != nil {
		logger.LogFatal("Cannot find any task set in the folder: " + taskSetPath)
	} else {
		// print the number of task sets
		logger.LogInfo("Number of founded task sets: " + strconv.Itoa(len(taskSetPaths)))
	}

	// now we have to generate the job sets in parallel
	var wg sync.WaitGroup
	wg.Add(len(taskSetPaths))
	for i := 0; i < len(taskSetPaths); i++ {
		go func(setIndex int) {
			defer wg.Done()
			// make sure that the task set is not generated before
			// remove folders before file name
			fileName := filepath.Base(taskSetPaths[setIndex])
			// get the taskPath without the file name
			mainPath := filepath.Dir(taskSetPaths[setIndex])
			// remove taskset from the taskPath
			mainPath = filepath.Dir(mainPath)
			// add jobsets folder to the taskPath
			mainPath = filepath.Join(mainPath, "jobsets")
			// add jobset before the file name
			mainPath = filepath.Join(mainPath, "jobset-"+fileName)
			if _, err := os.Stat(mainPath); os.IsNotExist(err) {
				generateJobSet(taskSetPaths[setIndex], priorityAssignment)
			} else {
				logger.LogInfo("Job set for " + taskSetPaths[setIndex] + " exists")
			}
		}(i)
	}
	wg.Wait()
}
