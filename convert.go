package hashdag

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/laser/hash-dag-go/vanilla"
)

type Hash = string

type configuration struct {
	hasher   func(data vanilla.NodeData) Hash
	combiner func(nodeHash Hash, parentHashes []Hash) Hash
}

type ConversionOption = func(*configuration)

func WithHasher(hasher func(node vanilla.NodeData) Hash, combiner func(Hash, []Hash) Hash) func(*configuration) {
	return func(opts *configuration) {
		opts.hasher = hasher
		opts.combiner = combiner
	}
}

func DefaultHasher(data vanilla.NodeData) Hash {
	checksum := sha256.Sum256(data)
	return hex.EncodeToString(checksum[:])[:4]
}

func DefaultCombiner(nodeHash Hash, parentHashes []Hash) Hash {
	sort.Strings(parentHashes)
	concatenated := fmt.Sprintf("%s%s", nodeHash, strings.Join(parentHashes, ""))
	checksum := sha256.Sum256([]byte(concatenated))
	return hex.EncodeToString(checksum[:])[:4]
}

func NaiveHasher(data vanilla.NodeData) Hash {
	return string(data)
}

func NaiveCombiner(nodeHash Hash, parentHashes []Hash) Hash {
	if len(parentHashes) == 0 {
		return nodeHash
	}

	return fmt.Sprintf("%s-(%s)", nodeHash, strings.Join(parentHashes, ","))
}

var defaultConfiguration = configuration{
	hasher:   DefaultHasher,
	combiner: DefaultCombiner,
}

func From(input vanilla.Graph, options ...ConversionOption) (output Graph) {
	cfg := defaultConfiguration

	for _, option := range options {
		option(&cfg)
	}

	nodes := make(map[Hash]Node)
	edges := make(map[Hash]Edge)

	// walk the graph to build lookup tables
	meta := walk(input)

	// memoize the hashes of each node
	memo := make(map[vanilla.NodeId]Hash)

	// we'll start at the root level
	curr := meta.roots

	// traverse the DAG one level at a time, starting with the roots
	for len(curr) > 0 {
		next := make(map[vanilla.NodeId]vanilla.Node)

		for _, node := range curr {
			parentHashes := make([]Hash, len(meta.parents[node.Id]))
			for i, parent := range meta.parents[node.Id] {
				parentHashes[i] = memo[parent.Id]
			}

			// ensure that the order of the parent hashes is deterministic
			sort.Strings(parentHashes)

			// compute the Hash of the current node
			hash := cfg.combiner(cfg.hasher(node.Data), parentHashes)

			// add the node to the DAG
			nodes[hash] = Node{Id: NodeId(hash), Data: NodeData(node.Data)}

			// add the edge to the DAG
			for _, parent := range parentHashes {
				edges[fmt.Sprintf("%s-%s", parent, hash)] = Edge{SourceNodeId: NodeId(parent), TargetNodeId: NodeId(hash)}
			}

			// cache our Hash so that we can use it from descendant nodes
			memo[node.Id] = hash

			// add the children of this node to the next level of the traversal
			for _, child := range meta.children[node.Id] {
				meta.qtyParentsLeftToHash[child.Id]--
				if meta.qtyParentsLeftToHash[child.Id] == 0 {
					next[child.Id] = child
				}
			}
		}

		curr = next
	}

	for _, edge := range edges {
		output.Edges = append(output.Edges, edge)
	}

	for _, node := range nodes {
		output.Nodes = append(output.Nodes, node)
	}

	sort.Slice(output.Edges, func(i, j int) bool {
		return output.Edges[i].SourceNodeId < output.Edges[j].SourceNodeId
	})

	sort.Slice(output.Nodes, func(i, j int) bool {
		return output.Nodes[i].Id < output.Nodes[j].Id
	})

	return
}

type metadata struct {
	roots                map[vanilla.NodeId]vanilla.Node
	parents              map[vanilla.NodeId][]vanilla.Node
	children             map[vanilla.NodeId][]vanilla.Node
	qtyParentsLeftToHash map[vanilla.NodeId]int
}

func walk(dag vanilla.Graph) metadata {
	lookup := make(map[vanilla.NodeId]vanilla.Node, len(dag.Nodes))

	m := metadata{
		roots:                make(map[vanilla.NodeId]vanilla.Node, 0),
		parents:              make(map[vanilla.NodeId][]vanilla.Node, 0),
		children:             make(map[vanilla.NodeId][]vanilla.Node, 0),
		qtyParentsLeftToHash: make(map[vanilla.NodeId]int, 0),
	}

	for _, node := range dag.Nodes {
		lookup[node.Id] = node
		m.roots[node.Id] = node
	}

	for _, edge := range dag.Edges {
		// anything that has a parent is not a root
		delete(m.roots, edge.TargetNodeId)

		if _, ok := m.parents[edge.TargetNodeId]; !ok {
			m.parents[edge.TargetNodeId] = []vanilla.Node{lookup[edge.SourceNodeId]}
		} else {
			m.parents[edge.TargetNodeId] = append(m.parents[edge.TargetNodeId], lookup[edge.SourceNodeId])
		}

		if _, ok := m.children[edge.SourceNodeId]; !ok {
			m.children[edge.SourceNodeId] = []vanilla.Node{lookup[edge.TargetNodeId]}
		} else {
			m.children[edge.SourceNodeId] = append(m.children[edge.SourceNodeId], lookup[edge.TargetNodeId])
		}

		// we'll use this to determine when we can Hash a node
		m.qtyParentsLeftToHash[edge.TargetNodeId]++
	}

	return m
}
