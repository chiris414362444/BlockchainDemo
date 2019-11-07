/*
  默克尔树包，存放默克尔树类及其方法函数
*/
package algorithm

import (
	"crypto/sha256"
	"utils"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}

// 构建默克尔树节点
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	merkleNode := MerkleNode{}

	// 若左右节点都为空，表示是叶子节点，否则是中间节点
	if left == nil && right == nil {
		merkleNode.Data = data
	} else {
		// 拼接左右子节点的hash值
		prevHash := append(left.Data, right.Data...)

		// 将左右子节点拼接的hash值进行双重hash，得到当前节点的hash值
		firstHash := sha256.Sum256(prevHash)
		secondHash := sha256.Sum256(firstHash[:])
		merkleNode.Data = secondHash[:]
	}

	merkleNode.Left = left
	merkleNode.Right = right
	return &merkleNode
}

// 构建默克尔树，参数data是传入的所有叶子节点值，因为节点本身的hash值也是字节数组[]byte，因此这里是二维字节数组
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// 遍历所有叶子节点
	for _, nodeData := range data {
		node := NewMerkleNode(nil, nil, nodeData)
		nodes = append(nodes, *node)
	}

	j := 0
	// nSize表示树的深度
	for nSize := len(data); nSize > 1; nSize = (nSize + 1) / 2 {
		for i := 0; i < nSize; i+=2 {
			// i2用于处理奇偶节点问题，当节点个数是奇数，将赋值节点本身作为左右节点
			i2 := utils.Min(i+1, nSize-1)
			node := NewMerkleNode(&nodes[j+i], &nodes[j+i2], nil)
			nodes = append(nodes, *node)
		}
		j+=nSize
	}

	// 默克尔树的根节点就是nodes的最后一个节点
	merkleTree := MerkleTree{&(nodes[len(nodes) - 1])}
	return &merkleTree
}
