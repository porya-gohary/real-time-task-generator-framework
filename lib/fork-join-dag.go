package lib

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"task-generator/lib/common"
)

//	Partly adapted from D. Casini, A. Biondi, G. Nelissen, and G. Buttazzo, "Partitioned Fixed-Priority Scheduling of
//	Parallel Tasks Without Preemptions", (RTSS 2018), 2018.
//	https://retis.sssup.it/~d.casini/resources/DAG_Generator/cptasks.zip

func expandDAG(vertices common.VertexSet, source, sink, depth, numBranches, maxParBranches,
	maxVertices int, pPar float64) common.VertexSet {
	parBranches := rand.Intn(maxParBranches-1) + 2

	if source == 0 && sink == 0 {
		// add the source and sink vertices
		so := &common.Vertex{VertexID: 0, Depth: depth}
		si := &common.Vertex{VertexID: 1, Depth: -depth}

		vertices = append(vertices, so, si)

		vertices = expandDAG(vertices, 0, 1, depth-1, parBranches, maxParBranches, maxVertices, pPar)
	} else {
		for i := 0; i < numBranches; i++ {
			current := len(vertices)
			vertices = append(vertices, &common.Vertex{VertexID: current})

			r := rand.Float64()
			isParallelNode := depth > 0 && r < pPar && len(vertices) < maxVertices

			if !isParallelNode {
				vertices[current].Predecessors = []int{source}
				vertices[current].Successors = []int{sink}
				vertices[current].Depth = depth

				vertices[source].Successors = append(vertices[source].Successors, current)
				vertices[sink].Predecessors = append(vertices[sink].Predecessors, current)
			} else {
				vertices = append(vertices, &common.Vertex{VertexID: current + 1})

				vertices[current].Predecessors = []int{source}
				vertices[source].Successors = append(vertices[source].Successors, current)

				vertices[current+1].Successors = []int{sink}
				vertices[sink].Predecessors = append(vertices[sink].Predecessors, current+1)

				vertices[current].Depth = depth
				vertices[current+1].Depth = -depth

				vertices = expandDAG(vertices, current, current+1, depth-1, parBranches, maxParBranches, maxVertices, pPar)
			}
		}
	}
	return vertices
}

func addRandomEdgesToDAG(vertices common.VertexSet, pAdd float64) common.VertexSet {
	for i := range vertices {
		for j := range vertices {
			r := rand.Float64()

			if vertices[i].Depth > vertices[j].Depth && !contains(vertices[i].Successors, j) && r < pAdd {
				vertices[i].Successors = append(vertices[i].Successors, j)
				vertices[j].Predecessors = append(vertices[j].Predecessors, i)
			}
		}
	}
	return vertices
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// Generates n random integers that sum to s
func generateRandomSum(n, s int) []int {
	// n random floats
	randN := make([]float64, n)
	for i := range randN {
		randN[i] = rand.Float64()
	}

	// extend the floats so the sum is approximately x (might be up to 3 less, because of flooring)
	result := make([]int, n)
	sumRandN := 0.0
	for _, v := range randN {
		sumRandN += v
	}

	for i := range randN {
		result[i] = int(math.Floor(randN[i] * float64(s) / sumRandN))
	}

	for i := 0; i < s-sum(result); i++ {
		idx := rand.Intn(n)
		result[idx]++
	}

	return result
}

func sum(slice []int) int {
	sum := 0
	for _, v := range slice {
		sum += v
	}
	return sum
}

func generateBCET(totalBCET, totalWCET int, wcetList []int) []int {
	bcetList := make([]int, len(wcetList))
	for i := range wcetList {
		bcetList[i] = int(math.Round(float64(wcetList[i]) * float64(totalBCET) / float64(totalWCET)))
	}

	for sum(bcetList) < totalBCET {
		idx := rand.Intn(len(bcetList))

		if bcetList[idx] < wcetList[idx] {
			bcetList[idx]++
		}
	}

	return bcetList
}
func generateDAGFromTask(task common.Task, pPar, pAdd float64, maxParBranches, maxVertices, maxDepth int) common.VertexSet {

	vertices := common.VertexSet{}
	vertices = expandDAG(vertices, 0, 0, maxDepth, 1, maxParBranches, maxVertices, pPar)
	vertices = addRandomEdgesToDAG(vertices, pAdd)
	wcetList := generateRandomSum(len(vertices), task.WCET)
	bcetList := generateBCET(task.BCET, task.WCET, wcetList)

	for i := range vertices {
		vertices[i].TaskID = task.TaskID
		vertices[i].Jitter = task.Jitter
		vertices[i].BCET = bcetList[i]
		vertices[i].WCET = wcetList[i]
	}
	return vertices
}

func generateDAGSet(taskPath string, pPar, pAdd float64, maxParBranches, maxVertices, maxDepth int,
	makeDotFile bool, outputFormat string) {
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

	// add ".prec" at the end of file before its format and write the set of vertices to a file
	mainPath := taskPath[:strings.LastIndex(taskPath, ".")] + ".prec." + outputFormat

	// create the whole path
	err = os.MkdirAll(filepath.Dir(mainPath), os.ModePerm)
	file, err := os.Create(mainPath)
	if err != nil {
		logger.LogFatal("Error creating file: " + err.Error())
	}

	dotFile := ""

	if outputFormat == "csv" {
		// write the job set to a file
		writer := csv.NewWriter(file)
		writer.Write([]string{"Task ID", "Vertex ID", "Jitter", "BCET", "WCET", "Period", "Deadline", "Successors"})

		defer writer.Flush()

		vertexIDCounter := 0
		for _, task := range taskSet {
			newDAG := generateDAGFromTask(*task, pPar, pAdd, maxParBranches, maxVertices, maxDepth)
			// first we have to write the task
			if makeDotFile {
				dotFile += newDAG.GenerateDotFile("T"+string(task.TaskID), vertexIDCounter)
			}
			for _, vertex := range newDAG {
				lineTemp := []string{strconv.Itoa(vertex.TaskID), strconv.Itoa(vertex.VertexID + vertexIDCounter),
					strconv.Itoa(vertex.Jitter), strconv.Itoa(vertex.BCET), strconv.Itoa(vertex.WCET),
					strconv.Itoa(task.Period), strconv.Itoa(task.Deadline)}

				succStr := "["
				for _, succ := range vertex.Successors {
					succStr += strconv.Itoa(succ+vertexIDCounter) + ","
				}
				if len(succStr) > 1 {
					succStr = succStr[:len(succStr)-1]
				}
				succStr += "]"
				lineTemp = append(lineTemp, succStr)
				if err := writer.Write(lineTemp); err != nil {
					logger.LogFatal("Error writing to file: " + err.Error())
				}

			}
			vertexIDCounter += len(newDAG)
		}
	} else {
		// we need to add vertexset as the root element
		_, err = file.WriteString("vertexset:\n")
		if err != nil {
			logger.LogFatal("Error writing to file: " + err.Error())
		}
		vertexIDCounter := 0
		for _, task := range taskSet {
			newDAG := generateDAGFromTask(*task, pPar, pAdd, maxParBranches, maxVertices, maxDepth)
			// first we have to write the task
			if makeDotFile {
				dotFile += newDAG.GenerateDotFile("T"+string(task.TaskID), vertexIDCounter)
			}
			for _, vertex := range newDAG {
				// write the vertex set to the file with yaml format
				_, err = file.WriteString(fmt.Sprintf("  - TaskID: %d\n", vertex.TaskID))
				_, err = file.WriteString(fmt.Sprintf("    VertexID: %d\n", vertex.VertexID+vertexIDCounter))
				_, err = file.WriteString(fmt.Sprintf("    Jitter: %d\n", vertex.Jitter))
				_, err = file.WriteString(fmt.Sprintf("    BCET: %d\n", vertex.BCET))
				_, err = file.WriteString(fmt.Sprintf("    WCET: %d\n", vertex.WCET))
				_, err = file.WriteString(fmt.Sprintf("    Period: %d\n", task.Period))
				_, err = file.WriteString(fmt.Sprintf("    Deadline: %d\n", task.Deadline))
				_, err = file.WriteString(fmt.Sprintf("    PE: %d\n", task.PE))

				succStr := "["
				for _, succ := range vertex.Successors {
					succStr += strconv.Itoa(succ+vertexIDCounter) + ","
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
			vertexIDCounter += len(newDAG)
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

		dotFile = "digraph G {\n" + dotFile + "}\n"

		if _, err := writerDot.WriteString(dotFile); err != nil {
			logger.LogFatal("Error writing to file: " + err.Error())
		}
		// close the file
		writerDot.Close()
	}

}

// findTaskSetPaths A function to find the path of all the task sets in the task set folder
func findTaskSetPaths(taskSetPath string, outputFormat string) []string {
	// we have to find all the task sets with csv extension in
	var taskSetPaths []string
	err := filepath.Walk(taskSetPath, func(path string, info os.FileInfo, err error) error {
		// check folder name to be "tasksets"
		if filepath.Ext(path) == "."+outputFormat && filepath.Base(filepath.Dir(path)) == "tasksets" {
			if strings.LastIndex(path, ".prec."+outputFormat) == -1 {
				taskSetPaths = append(taskSetPaths, path)
			}
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

	return taskSetPaths
}

// GenerateDAGSets generates DAG sets for each task set in the task set folder
func GenerateDAGSets(taskSetPath string, pPar, pAdd float64, maxParBranches, maxVertices, maxDepth int,
	makeDotFile bool, outputFormat string) {
	// first we have to find all the task sets with csv extension in
	taskSetPaths := findTaskSetPaths(taskSetPath, outputFormat)

	// now we have to generate the job sets
	for _, taskSetPath := range taskSetPaths {
		// make sure that the file does not exist
		predPath := taskSetPath[:strings.LastIndex(taskSetPath, ".")] + ".prec." + outputFormat
		if _, err := os.Stat(predPath); os.IsNotExist(err) {
			logger.LogInfo("Generating DAG for: " + taskSetPath)
			generateDAGSet(taskSetPath, pPar, pAdd, maxParBranches, maxVertices, maxDepth, makeDotFile, outputFormat)
		} else {
			logger.LogInfo(fmt.Sprintf("%s exists", predPath))
		}
	}
}

// GenerateDAGSetsParallel generates DAG sets for each task set in the task set folder in parallel
func GenerateDAGSetsParallel(taskSetPath string, pPar, pAdd float64, maxParBranches, maxVertices, maxDepth int,
	makeDotFile bool, outputFormat string) {
	// first we have to find all the task sets with csv extension in
	taskSetPaths := findTaskSetPaths(taskSetPath, outputFormat)

	// now we have to generate the job sets in parallel
	var wg sync.WaitGroup
	wg.Add(len(taskSetPaths))
	for i := 0; i < len(taskSetPaths); i++ {
		go func(setIndex int) {
			defer wg.Done()
			// make sure that the file does not exist

			predPath := taskSetPaths[setIndex][:strings.LastIndex(taskSetPaths[setIndex], ".")] + ".prec." + outputFormat
			if _, err := os.Stat(predPath); os.IsNotExist(err) {
				logger.LogInfo("Generating DAG for: " + taskSetPaths[setIndex])
				generateDAGSet(taskSetPaths[setIndex], pPar, pAdd, maxParBranches, maxVertices, maxDepth,
					makeDotFile, outputFormat)
			} else {
				logger.LogInfo(fmt.Sprintf("%s exists", predPath))
			}
		}(i)
	}
	wg.Wait()
}
