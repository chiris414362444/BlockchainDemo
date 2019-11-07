/*
  区块包，存放区块类及其方法函数
*/
package blockchain

import (
	"bytes"
	"core/algorithm"
	"core/transaction"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"time"
)

type Block struct {
	Version int32
	PrevBlockHash []byte
	MerkleRoot []byte
	Hash []byte
	Time int32
	Bits int32
	Nonce int32
	Height int32
	Transactions []*transaction.Transaction
}

// 创世区块函数
func NewGensisBlock(transactions []*transaction.Transaction) *Block {
	// 初始化区块
	block := &Block{
		Version:       2,
		PrevBlockHash: []byte{},
		MerkleRoot:    []byte{},
		Hash:          []byte{},
		Time:          int32(time.Now().Unix()),
		Bits:          404454260,
		Nonce:         0,
		Height:        0,
		Transactions:  transactions,
	}

	// 工作量证明
	pow := NewProofOfWork(block)
	// 开始挖矿, 并返回当前区块的随机数Nonce和Hash值
	nonce, hash := pow.Mine()
	block.Nonce = nonce
	block.Hash = hash

	// 打印当前区块
	block.String()
	return block
}

func NewBlock(transactions []*transaction.Transaction, prevBlockHash []byte, height int32) *Block {
	block := &Block{
		Version:       2,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    []byte{},
		Hash:          []byte{},
		Time:          int32(time.Now().Unix()),
		Bits:          404454260,
		Nonce:         0,
		Height:        height,
		Transactions:  transactions,
	}

	// 工作量证明
	pow := NewProofOfWork(block)
	// 开始挖矿, 并返回当前区块的随机数Nonce和Hash值
	nonce, hash := pow.Mine()
	block.Nonce = nonce
	block.Hash = hash
	return block
}

// 创建当前区块的默克尔根
func (block * Block) CreateMerkleTreeRoot(transactions []*transaction.Transaction) {
	var transHash [][]byte
	for _, tx := range transactions {
		transHash = append(transHash, tx.Hash())
	}

	mTree := algorithm.NewMerkleTree(transHash)
	block.MerkleRoot = mTree.RootNode.Data
}

// 打印当前区块内容
func (block *Block) String() {
	fmt.Println("----------------------------------------------------------------------------------")
	fmt.Printf("当前区块的版本号：%s\n", strconv.FormatInt(int64(block.Version), 10))
	fmt.Printf("当前区块的前一区块HASH值：%x\n", block.PrevBlockHash)
	fmt.Printf("当前区块的默克尔根：%x\n", block.MerkleRoot)
	fmt.Printf("当前区块的HASH值：%x\n", block.Hash)
	fmt.Printf("当前区块的时间：%s\n", strconv.FormatInt(int64(block.Time), 10))
	fmt.Printf("当前区块的难度：%s\n", strconv.FormatInt(int64(block.Bits), 10))
	fmt.Printf("当前区块的Nonce：%s\n", strconv.FormatInt(int64(block.Nonce), 10))
	fmt.Println("----------------------------------------------------------------------------------")
}

// 比特币中区块结构体的序列化方法
//func (core *Block) Serialize() []byte {
//	// 将当前区块中所有信息拼接序列化
//	result := bytes.Join([][]byte{
//		utils.IntToHex(core.Version, true),
//		core.PrveBlockHash,
//		core.MerkleRoot,
//		utils.IntToHex(core.Time, true),
//		utils.IntToHex(core.Bits, true),
//		utils.IntToHex(core.Nonce, true),
//	}, []byte{})
//	return result
//}

// 通过gob序列化区块为二进制
func (block *Block) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

// 通过gob反序列化二进制为区块对象
func DeserializeBlock(d []byte) *Block {
	var block Block
	decode := gob.NewDecoder(bytes.NewReader(d))
	err := decode.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}