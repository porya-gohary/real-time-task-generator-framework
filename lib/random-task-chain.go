package lib

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"task-generator/lib/common"
	"time"
)

func generateTaskChain(taskPath string, makeDotFile bool, outputFormat string) {
	rand.Seed(time.Now().UnixNano())

	// first we have to read the task set
	var taskSet common.TaskSet
	var err error

	// first we have to read the task set
	if outputFormat == "csv" {
		taskSet, err = common.ReadTaskSet(taskPath)
		if err != nil {
			logger.LogFatal("Error reading task set: " + err.Error())
		}
	} else {
		taskSet, err = common.ReadTaskSetYAML(taskPath)
		if err != nil {
			logger.LogFatal("Error reading task set: " + err.Error())
		}

	}

	// make a vertex set and assign each task to a vertex
	var vertices common.VertexSet
	for _, task := range taskSet {
		vertices = append(vertices, &common.Vertex{
			TaskID:   task.TaskID,
			VertexID: task.TaskID,
			Jitter:   task.Jitter,
			BCET:     task.BCET,
			WCET:     task.WCET,
		})
	}

	// shuffle the vertices
	rand.Shuffle(len(vertices), func(i, j int) { vertices[i], vertices[j] = vertices[j], vertices[i] })

	// Create a chain of tasks without any branches and cycles
	for i := 0; i < len(vertices)-1; i++ {
		vertices[i].Successors = append(vertices[i].Successors, vertices[i+1].VertexID)
	}

	// sort the vertices again based on the vertex ID
	vertices.Sort()

	// add ".prec" at the end of file before ".csv" and write the set of vertices to a file
	mainPath := taskPath[:strings.LastIndex(taskPath, ".")] + ".prec." + outputFormat
	// create the whole path
	err = os.MkdirAll(filepath.Dir(mainPath), os.ModePerm)
	file, err := os.Create(mainPath)
	if err != nil {
		logger.LogFatal("Error creating file: " + err.Error())
	}
	if outputFormat == "csv" {
		// write the job set to a file
		writer := csv.NewWriter(file)
		defer writer.Flush()

		writer.Write([]string{"Task ID", "Vertex ID", "Jitter", "BCET", "WCET", "Period", "Deadline", "Successors"})

		for i, task := range taskSet {
			// first we have to write the task
			lineTemp := []string{strconv.Itoa(vertices[i].TaskID), strconv.Itoa(vertices[i].VertexID),
				strconv.Itoa(vertices[i].Jitter), strconv.Itoa(vertices[i].BCET), strconv.Itoa(vertices[i].WCET),
				strconv.Itoa(task.Period), strconv.Itoa(task.Deadline)}

			succStr := "["
			for _, succ := range vertices[i].Successors {
				succStr += strconv.Itoa(succ) + ","
			}
			if len(succStr) > 1 {
				succStr = succStr[:len(succStr)-1]
			}
			succStr += "]"
			lineTemp = append(lineTemp, succStr)
			writer.Write(lineTemp)
		}
	} else {
		// we need to add vertexset as the root element
		_, err = file.WriteString("vertexset:\n")
		if err != nil {
			logger.LogFatal("Error writing to file: " + err.Error())
		}

		// then, we add the vertices
		for i, task := range taskSet {
			_, err = file.WriteString(fmt.Sprintf("  - TaskID: %d\n", vertices[i].TaskID))
			_, err = file.WriteString(fmt.Sprintf("    VertexID: %d\n", vertices[i].VertexID))
			_, err = file.WriteString(fmt.Sprintf("    Jitter: %d\n", vertices[i].Jitter))
			_, err = file.WriteString(fmt.Sprintf("    BCET: %d\n", vertices[i].BCET))
			_, err = file.WriteString(fmt.Sprintf("    WCET: %d\n", vertices[i].WCET))
			_, err = file.WriteString(fmt.Sprintf("    Period: %d\n", task.Period))
			_, err = file.WriteString(fmt.Sprintf("    Deadline: %d\n", task.Deadline))
			_, err = file.WriteString(fmt.Sprintf("    PE: %d\n", task.PE))

			succStr := "["
			for _, succ := range vertices[i].Successors {
				succStr += strconv.Itoa(succ) + ","
			}
			if len(succStr) > 1 {
				succStr = succStr[:len(succStr)-1]
			}
			succStr += "]"
			_, err = file.WriteString(fmt.Sprintf("    Successors: %s\n", succStr))

			if err != nil {
				logger.LogFatal("Error writing to file: " + err.Error())
			}
		}
	}

	// write the dot file
	if makeDotFile {
		dotPath := taskPath[:strings.LastIndex(taskPath, ".")] + ".dot"
		os.MkdirAll(filepath.Dir(dotPath), os.ModePerm)
		// write the dot file
		writerDot, err := os.Create(dotPath)
		if err != nil {
			logger.LogFatal("Error creating file: " + err.Error())
		}
		defer writerDot.Close()
		_, err = writerDot.WriteString("digraph G {\n" + vertices.GenerateDotFile("DAG", 0) + "\n}")
		if err != nil {
			logger.LogFatal("Error writing to file: " + err.Error())
		}

	}
}

// GenerateTaskChains generates task chains for a set of task sets
func GenerateTaskChains(taskSetPath string, makeDotFile bool, outputFormat string) {

	taskSetPaths := findTaskSetPaths(taskSetPath, outputFormat)
	// now we have to generate the task chain
	for _, taskSetPath := range taskSetPaths {
		// make sure that the file does not exist
		predPath := taskSetPath[:strings.LastIndex(taskSetPath, ".")] + ".prec." + outputFormat
		if _, err := os.Stat(predPath); os.IsNotExist(err) {
			logger.LogInfo("Generating task chain for: " + taskSetPath)
			generateTaskChain(taskSetPath, makeDotFile, outputFormat)
		} else {
			logger.LogInfo(fmt.Sprintf("%s exists", predPath))
		}
	}
}

// GenerateTaskChainsParallel generates task chains for a set of task sets in parallel
func GenerateTaskChainsParallel(taskSetPath string, makeDotFile bool, outputFormat string) {

	taskSetPaths := findTaskSetPaths(taskSetPath, outputFormat)
	// now we have to generate the task chain
	var wg sync.WaitGroup
	wg.Add(len(taskSetPaths))
	for _, taskSetPath := range taskSetPaths {
		go func(taskSetPath string) { // pass taskSetPath as an argument
			defer wg.Done()
			// make sure that the file does not exist
			predPath := taskSetPath[:strings.LastIndex(taskSetPath, ".")] + ".prec." + outputFormat
			if _, err := os.Stat(predPath); os.IsNotExist(err) {
				logger.LogInfo("Generating task chain for: " + taskSetPath)
				generateTaskChain(taskSetPath, makeDotFile, outputFormat)
			} else {
				logger.LogInfo(fmt.Sprintf("%s exists", predPath))
			}
		}(taskSetPath) // use taskSetPath here
	}
	wg.Wait() // wait for all goroutines to finish
}
