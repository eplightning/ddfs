package util

import (
	"git.eplight.org/eplightning/ddfs/pkg/api"
	"github.com/OneOfOne/xxhash"
	"github.com/eplightning/hashring"
)

type HashRing interface {
	Block(hash *BlockHash) *api.Node
	Shard(shard string) *api.Node
	Node(name string) *api.Node
}

type KetamaRing struct {
	base  *hashring.HashRing
	nodes map[string]*api.Node
}

func NewHashRing(nodes *api.NodeReplicaSets) *KetamaRing {
	nodeMap := make(map[string]*api.Node)
	var nodeNames []string

	for _, set := range nodes.Sets {
		for _, node := range set.Nodes {
			nodeMap[node.Name] = node
			nodeNames = append(nodeNames, node.Name)
		}
	}

	return &KetamaRing{
		base:  hashring.New(nodeNames),
		nodes: nodeMap,
	}
}

func (ring *KetamaRing) Block(hash *BlockHash) *api.Node {
	result, ok := ring.base.GetNodePosFromKey(ring.base.KeyFromBytes(hash.Bytes))
	if !ok {
		return nil
	}

	name := ring.base.GetNodeFromPos(result)
	return ring.nodes[name]
}

func (ring *KetamaRing) Shard(shard string) *api.Node {
	h := hashring.HashKey(xxhash.ChecksumString32(shard))
	result, ok := ring.base.GetNodePosFromKey(h)
	if !ok {
		return nil
	}

	name := ring.base.GetNodeFromPos(result)
	return ring.nodes[name]
}

func (ring *KetamaRing) Node(name string) *api.Node {
	return ring.nodes[name]
}
