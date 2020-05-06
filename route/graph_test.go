package route

import "testing"

func TestShortestPath(t *testing.T) {
	g := NewGraph()
	g.addEdge("a", "b", 5)
	g.addEdge("a", "c", 1)
	g.addEdge("a", "d", 10)
	g.addEdge("b", "d", 2)
	g.addEdge("c", "d", 2)
	g.addEdge("d", "b", 1)
	g.addEdge("d", "f", 1)
	g.addEdge("d", "e", 5)
	g.addEdge("f", "g", 2)
	g.addEdge("e", "g", 2)
	g.addEdge("g", "e", 1)
	g.addEdge("g", "h", 1)

	dists, paths := g.dijkstra("a")
	vertices := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	distsExp := map[VertexType]WeightType{
		"a": 0,
		"b": 4,
		"c": 1,
		"d": 3,
		"e": 7,
		"f": 4,
		"g": 6,
		"h": 7,
	}
	pathsExp := map[VertexType]string{
		"a": "a",
		"b": "a->c->d->b",
		"c": "a->c",
		"d": "a->c->d",
		"e": "a->c->d->f->g->e",
		"f": "a->c->d->f",
		"g": "a->c->d->f->g",
		"h": "a->c->d->f->g->h",
	}
	for _, v := range vertices {
		if dists[v] != distsExp[v] {
			t.Errorf("vertex %s expect distance %d got %d", v, distsExp[v], dists[v])
		}
		pathStr := printPath(paths[v])
		if pathStr != pathsExp[v] {
			t.Errorf("vertex %s expect path %s got %s", v, pathStr, pathsExp[v])
		}
	}

}
