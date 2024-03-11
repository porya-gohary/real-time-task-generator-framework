package lib

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"task-generator/lib/common"
	"time"
)

func generateDAG(taskSet common.TaskSet, rootNodeNum, maxBranch, maxDepth int) common.VertexSet {
	rand.Seed(time.Now().UnixNano())

	// Check if there are enough tasks for DAG generation
	if len(taskSet) < rootNodeNum+maxDepth {
		logger.LogFatal("Small number of tasks for DAG!")
	}

	// make a vertex set and assign each task to a vertex
	var vertices common.VertexSet
	for _, task := range taskSet {
		taskID, _ := strconv.Atoi(task.Name[1:])

		vertices = append(vertices, &common.Vertex{
			TaskID:          taskID,
			VertexID:        taskID,
			RelativeRelease: task.Jitter,
			BCET:            task.BCET,
			WCET:            task.WCET,
		})
	}

	// first determine the depth of the DAG (between 2 and maxDepth)
	depth := rand.Intn(maxDepth-1) + 2

	// Classify tasks by randomly selecting depths
	levelArr := make([][]int, depth)
	for i := range levelArr {
		levelArr[i] = make([]int, 0)
	}

	// Put start nodes in level 0
	for i := 0; i < rootNodeNum; i++ {
		levelArr[0] = append(levelArr[0], i)
		vertices[i].Depth = 0
	}

	// Each level must have at least one node
	for i := 1; i < depth; i++ {
		levelArr[i] = append(levelArr[i], rootNodeNum+i-1)
		vertices[rootNodeNum+i-1].Depth = i
	}

	// Put other nodes in other levels randomly
	for i := rootNodeNum + depth - 1; i < len(taskSet); i++ {
		level := rand.Intn(depth-1) + 1
		vertices[i].Depth = level
		levelArr[level] = append(levelArr[level], i)
	}

	// Make edges
	for level := 0; level < depth-1; level++ {
		for _, taskIdx := range levelArr[level] {
			obNum := rand.Intn(maxBranch + 1)

			childIdxList := make([]int, 0)

			// If desired outbound edge number is larger than the number of next level nodes, select every node
			if obNum >= len(levelArr[level+1]) {
				childIdxList = append(childIdxList, levelArr[level+1]...)
			} else {
				for len(childIdxList) < obNum {
					childIdx := levelArr[level+1][rand.Intn(len(levelArr[level+1]))]
					if !contains(childIdxList, childIdx) {
						childIdxList = append(childIdxList, childIdx)
					}
				}
			}

			for _, childIdx := range childIdxList {
				vertices[taskIdx].Successors = append(vertices[taskIdx].Successors, childIdx)
				vertices[childIdx].Predecessors = append(vertices[childIdx].Predecessors, taskIdx)
			}
		}
	}
	return vertices
}

func generateRandomDAG(taskPath string, rootNodeNum, maxBranch, maxDepth int, makeDotFile bool) {
	// first we have to read the task set
	taskSet, err := common.ReadTaskSet(taskPath)
	if err != nil {
		logger.LogFatal("Error reading task set: " + err.Error())
	}

	// generate the DAG
	vertices := generateDAG(taskSet, rootNodeNum, maxBranch, maxDepth)

	// add ".prec" at the end of file before ".csv" and write the set of vertices to a file
	mainPath := taskPath[:len(taskPath)-4] + ".prec" + ".csv"
	// create the whole path
	err = os.MkdirAll(filepath.Dir(mainPath), os.ModePerm)
	file, err := os.Create(mainPath)
	if err != nil {
		logger.LogFatal("Error creating file: " + err.Error())
	}
	// write the job set to a file
	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Task ID", "Vertex ID", "Jitter", "BCET", "WCET", "Period", "Deadline", "Successors"})

	for i, task := range taskSet {
		// first we have to write the task
		lineTemp := []string{strconv.Itoa(vertices[i].TaskID), strconv.Itoa(vertices[i].VertexID),
			strconv.Itoa(vertices[i].RelativeRelease), strconv.Itoa(vertices[i].BCET), strconv.Itoa(vertices[i].WCET),
			strconv.Itoa(task.Period), strconv.Itoa(task.Deadline)}

		predStr := "["
		for _, pred := range vertices[i].Successors {
			predStr += strconv.Itoa(pred) + ","
		}
		if len(predStr) > 1 {
			predStr = predStr[:len(predStr)-1]
		}
		predStr += "]"
		lineTemp = append(lineTemp, predStr)
		writer.Write(lineTemp)
	}

	// write the dot file
	if makeDotFile {
		dotPath := taskPath[:len(taskPath)-4] + ".dot"
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

// GenerateRandomDAGs function to generate random DAGs
func GenerateRandomDAGs(taskSetPath string, rootNodeNum, maxBranch, maxDepth int, makeDotFile bool) {
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
		// make sure that the file does not exist
		predPath := taskSetPath[:len(taskSetPath)-4] + ".prec" + ".csv"
		if _, err := os.Stat(predPath); os.IsNotExist(err) {
			logger.LogInfo("Generating DAG for: " + taskSetPath)
			generateRandomDAG(taskSetPath, rootNodeNum, maxBranch, maxDepth, makeDotFile)
		} else {
			logger.LogInfo(fmt.Sprintf("%s exists", predPath))
		}
	}
}

// GenerateRandomDAGsParallel function to generate random DAGs in parallel
func GenerateRandomDAGsParallel(taskSetPath string, rootNodeNum, maxBranch, maxDepth int, makeDotFile bool) {
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

	// now we have to generate the DAGs
	var wg sync.WaitGroup
	wg.Add(len(taskSetPaths))
	for _, taskSetPath := range taskSetPaths {
		go func(taskSetPaths string) {
			defer wg.Done()
			// make sure that the file does not exist
			predPath := taskSetPath[:len(taskSetPath)-4] + ".prec" + ".csv"
			if _, err := os.Stat(predPath); os.IsNotExist(err) {
				logger.LogInfo("Generating DAG for: " + taskSetPath)
				go generateRandomDAG(taskSetPath, rootNodeNum, maxBranch, maxDepth, makeDotFile)
			} else {
				logger.LogInfo(fmt.Sprintf("%s exists", predPath))
			}
		}(taskSetPath)
	}
	wg.Wait()
}
