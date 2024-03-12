package common

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"
)

type Vertex struct {
	TaskID       int
	VertexID     int
	Jitter       int
	BCET         int
	WCET         int
	Period       int
	Deadline     int
	Predecessors []int
	Successors   []int
	Depth        int
}

type VertexSet []*Vertex

// GenerateDotFile generates a dot file from the vertex set
func (vs *VertexSet) GenerateDotFile(name string, offset int) string {
	// create the file
	var str string

	// write the header
	str += "subgraph cluster_" + name + " {\n" + "label=\"" + name + "\";\n"

	// write the vertices
	for _, vertex := range *vs {
		name := "V" + strconv.Itoa(vertex.VertexID+offset)
		str += "\t" + strconv.Itoa(vertex.VertexID+offset) + " [label=\"" + name + "\"];\n"
	}

	// write the edges
	for _, vertex := range *vs {
		for _, successor := range vertex.Successors {
			str += "\t" + strconv.Itoa(vertex.VertexID+offset) + " -> " + strconv.Itoa((*vs)[successor].VertexID+offset) + ";\n"
		}
	}

	// write the footer
	str += "}\n"

	return str
}

// ReadVertexSet reads a vertex set from a file
func ReadVertexSet(path string) (VertexSet, error) {
	// read the vertex set from the CSV file
	var vertices VertexSet
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
		tempTaskID, _ := strconv.Atoi(record[0])
		tempVertexID, _ := strconv.Atoi(record[1])
		tempJitter, _ := strconv.Atoi(record[2])
		tempBCET, _ := strconv.Atoi(record[3])
		tempWCET, _ := strconv.Atoi(record[4])
		tempPeriod, _ := strconv.Atoi(record[5])
		tempDeadline, _ := strconv.Atoi(record[6])
		// for successors, we have to first remove the brackets
		tempSc := record[7][1 : len(record[7])-1]
		tempSuccessors := strings.Split(tempSc, ",")

		var successors []int
		for _, successor := range tempSuccessors {
			temp, _ := strconv.Atoi(successor)
			successors = append(successors, temp)
		}

		vertices = append(vertices, &Vertex{
			TaskID:     tempTaskID,
			VertexID:   tempVertexID,
			Jitter:     tempJitter,
			BCET:       tempBCET,
			WCET:       tempWCET,
			Period:     tempPeriod,
			Deadline:   tempDeadline,
			Successors: successors,
		})
	}

	return vertices, nil
}

// HyperPeriod function to calculate the hyperperiod of the vertex set
func (vs VertexSet) HyperPeriod() int {
	// calculate the hyperperiod
	hyperperiod := 1
	start := time.Now()
	for _, t := range vs {
		if time.Since(start) < 60*time.Second {
			hyperperiod = lcm(hyperperiod, t.Period)
		} else {
			return -1
		}
	}

	return hyperperiod
}
