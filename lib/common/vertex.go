package common

import (
	"strconv"
)

type Vertex struct {
	TaskID          int
	VertexID        int
	RelativeRelease int
	BCET            int
	WCET            int
	Predecessors    []int
	Successors      []int
	Depth           int
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
