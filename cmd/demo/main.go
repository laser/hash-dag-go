package main

import (
	"flag"
	"math/rand"

	gvz "github.com/laser/random-dag-generator-go/render/graphviz"

	"github.com/laser/merkle-dag-go"
	"github.com/laser/merkle-dag-go/vanilla"
)

var (
	nodeQty      = flag.Int("node-qty", 1+rand.Intn(20), "number of nodes in the DAG")
	maxOutdegree = flag.Int("max-outdegree", 1+rand.Intn(4), "max number of edges directed out of a node")
	edgeFactor   = flag.Float64("edge-factor", 1.0-rand.Float64(), "probability of adding a new edge between nodes during the graph generation")
	outputPath   = flag.String("output-path", "/tmp/merkle-dag.png", "path to which the generated DAG-PNG will be saved")
)

func main() {
	flag.Parse()

	mdag := merkle.From(vanilla.Random(*nodeQty, *maxOutdegree, *edgeFactor))

	converted := gvz.From(mdag)

	gvz.RenderTo(converted, *outputPath)
}
