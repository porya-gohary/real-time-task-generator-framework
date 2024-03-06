package lib

import (
	"encoding/csv"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"task-generator/lib/common"
	"time"
)

var logger *common.VerboseLogger
var bar *progressbar.ProgressBar

// create a task set
func createTaskSet(path string, numCore, nTasks int, seed int64, totalUtilization float64, utilDist string, periodDist string,
	periodRange []int, disPeriods []int, alpha float64, jitter float64, isPreemptive bool, constantJitter bool, maxJobs int) error {
	rand.Seed(seed)

	tasks := common.TaskSet{}
	var periods []int
	var util []float64
	for {
		// clear tasks
		periods = periods[:0]
		util = util[:0]
		tasks = tasks[:0]
		// First, we generate the utilization
		if utilDist == "uunifast" {
			// 1.UUnifastDiscard algorithm
			util = uunifastDiscard(nTasks, totalUtilization, 1.0)
		} else if utilDist == "rand-fixed-sum" {
			// 2. RandFixedSum algorithm
			util = StaffordRandFixedSum(nTasks, totalUtilization)
		} else if utilDist == "automotive" && periodDist == "automotive" {
			// 3. Automotive method
			tasks := generateAutomotiveTaskSet(totalUtilization)
			for _, task := range tasks {
				periods = append(periods, task[0])
				util = append(util, float64(task[1])/float64(task[0]))
			}
		} else {
			logger.LogFatal(fmt.Sprintf("Unknown utilization distribution: %s", utilDist))
		}

		// now we generate the periods
		if periodDist == "uniform" {
			// 1. Uniform distribution
			periods = generatePeriodsUniform(nTasks, float64(periodRange[0]), float64(periodRange[1]))
		} else if periodDist == "log-uniform" {
			// 2. Log-uniform distribution
			periods = generatePeriodsLogUniform(nTasks, float64(periodRange[0]), float64(periodRange[1]))
		} else if periodDist == "uniform-discrete" {
			// 3. Uniform discrete distribution
			periods = generatePeriodsUniformDiscrete(nTasks, float64(periodRange[0]), float64(periodRange[1]), disPeriods)
		} else if periodDist == "log-uniform-discrete" {
			// 4. Log-uniform discrete distribution
			periods = generatePeriodsLogUniformDiscrete(nTasks, float64(periodRange[0]), float64(periodRange[1]), disPeriods)
		} else if periodDist == "automotive" {
			// 5. Automotive method
			tasks := generateAutomotiveTaskSet(totalUtilization)
			for _, task := range tasks {
				periods = append(periods, task[0])
			}
		} else {
			logger.LogFatal(fmt.Sprintf("Unknown period distribution: %s", periodDist))
		}

		scale := 10
		for i, u := range util {
			wcet := int(float64(periods[i]) * u * float64(scale))
			bcet := int(alpha * float64(wcet))
			period := int(float64(periods[i]) * float64(scale))
			tasks = append(tasks, &common.Task{
				Period:   period,
				Deadline: period,
				WCET:     wcet,
				BCET:     bcet,
				PE:       0,
			})
		}

		// if WCET == 0, then we need to regenerate the task set
		flag := true
		for _, task := range tasks {
			if task.WCET == 0 {
				flag = false
				logger.LogInfo("Regenerating task set because of zero WCET")
				break
			}
			if constantJitter {
				task.Jitter = int(jitter)
			} else {
				task.Jitter = int(jitter * float64(task.Period))
			}
			if task.Period < task.WCET+task.Jitter {
				flag = false
				logger.LogInfo("Regenerating task set because of period less than WCET + Jitter")
				break
			}

		}
		// now we check the number of jobs in the hyperperiod
		if maxJobs > 0 {
			// If this took too long, we need to regenerate the task set
			// get the hyperperiod
			hyperperiod := tasks.HyperPeriod()
			if hyperperiod == -1 {
				// this means that the hyperperiod is too large
				flag = false
				logger.LogInfo("Regenerating task set because of large hyperperiod")
			} else {
				// get the number of jobs
				numJobs := tasks.NumJobs(hyperperiod)
				if numJobs > maxJobs {
					// this means that the number of jobs is too large
					flag = false
					logger.LogInfo("Regenerating task set because of large number of jobs")
				}
			}
		}
		if flag {
			break
		}

		if len(util) != nTasks || len(periods) != nTasks || len(tasks) != nTasks {
			logger.LogFatal("Error generating task set")
		}
	}

	// sort the tasks by period
	tasks.SortByPeriod()

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

	for i := range tasks {

		row := []string{
			fmt.Sprintf("T%d", i),
			fmt.Sprintf("%d", tasks[i].Jitter),
			strconv.Itoa(tasks[i].BCET),
			strconv.Itoa(tasks[i].WCET),
			strconv.Itoa(tasks[i].Period),
			strconv.Itoa(tasks[i].Deadline),
			strconv.Itoa(tasks[i].PE),
		}
		writer.Write(row)
	}
	return nil
}

// CreateTaskSets creates a number of task sets and writes them to the specified path
func CreateTaskSets(path string, numCore, numSets int, tasks int, utilization float64, utilDistribution string,
	periodDistribution string, periodRange []int, disPeriods []int, execVariation float64, jitter float64, isPreemptive bool,
	constantJitter bool, maxJobs int, lr *common.VerboseLogger) {
	// add spec to the path before output folder
	path = filepath.Join(path, fmt.Sprintf("%s-utilDist", utilDistribution))
	path = filepath.Join(path, fmt.Sprintf("%s-perDist", periodDistribution))
	path = filepath.Join(path, fmt.Sprintf("%d-core", numCore))
	path = filepath.Join(path, fmt.Sprintf("%d-task", tasks))
	if constantJitter {
		path = filepath.Join(path, fmt.Sprintf("%d-jitter", int(jitter)))
	} else {
		path = filepath.Join(path, fmt.Sprintf("%d-percent-jitter", int(jitter*100)))
	}
	path = filepath.Join(path, fmt.Sprintf("%.2f-util", utilization))
	path = filepath.Join(path, "tasksets")
	// sort disPeriods
	// from large to small
	sort.Slice(disPeriods, func(i, j int) bool {
		return disPeriods[i] > disPeriods[j]
	})
	logger = lr
	if lr.GetVerboseLevel() == common.VerboseLevelNone {
		bar = progressbar.Default(int64(numSets))
	}
	for i := 0; i < numSets; i++ {
		file := fmt.Sprintf("%s_%d.csv", periodDistribution, i)
		taskSetPath := filepath.Join(path, file)
		if _, err := os.Stat(taskSetPath); os.IsNotExist(err) {
			if err := createTaskSet(taskSetPath, numCore, tasks, time.Now().UnixNano(), utilization, utilDistribution,
				periodDistribution, periodRange, disPeriods, execVariation, jitter,
				isPreemptive, constantJitter, maxJobs); err != nil {
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
func CreateTaskSetsParallel(path string, numCore, numSets int, tasks int, utilization float64, utilDistribution string,
	periodDistribution string, periodRange []int, disPeriods []int, execVariation float64, jitter float64, isPreemptive bool,
	constantJitter bool, maxJobs int, lr *common.VerboseLogger) {
	// add spec to the path before output folder
	path = filepath.Join(path, fmt.Sprintf("%s-utilDist", utilDistribution))
	path = filepath.Join(path, fmt.Sprintf("%s-perDist", periodDistribution))
	path = filepath.Join(path, fmt.Sprintf("%d-core", numCore))
	path = filepath.Join(path, fmt.Sprintf("%d-task", tasks))
	if constantJitter {
		path = filepath.Join(path, fmt.Sprintf("%d-jitter", int(jitter)))
	} else {
		path = filepath.Join(path, fmt.Sprintf("%d-percent-jitter", int(jitter*100)))
	}
	path = filepath.Join(path, fmt.Sprintf("%.2f-util", utilization))
	path = filepath.Join(path, "tasksets")

	// sort disPeriods
	// from large to small
	sort.Slice(disPeriods, func(i, j int) bool {
		return disPeriods[i] > disPeriods[j]
	})
	var wg sync.WaitGroup
	wg.Add(numSets)
	logger = lr

	for i := 0; i < numSets; i++ {
		go func(setIndex int) {
			defer wg.Done()
			file := fmt.Sprintf("%s_%d.csv", periodDistribution, setIndex)
			taskSetPath := filepath.Join(path, file)
			if _, err := os.Stat(taskSetPath); os.IsNotExist(err) {
				if err := createTaskSet(taskSetPath, numCore, tasks, time.Now().UnixNano(), utilization, utilDistribution,
					periodDistribution, periodRange, disPeriods, execVariation, jitter, isPreemptive,
					constantJitter, maxJobs); err != nil {
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
