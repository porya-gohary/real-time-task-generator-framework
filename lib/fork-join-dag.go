package lib

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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
		vertices = append(vertices, &common.Vertex{VertexID: 0, Depth: depth})
		vertices = append(vertices, &common.Vertex{VertexID: 1, Depth: -depth})

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
		vertices[i].TaskID, _ = strconv.Atoi(task.Name[1:])
		vertices[i].RelativeRelease = task.Jitter
		vertices[i].BCET = bcetList[i]
		vertices[i].WCET = wcetList[i]
	}
	return vertices
}

func generateDAGSet(taskPath string, pPar, pAdd float64, maxParBranches, maxVertices, maxDepth int) {
	// first we have to read the task set
	taskSet, err := common.ReadTaskSet(taskPath)
	if err != nil {
		logger.LogFatal("Error reading task set: " + err.Error())
	}

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

	for _, task := range taskSet {
		newDAG := generateDAGFromTask(*task, pPar, pAdd, maxParBranches, maxVertices, maxDepth)
		// first we have to write the task
		writer.Write([]string{"T", strconv.Itoa(newDAG[0].TaskID), "0", strconv.Itoa(task.Period),
			strconv.Itoa(task.Deadline)})
		for _, vertex := range newDAG {
			row := []string{
				"V",
				strconv.Itoa(vertex.TaskID),
				strconv.Itoa(vertex.VertexID),
				strconv.Itoa(vertex.RelativeRelease),
				strconv.Itoa(vertex.BCET),
				strconv.Itoa(vertex.WCET),
			}
			for _, pred := range vertex.Predecessors {
				row = append(row, strconv.Itoa(pred))
			}
			if err := writer.Write(row); err != nil {
				logger.LogFatal("Error writing to file: " + err.Error())
			}
		}
	}

}

// GenerateDAGSets generates DAG sets for each task set in the task set folder
func GenerateDAGSets(taskSetPath string, pPar, pAdd float64, maxParBranches, maxVertices, maxDepth int) {
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
			generateDAGSet(taskSetPath, pPar, pAdd, maxParBranches, maxVertices, maxDepth)
		} else {
			logger.LogInfo(fmt.Sprintf("%s exists", predPath))
		}
	}
}

// GenerateDAGSetsParallel generates DAG sets for each task set in the task set folder in parallel
func GenerateDAGSetsParallel(taskSetPath string, pPar, pAdd float64, maxParBranches, maxVertices, maxDepth int) {
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
	// now we have to generate the job sets in parallel
	var wg sync.WaitGroup
	wg.Add(len(taskSetPaths))
	for _, taskSetPath := range taskSetPaths {
		go func(taskSetPaths string) {
			defer wg.Done()
			// make sure that the file does not exist
			predPath := taskSetPath[:len(taskSetPath)-4] + ".prec" + ".csv"
			if _, err := os.Stat(predPath); os.IsNotExist(err) {
				logger.LogInfo("Generating DAG for: " + taskSetPath)
				go generateDAGSet(taskSetPath, pPar, pAdd, maxParBranches, maxVertices, maxDepth)
			} else {
				logger.LogInfo(fmt.Sprintf("%s exists", predPath))
			}
		}(taskSetPath)
	}
	wg.Wait()
}
