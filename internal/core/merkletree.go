package core

import (
	"crypto/sha256"
	"slices"
)

type Content interface {
	Hash() [32]byte
}

type Node struct {
	Tree    *MerkleTree
	Parent  *Node
	Left    *Node
	Right   *Node
	Hash    [32]byte
	Content Content
	leaf    bool
	dup     bool
}

type MerkleTree struct {
	Root   *Node
	Leaves []*Node
}

func NewMarkleTree(contents []Content) *MerkleTree {
	tree := &MerkleTree{}

	root, leaves := buildWithContent(contents, tree)

	tree.Root = root
	tree.Leaves = leaves

	return tree
}

func (tree *MerkleTree) MerkleRoot() [32]byte {
	return tree.Root.Hash
}

func (tree *MerkleTree) GetMerklePath(content Content) ([][32]byte, []int) {
	contentHash := content.Hash()
	for _, current := range tree.Leaves {
		if current.Hash != contentHash {
			continue
		}

		currentParent := current.Parent
		var merklePath [][32]byte
		var index []int
		for currentParent != nil {
			if currentParent.Left.Hash == current.Hash {
				merklePath = append(merklePath, currentParent.Right.Hash)
				index = append(index, 1) // right leaf, concat second
			} else {
				merklePath = append(merklePath, currentParent.Left.Hash)
				index = append(index, 0) // left leaf, concat first
			}
			current = currentParent
			currentParent = currentParent.Parent
		}

		return merklePath, index
	}

	return nil, nil
}

func buildWithContent(contents []Content, tree *MerkleTree) (*Node, []*Node) {
	var leaves []*Node
	for _, content := range contents {
		hash := content.Hash()
		leaves = append(leaves, &Node{
			Tree:    tree,
			Hash:    hash,
			Content: content,
			leaf:    true,
		})
	}

	if len(leaves)%2 == 1 {
		leaves = append(leaves, &Node{
			Tree:    tree,
			Hash:    leaves[len(leaves)-1].Hash,
			Content: leaves[len(leaves)-1].Content,
			leaf:    true,
			dup:     true,
		})
	}

	root := buildIntermediate(leaves, tree)

	return root, leaves
}

func buildIntermediate(nl []*Node, tree *MerkleTree) *Node {
	var nodes []*Node
	for i := 0; i < len(nl); i += 2 {
		left, right := i, i+1
		if right == len(nl) {
			right = left
		}

		node := &Node{
			Tree:  tree,
			Left:  nl[left],
			Right: nl[right],
			Hash:  sha256.Sum256(slices.Concat(nl[left].Hash[:], nl[right].Hash[:])),
		}

		nodes = append(nodes, node)

		nl[left].Parent = node
		nl[right].Parent = node

		if len(nl) == 2 {
			return node
		}
	}

	return buildIntermediate(nodes, tree)
}
