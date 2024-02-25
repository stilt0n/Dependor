package dependor

import "testing"

func TestReverseEdges(t *testing.T) {
	testGraph := DependencyGraph{"goose": {"wild", "chase"}}
	expected := DependencyGraph{"wild": {"goose"}, "chase": {"goose"}}
	reversed := testGraph.ReverseEdges()
	if len(expected) != len(reversed) {
		t.Fatalf("Expected revesed length to be %d but received %d\n", len(expected), len(reversed))
	}

	for node, edges := range reversed {
		if expectedEdges, ok := expected[node]; !ok {
			t.Errorf("Received unexpected node %q\n", node)
		} else {
			if len(expectedEdges) != len(edges) {
				t.Fatalf("Received wrong number of edges. Expected %d but received %d\n", len(expectedEdges), len(edges))
			}
			for i := 0; i < len(edges); i++ {
				if expectedEdges[i] != edges[i] {
					t.Errorf("Expected edge %d to be %q but received %q\n", i, expectedEdges[i], edges[i])
				}
			}
		}
	}
}

func TestTraverse(t *testing.T) {
	testGraph := DependencyGraph{"wild": {"goose"}, "goose": {"chase"}, "chase": {}, "notRelated": {}}
	var appendTo []string
	testGraph.Traverse("wild", func(node string) {
		appendTo = append(appendTo, node)
	})
	expected := []string{"wild", "goose", "chase"}
	if len(appendTo) != len(expected) {
		t.Fatalf("Expected slice of traversed nodes to have %d nodes but received slice with %d nodes", len(expected), len(appendTo))
	}
	for i, node := range appendTo {
		if expected[i] != node {
			t.Errorf("Expected %q at index %d but received %q\n", expected[i], i, node)
		}
	}
}
