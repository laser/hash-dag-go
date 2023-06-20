package merkle

import gvz "github.com/laser/random-dag-generator-go/render/graphviz"

func (g Graph) GetNodes() (out []gvz.RenderableGraphNode) {
	for _, node := range g.Nodes {
		out = append(out, node)
	}

	return
}

func (g Graph) GetEdges() (out []gvz.RenderableGraphEdge) {
	for _, edge := range g.Edges {
		out = append(out, edge)
	}

	return
}

func (n Node) GetID() string {
	return string(n.Id)
}

func (e Edge) GetSourceNodeID() string {
	return string(e.SourceNodeId)
}

func (e Edge) GetTargetNodeID() string {
	return string(e.TargetNodeId)
}
