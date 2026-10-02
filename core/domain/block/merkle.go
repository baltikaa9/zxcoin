package block

import (
	"crypto/sha256"

	"github.com/baltikaa9/zxcoin/core/domain/types"
)

type Node struct {
	Hash  types.Hash
	Left  *Node
	Right *Node
}

func buildMerkleTree(hashes []types.Hash) Node {
	if len(hashes) == 0 {
		return Node{}
	}

	var leafs []Node

	for _, hash := range hashes {
		leafs = append(leafs, Node{Hash: hash})
	}

	return buildLevel(leafs)[0]
}

func buildLevel(nodes []Node) []Node {
	if len(nodes) == 1 {
		return nodes
	}

	var newLevel []Node

	if len(nodes)%2 == 1 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	for i := 0; i < len(nodes); i += 2 {
		newLevel = append(newLevel, Node{
			Hash:  sha256.Sum256(append(nodes[i].Hash[:], nodes[i+1].Hash[:]...)),
			Left:  &nodes[i],
			Right: &nodes[i+1],
		})
	}

	return buildLevel(newLevel)
}
