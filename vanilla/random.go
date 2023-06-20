package vanilla

import (
	"math/rand"
	"time"

	dag "github.com/laser/random-dag-generator-go"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func Random(nodeQty, maxOutdegree int, edgeFactor float64) (out Graph) {
	graph := dag.Random(
		dag.WithNodeQty(nodeQty),
		dag.WithMaxOutdegree(maxOutdegree),
		dag.WithEdgeFactor(edgeFactor))

	for _, node := range graph.Nodes {
		buf := make([]byte, 32)
		rng.Read(buf)

		out.Nodes = append(out.Nodes, Node{Id: NodeId(node.Id), Data: buf})
	}

	for _, edge := range graph.Edges {
		out.Edges = append(out.Edges, Edge{
			SourceNodeId: NodeId(edge.SourceNodeId),
			TargetNodeId: NodeId(edge.TargetNodeId),
		})
	}

	return
}
