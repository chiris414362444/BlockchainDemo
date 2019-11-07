package blockchain

import (
	"github.com/boltdb"
	"log"
)

// 区块链迭代器结构体
type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

// 根据当前区块的Hash值获取上一个区块
func (bci *BlockchainIterator) Next() *Block {
	var block *Block

	err := bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		// 通过当前Hash值获取序列化的区块
		deblock := bucket.Get(bci.currentHash)
		// 反序列化得到区块
		block = DeserializeBlock(deblock)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// 将迭代器的Hash设置为当前区块的前一个区块的Hash, 实现迭代器功能
	bci.currentHash = block.PrevBlockHash
	return block
}
