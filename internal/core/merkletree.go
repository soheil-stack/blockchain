package core

import (
	"crypto/sha256"
	"slices"

	"github.com/ethereum/go-ethereum/common"
)

type Content interface {
	Hash() common.Hash
}

type Node[T Content] struct {
	Tree    *MerkleTree[T]
	Parent  *Node[T]
	Left    *Node[T]
	Right   *Node[T]
	Hash    common.Hash
	Content Content
	leaf    bool
	dup     bool
}

type MerkleTree[T Content] struct {
	Root   *Node[T]
	Leaves []*Node[T]
}

func NewMarkleTree[T Content](contents []T) *MerkleTree[T] {
	tree := &MerkleTree[T]{}

	root, leaves := buildWithContent(contents, tree)

	tree.Root = root
	tree.Leaves = leaves

	return tree
}

func (tree *MerkleTree[T]) MerkleRoot() common.Hash {
	return tree.Root.Hash
}

func (tree *MerkleTree[T]) GetMerklePath(content Content) ([]common.Hash, []int) {
	contentHash := content.Hash()
	for _, current := range tree.Leaves {
		if current.Hash != contentHash {
			continue
		}

		currentParent := current.Parent
		var merklePath []common.Hash
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

func buildWithContent[T Content](contents []T, tree *MerkleTree[T]) (*Node[T], []*Node[T]) {
	var leaves []*Node[T]
	for _, content := range contents {
		hash := content.Hash()
		leaves = append(leaves, &Node[T]{
			Tree:    tree,
			Hash:    hash,
			Content: content,
			leaf:    true,
		})
	}

	if len(leaves)%2 == 1 {
		leaves = append(leaves, &Node[T]{
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

func buildIntermediate[T Content](nl []*Node[T], tree *MerkleTree[T]) *Node[T] {
	var nodes []*Node[T]
	for i := 0; i < len(nl); i += 2 {
		left, right := i, i+1
		if right == len(nl) {
			right = left
		}

		node := &Node[T]{
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
