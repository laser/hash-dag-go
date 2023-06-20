package hashdag_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/laser/hash-dag-go/vanilla"

	rdag "github.com/laser/random-dag-generator-go"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/require"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func TestConvert(t *testing.T) {
	t.Run("empty DAG", func(t *testing.T) {
		input := vanilla.Graph{
			Nodes: []vanilla.Node{},
			Edges: []vanilla.Edge{},
		}

		expectedNodeIds := mapset.NewSet[string]()
		expectedEdges := mapset.NewSet[string]()

		runConvertTest(t, input, expectedNodeIds, expectedEdges)
	})

	t.Run("one-node DAG", func(t *testing.T) {
		input := vanilla.Graph{
			Nodes: []vanilla.Node{
				{Id: "41c", Data: []byte("a")},
			},
			Edges: []vanilla.Edge{},
		}

		expectedNodeIds := mapset.NewSet("a")
		expectedEdges := mapset.NewSet[string]()

		runConvertTest(t, input, expectedNodeIds, expectedEdges)
	})

	t.Run("two-node DAG, one root", func(t *testing.T) {
		input := vanilla.Graph{
			Nodes: []vanilla.Node{{Id: "41c", Data: []byte("a")}, {Id: "ff3", Data: []byte("b")}},
			Edges: []vanilla.Edge{{SourceNodeId: "41c", TargetNodeId: "ff3"}},
		}

		expectedNodeIds := mapset.NewSet("a", "b-(a)")
		expectedEdges := mapset.NewSet("a->b-(a)")

		runConvertTest(t, input, expectedNodeIds, expectedEdges)
	})

	t.Run("three-node DAG, two roots, one terminal node", func(t *testing.T) {
		input := vanilla.Graph{
			Nodes: []vanilla.Node{
				{Id: "41c", Data: []byte("a")},
				{Id: "ff3", Data: []byte("b")},
				{Id: "999", Data: []byte("c")},
			},
			Edges: []vanilla.Edge{
				{SourceNodeId: "41c", TargetNodeId: "ff3"},
				{SourceNodeId: "999", TargetNodeId: "ff3"},
			},
		}

		expectedNodeIds := mapset.NewSet("a", "c", "b-(a,c)")
		expectedEdges := mapset.NewSet("a->b-(a,c)", "c->b-(a,c)")

		runConvertTest(t, input, expectedNodeIds, expectedEdges)
	})

	t.Run("four-node DAG, two roots, one terminal node", func(t *testing.T) {
		input := vanilla.Graph{
			Nodes: []vanilla.Node{
				{Id: "41c", Data: []byte("a")},
				{Id: "ff3", Data: []byte("b")},
				{Id: "999", Data: []byte("c")},
				{Id: "ddd", Data: []byte("d")},
			},
			Edges: []vanilla.Edge{
				{SourceNodeId: "41c", TargetNodeId: "ff3"},
				{SourceNodeId: "41c", TargetNodeId: "999"},
				{SourceNodeId: "ff3", TargetNodeId: "999"},
				{SourceNodeId: "ddd", TargetNodeId: "999"},
			},
		}

		expectedNodeIds := mapset.NewSet("a", "b-(a)", "c-(a,b-(a),d)", "d")
		expectedEdges := mapset.NewSet("a->b-(a)", "a->c-(a,b-(a),d)", "b-(a)->c-(a,b-(a),d)", "d->c-(a,b-(a),d)")

		runConvertTest(t, input, expectedNodeIds, expectedEdges)
	})

	t.Run("parent Hash-sorting is lexicographical", func(t *testing.T) {
		input := vanilla.Graph{
			Nodes: []vanilla.Node{
				{Id: "41c", Data: []byte("a")},
				{Id: "ff3", Data: []byte("b")},
				{Id: "999", Data: []byte("c")}},
			Edges: []vanilla.Edge{
				{SourceNodeId: "999", TargetNodeId: "ff3"},
				{SourceNodeId: "41c", TargetNodeId: "ff3"},
			},
		}

		expectedNodeIds := mapset.NewSet("a", "c", "b-(a,c)")
		expectedEdges := mapset.NewSet("a->b-(a,c)", "c->b-(a,c)")

		runConvertTest(t, input, expectedNodeIds, expectedEdges)
	})

	t.Run("output DAG always has the same number of edges as input DAG", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			input := randomDAG()

			actual := hashdag.From(input)

			require.Equal(t, len(input.Edges), len(actual.Edges))
		}
	})

	t.Run("output DAG always has same quantity of nodes as input DAG", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			input := randomDAG()

			actual := hashdag.From(input)

			require.Equal(t, len(input.Nodes), len(actual.Nodes))
		}
	})

	t.Run("output DAG has at least N edges where N is greater or equal to the number of nodes in the DAG, minus one", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			input := randomDAG()

			actual := hashdag.From(input)

			require.True(t, len(actual.Edges) >= len(actual.Nodes)-1)
		}
	})
}

func runConvertTest(t *testing.T, input vanilla.Graph, expectedNodeIds mapset.Set[string], expectedEdges mapset.Set[string]) {
	actual := hashdag.From(input, hashdag.WithHasher(hashdag.NaiveHasher, hashdag.NaiveCombiner))

	actualNodeIds := mapset.NewSet[string]()
	for _, node := range actual.Nodes {
		actualNodeIds.Add(string(node.Id))
	}

	actualEdges := mapset.NewSet[string]()
	for _, edge := range actual.Edges {
		actualEdges.Add(fmt.Sprintf("%s->%s", edge.SourceNodeId, edge.TargetNodeId))
	}

	w := expectedNodeIds.ToSlice()
	x := actualNodeIds.ToSlice()
	y := expectedEdges.ToSlice()
	z := actualEdges.ToSlice()

	for _, v := range [][]string{w, x, y, z} {
		sort.Strings(v)
	}

	require.Equal(t, w, x, "nodes did not match")
	require.Equal(t, y, z, "edges did not match")
}

func randomDAG() (out vanilla.Graph) {
	graph := rdag.Random(
		rdag.WithNodeQty(1+rand.Intn(20)),
		rdag.WithMaxOutdegree(1+rand.Intn(4)),
		rdag.WithEdgeFactor(1.0-rand.Float64()))

	for _, node := range graph.Nodes {
		// create a 32-byte buffer of random stuff
		buf := make([]byte, 32)
		rng.Read(buf)

		out.Nodes = append(out.Nodes, vanilla.Node{
			Id:   vanilla.NodeId(node.Id),
			Data: buf,
		})
	}

	for _, edge := range graph.Edges {
		out.Edges = append(out.Edges, vanilla.Edge{
			SourceNodeId: vanilla.NodeId(edge.SourceNodeId),
			TargetNodeId: vanilla.NodeId(edge.TargetNodeId),
		})
	}

	return
}
