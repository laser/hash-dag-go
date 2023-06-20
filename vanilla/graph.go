package vanilla

type Node struct {
	Id   NodeId
	Data NodeData
}

type NodeId string
type NodeData []byte

type Edge struct {
	SourceNodeId NodeId
	TargetNodeId NodeId
}

type Graph struct {
	Nodes []Node
	Edges []Edge
}
