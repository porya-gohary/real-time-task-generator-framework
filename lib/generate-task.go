package lib

import (
	"encoding/csv"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"task-generator/lib/common"
	"time"
)

var logger *common.VerboseLogger
var bar *progressbar.ProgressBar

// create a task set
func createTaskSet(path string, nTasks int, seed int64, totalUtilization float64, method string, alpha float64,
	jitter float64, isPreemptive bool, constantJitter bool, maxJobs int) error {
	rand.Seed(seed)

	var periods []int
	var wcets []int
	if method == "automotive" {
		tasks := generateAutomotiveTaskSet(totalUtilization)
		for _, task := range tasks {
			periods = append(periods, task[0])
			wcets = append(wcets, task[1])
		}
	}
	//else {
	//	var err error
	//	periods, wcets, err = generateLogUniformTaskSet(nTasks, totalUtilization, isPreemptive, maxJobs)
	//	if err != nil {
	//		return err
	//	}
	//}

	scale := 1
	bcets := make([]int, len(wcets))
	for i, wcet := range wcets {
		bcets[i] = int(alpha * float64(wcet) / float64(scale))
		wcets[i] = int(float64(wcet) / float64(scale))
		periods[i] = int(float64(periods[i]) / float64(scale))
	}

	// create the whole path
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Name", "Jitter", "BCET", "WCET", "Period", "Deadline", "PE"}
	writer.Write(headers)

	for i := range periods {
		period := periods[i]
		bcet := bcets[i]
		wcet := wcets[i]

		var maxJitter int
		if constantJitter {
			maxJitter = int(jitter)
		} else {
			maxJitter = int(jitter * float64(period))
		}

		if period-wcet < maxJitter {
			return fmt.Errorf("jitter is larger than Ti - Ci in file %s", path)
		}

		row := []string{
			fmt.Sprintf("T%d", i),
			fmt.Sprintf("%d", maxJitter),
			strconv.Itoa(bcet),
			strconv.Itoa(wcet),
			strconv.Itoa(period),
			strconv.Itoa(period),
			"1",
		}
		writer.Write(row)
	}
	return nil
}

// CreateTaskSets creates a number of task sets and writes them to the specified path
func CreateTaskSets(path string, numSets int, tasks int, utilization float64, periodDistribution string,
	execVariation float64, jitter float64, isPreemptive bool, constantJitter bool, maxJobs int, lr *common.VerboseLogger) {
	logger = lr
	if lr.GetVerboseLevel() == common.VerboseLevelNone {
		bar = progressbar.Default(int64(numSets))
	}
	for i := 0; i < numSets; i++ {
		file := fmt.Sprintf("%s_%d.csv", periodDistribution, i)
		taskSetPath := filepath.Join(path, file)
		if _, err := os.Stat(taskSetPath); os.IsNotExist(err) {
			if err := createTaskSet(taskSetPath, tasks, time.Now().UnixNano(), utilization, periodDistribution,
				execVariation, jitter, isPreemptive, constantJitter, maxJobs); err != nil {
				fmt.Println(err)
			} else {
				logger.LogInfo(fmt.Sprintf("%s created", taskSetPath))
			}
		} else {
			logger.LogInfo(fmt.Sprintf("%s exists", taskSetPath))
		}
		if lr.GetVerboseLevel() == common.VerboseLevelNone {
			bar.Add(1)
		}
	}
}

// CreateTaskSetsParallel creates task sets in parallel using the given parameters
func CreateTaskSetsParallel(path string, numSets int, tasks int, utilization float64, periodDistribution string,
	execVariation float64, jitter float64, isPreemptive bool, constantJitter bool, maxJobs int, lr *common.VerboseLogger) {
	var wg sync.WaitGroup
	wg.Add(numSets)
	logger = lr

	for i := 0; i < numSets; i++ {
		go func(setIndex int) {
			defer wg.Done()
			file := fmt.Sprintf("%s_%d.csv", periodDistribution, setIndex)
			taskSetPath := filepath.Join(path, file)
			if _, err := os.Stat(taskSetPath); os.IsNotExist(err) {
				if err := createTaskSet(taskSetPath, tasks, time.Now().UnixNano(), utilization, periodDistribution,
					execVariation, jitter, isPreemptive, constantJitter, maxJobs); err != nil {
					fmt.Println(err)
				} else {
					logger.LogInfo(fmt.Sprintf("%s created", taskSetPath))
				}
			} else {
				logger.LogInfo(fmt.Sprintf("%s exists", taskSetPath))
			}
		}(i)
	}

	wg.Wait()
}
