package main

import (
	"crypto/sha256"
	"fmt"
)

type Node struct {
	Hash [32]byte
	Left  *Node
	Right *Node
}

func BuildMerkleTree(transactions []Transaction) Node {
	var leafs []Node

	for i := range transactions {
		leafs = append(leafs, Node{Hash: transactions[i].Hash()})
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

func PrintTree(root *Node, level int) {
	if root == nil {
		return
	}

	fmt.Printf("level = %v: %v\n", level, root)

	PrintTree(root.Left, level+1)
	PrintTree(root.Right, level+1)
}
