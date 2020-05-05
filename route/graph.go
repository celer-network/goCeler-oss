// Copyright 2020 Celer Network

package route

import (
	"container/heap"
	"fmt"

	"github.com/celer-network/goutils/log"
)

type VertexType = string
type WeightType = int

type QueueElement struct {
	vertex VertexType
	weight WeightType
}

// A VertexQueue is the priority queue for elements
type VertexQueue struct {
	elements []QueueElement
	// element vertex to index
	indexes map[VertexType]int
}

func (q *VertexQueue) Len() int { return len(q.elements) }

func (q *VertexQueue) Less(i, j int) bool {
	return q.elements[i].weight < q.elements[j].weight
}

func (q *VertexQueue) Swap(i, j int) {
	q.elements[i], q.elements[j] = q.elements[j], q.elements[i]
	q.indexes[q.elements[i].vertex] = i
	q.indexes[q.elements[j].vertex] = j
}

func (q *VertexQueue) Push(x interface{}) {
	element := x.(QueueElement)
	q.indexes[element.vertex] = len(q.elements)
	q.elements = append(q.elements, element)
}

func (q *VertexQueue) Pop() interface{} {
	element := q.elements[len(q.elements)-1]
	q.elements = q.elements[:len(q.elements)-1]
	delete(q.indexes, element.vertex)
	return element
}

func (q *VertexQueue) update(vertex VertexType, weight WeightType) {
	index := q.indexes[vertex]
	q.elements[index].weight = weight
	heap.Fix(q, index)
}

func (q *VertexQueue) push(vertex VertexType, weight WeightType) error {
	if _, ok := q.indexes[vertex]; ok {
		return fmt.Errorf("vertex already exist")
	}
	element := QueueElement{
		vertex: vertex,
		weight: weight,
	}
	heap.Push(q, element)
	return nil
}

func (q *VertexQueue) pop() (VertexType, error) {
	if len(q.elements) == 0 {
		return "", fmt.Errorf("vertex queue is empty")
	}
	element := heap.Pop(q).(QueueElement)
	return element.vertex, nil
}

const (
	Infinity = WeightType(^uint(0) >> 1)
)

type Graph struct {
	vertices map[VertexType]bool
	edges    map[VertexType]map[VertexType]WeightType
}

func NewGraph() *Graph {
	return &Graph{
		vertices: make(map[VertexType]bool),
		edges:    make(map[VertexType]map[VertexType]WeightType),
	}
}

func (g *Graph) getVertices() (vertices []VertexType) {
	for v := range g.vertices {
		vertices = append(vertices, v)
	}
	return vertices
}

func (g *Graph) getNeighbors(u VertexType) (vertices []VertexType) {
	for v := range g.edges[u] {
		vertices = append(vertices, v)
	}
	return vertices
}

func (g *Graph) getWeight(u, v VertexType) WeightType {
	return g.edges[u][v]
}

func (g *Graph) addEdge(u, v VertexType, w WeightType) {
	g.vertices[u] = true
	g.vertices[v] = true
	if _, ok := g.edges[u]; !ok {
		g.edges[u] = make(map[VertexType]WeightType)
	}
	g.edges[u][v] = w
}

// compute shortest distances and paths to each vertex
func (g *Graph) dijkstra(source VertexType) (map[VertexType]WeightType, map[VertexType][]VertexType) {
	distances := make(map[VertexType]WeightType)
	lastHops := make(map[VertexType]VertexType)
	distances[source] = 0
	vertexQueue := &VertexQueue{
		elements: []QueueElement{},
		indexes:  make(map[VertexType]int),
	}
	// add vertices
	for _, v := range g.getVertices() {
		if v != source {
			distances[v] = Infinity
		}
		lastHops[v] = ""
		vertexQueue.push(v, distances[v])
	}
	// compute shortest distances
	var vertices []VertexType
	for len(vertexQueue.elements) != 0 {
		u, _ := vertexQueue.pop()
		if distances[u] == Infinity {
			log.Debugf("vertex %s unreachable, stop graph computing", u)
			continue
		}
		vertices = append([]VertexType{u}, vertices...)
		for _, v := range g.getNeighbors(u) {
			newDist := distances[u] + g.getWeight(u, v)
			if newDist < distances[v] {
				distances[v] = newDist
				lastHops[v] = u
				vertexQueue.update(v, newDist)
			}
		}
	}
	// compute shortest paths
	paths := make(map[VertexType][]VertexType)
	for _, v := range vertices {
		paths[v] = append(paths[v], v)
	}
	for _, v := range vertices {
		prev := lastHops[paths[v][0]]
		for prev != "" {
			for _, u := range paths[v] {
				paths[u] = append(paths[prev], paths[u]...)
			}
			if paths[v][0] == source {
				break
			} else {
				prev = lastHops[prev]
			}
		}
	}
	return distances, paths
}

func printPath(path []VertexType) string {
	ret := ""
	if len(path) == 0 {
		return ret
	}
	for i := 0; i < len(path)-1; i++ {
		ret += path[i] + "->"
	}
	ret += path[len(path)-1]
	return ret
}
