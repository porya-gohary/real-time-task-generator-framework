package common

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
