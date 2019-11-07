package blockchain

import (
	"core/transaction"
	"encoding/hex"
	"github.com/boltdb"
	"log"
)

const utxoBucket = "chainset"

type UTXOSet struct {
	bc *Blockchain
}

// 构建UTXOSet
func NewUTXOSet(bc *Blockchain) *UTXOSet {
	return &UTXOSet{bc}
}

// 将UTXO放入数据库
func (u UTXOSet) Reindex() {
	db := u.bc.db
	bucketName := []byte(utxoBucket)

	// 删除已有的桶, 并创建新桶
	err := db.Update(func(tx *bolt.Tx) error {
		// 删除数据库中已有的bucketName桶
		err1 := tx.DeleteBucket(bucketName)
		// 允许桶不存在
		if err1 != nil && err1 != bolt.ErrBucketNotFound {
			log.Panic(err1)
		}

		// 重新创建桶
		_, err2 := tx.CreateBucket(bucketName)
		if err2 != nil {
			log.Panic(err2)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// 获取区块链的所有UTXO
	UTXO := u.bc.FindAllUTXO()

	// 将区块链的UTXO写入数据库
	err = db.Update(func(tx *bolt.Tx) error {
		// 打开桶
		bucket := tx.Bucket(bucketName)

		// 循环UTXO
		for txID, outs := range UTXO {
			// 将txID string转为字节
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			// 往桶放入UXTO, key: 交易ID, Value: 交易所有未花费的输出的序列化
			err = bucket.Put(key, transaction.SerializeOutputs(outs))
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// 根据公钥Hash获取对应公钥的UTXO
func (u UTXOSet) FindUTXOByPubkeyHash(pubkeyHash []byte) []transaction.TXOutput {
	var UTXOs []transaction.TXOutput

	db := u.bc.db
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))

		// Cursor游标
		cursor := bucket.Cursor()

		// 循环遍历当前桶  key: 交易ID, Value: 当前交易所有未花费的输出的序列化
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			// 反序列化输出
			outs := transaction.DeserializeOutputs(value)

			// 循环遍历当前交易所有未花费的输出
			for _, out := range outs.Outputs {
				// 如果当前输出属于当前公钥
				if out.CanBeUnlockedWith(pubkeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// 通过一个区块更新数据库中的UTXO（在生成或接受一个新的区块时使用）
func (u UTXOSet) UpdateUTXOByBlock(block *Block) {
	db := u.bc.db
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))

		// 循环遍历区块的所有交易
		for _, tx := range block.Transactions {
			// 如果交易不是Coinbase交易
			if tx.IsCoinBase() == false {
				// 循环遍历交易的输入
				for _, vin := range tx.Vin {
					// 存放当前输入所引用的输出所在的交易的所有输出
					updateOuts := transaction.TXOutputs{}

					// 获取当前输入对应的输出所在交易所有未花费的输出的序列化
					outsBytes := bucket.Get(vin.TXid)
					// 反序列化交易所有未花费的输出
					outs := transaction.DeserializeOutputs(outsBytes)

					// 循环遍历交易的输出
					for outIdx, out := range outs.Outputs {
						// 如果输出的ID不等于当前交易输入所引用的输出, 则记录未花费输出
						if outIdx != vin.VoutIndex {
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}

					// 如果为0表示当前交易的所有输出已被花费, 则从数据库中删除该交易的UTXO
					if len(updateOuts.Outputs) == 0 {
						err := bucket.Delete(vin.TXid)
						if err != nil {
							return err
						}
					} else {
						// 存在未花费输出则记录数据库
						err := bucket.Put(vin.TXid, transaction.SerializeOutputs(updateOuts))
						if err != nil {
							return err
						}
					}
				}
			}

			// 当前区块的交易所有输出均是未花费的输出, 存入数据库
			newOutputs := transaction.TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := bucket.Put(tx.ID, transaction.SerializeOutputs(newOutputs))
			if err != nil {
				return err
			}
		}

		return nil
	})


	if err != nil {
		log.Panic(err)
	}
}
