package common

import (
	"encoding/csv"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"sort"
	"strconv"
	"time"
)

type Task struct {
	TaskID   int
	Jitter   int
	BCET     int
	WCET     int
	Period   int
	Deadline int
	PE       int
}

type TaskSet []*Task

func (t *Task) String() string {
	return "{ " + string(t.TaskID) + " " + string(t.Jitter) + " " + string(t.BCET) + " " + string(t.WCET) +
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

// WriteTaskSet function to write a task set to a CSV file
func (ts TaskSet) WriteTaskSet(path string) error {
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"TaskID", "Jitter", "BCET", "WCET", "Period", "Deadline", "PE"}
	writer.Write(headers)

	for i := range ts {

		row := []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%d", ts[i].Jitter),
			strconv.Itoa(ts[i].BCET),
			strconv.Itoa(ts[i].WCET),
			strconv.Itoa(ts[i].Period),
			strconv.Itoa(ts[i].Deadline),
			strconv.Itoa(ts[i].PE),
		}
		writer.Write(row)
	}

	return nil
}

// WriteTaskSetYAML function to write a task set to a YAML file
func (ts TaskSet) WriteTaskSetYAML(path string) error {
	// write the task set to a YAML file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// write the task set to a YAML file
	// first, we need to add taskset as the root element
	_, err = file.WriteString("taskset:\n")
	// then, we add the tasks
	for i, t := range ts {
		_, err = file.WriteString(fmt.Sprintf("  - TaskID: %d\n", i))
		_, err = file.WriteString(fmt.Sprintf("    Jitter: %d\n", t.Jitter))
		_, err = file.WriteString(fmt.Sprintf("    BCET: %d\n", t.BCET))
		_, err = file.WriteString(fmt.Sprintf("    WCET: %d\n", t.WCET))
		_, err = file.WriteString(fmt.Sprintf("    period: %d\n", t.Period))
		_, err = file.WriteString(fmt.Sprintf("    deadline: %d\n", t.Deadline))
		_, err = file.WriteString(fmt.Sprintf("    PE: %d\n", t.PE))

	}
	return nil
}

// ReadTaskSet function to read a task set from a CSV file
func ReadTaskSet(path string) (TaskSet, error) {
	// read the task set from the CSV file
	var tasks TaskSet
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// skip the header
	if _, err := reader.Read(); err != nil {
		panic(err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		tempID, _ := strconv.Atoi(record[0])
		tempJitter, _ := strconv.Atoi(record[1])
		tempBCET, _ := strconv.Atoi(record[2])
		tempWCET, _ := strconv.Atoi(record[3])
		tempPeriod, _ := strconv.Atoi(record[4])
		tempDeadline, _ := strconv.Atoi(record[5])
		tempPE, _ := strconv.Atoi(record[6])

		tasks = append(tasks, &Task{
			TaskID:   tempID,
			Jitter:   tempJitter,
			BCET:     tempBCET,
			WCET:     tempWCET,
			Period:   tempPeriod,
			Deadline: tempDeadline,
			PE:       tempPE,
		})
	}

	return tasks, nil
}

// ReadTaskSetYAML function to read a task set from a YAML file
func ReadTaskSetYAML(path string) (TaskSet, error) {
	// read the task set from the YAML file
	var tasks TaskSet
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// read the task set from the YAML file
	// first, unmarshal the YAML file
	var taskSet map[string][]map[string]interface{}
	err = yaml.Unmarshal(file, &taskSet)
	if err != nil {
		return nil, err
	}

	// then, we need to iterate over the task set
	for _, t := range taskSet["taskset"] {
		tempID := int(t["TaskID"].(int))
		tempJitter := int(t["Jitter"].(int))
		tempBCET := int(t["BCET"].(int))
		tempWCET := int(t["WCET"].(int))
		tempPeriod := int(t["period"].(int))
		tempDeadline := int(t["deadline"].(int))
		tempPE := int(t["PE"].(int))

		tasks = append(tasks, &Task{
			TaskID:   tempID,
			Jitter:   tempJitter,
			BCET:     tempBCET,
			WCET:     tempWCET,
			Period:   tempPeriod,
			Deadline: tempDeadline,
			PE:       tempPE,
		})
	}

	return tasks, nil
}
