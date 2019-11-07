package blockchain

import (
	"bytes"
	"core/transaction"
	"core/wallet"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb"
	"log"
)

// 定义数据库文件名
const dbFile = "blockchain.db"

// 定义一个桶
const blockBucket = "blocks"

// 创世区块内容
const genesisData = "这是创世区块的内容"

// 区块链结构体
type Blockchain struct {
	currentHash []byte  // 最近的一个区块的Hash值
	db *bolt.DB
}

// 构建区块链的迭代器
func (bc *Blockchain) iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.currentHash, bc.db}
}

// 获取当前区块链, 最近的一个区块的Hash值
func (bc *Blockchain) GetCurrentHash() []byte {
	return bc.currentHash
}

// 获取当前区块链的高度
func (bc *Blockchain) GetBestHeight() int32 {
	// 当前数据库最长区块的高度
	var lastHeight int32

	// 读取当前区块链所在的数据库
	err := bc.db.View(func(tx *bolt.Tx) error {
		// 打开当前桶
		bucket := tx.Bucket([]byte(blockBucket))
		// 通过区块的Hash得到区块的序列化
		blockData := bucket.Get(bc.currentHash)
		// 反序列化得到区块结构体
		block := *DeserializeBlock(blockData)
		// 获取当前区块高度
		lastHeight = block.Height
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return lastHeight
}

// 获取区块链的所有区块的Hash集合
func (bc *Blockchain) GetAllBlockHash() [][]byte {
	var blocksHash [][]byte

	bci := bc.iterator()

	for {
		block := bci.Next()

		blocksHash = append(blocksHash, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocksHash
}

// 根据区块的ID获取区块
func (bc *Blockchain) GetBlockById(id []byte) (Block, error) {
	var block Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		blockData := bucket.Get(id)

		if blockData == nil {
			return errors.New("未找到区块")
		}

		block = *DeserializeBlock(blockData)
		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

// 往区块链中加入新区块
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))

		// 根据区块的Hash判断区块是否已在数据库中存在, 存在不添加
		blockIndb := bucket.Get(block.Hash)
		if blockIndb != nil {
			return nil
		}

		// 将区块数据进行序列化
		blockData := block.Serialize()

		// 加入数据库
		err := bucket.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		// 获取区块链中最后一个区块
		lastHash := bucket.Get([]byte("l"))
		lastBlockData := bucket.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		// 如果当前加入的区块的高度大于最后一个区块, 则更新最后一个区块
		if block.Height > lastBlock.Height {
			err = bucket.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}

			// 更新区块链当前最新区块的Hash
			bc.currentHash = block.Hash
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// 往区块链中加入区块(即挖矿)
func (bc * Blockchain) MineBlock(transactions []*transaction.Transaction) *Block{
	// 验证所有交易是否有效
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) == false {
			log.Panic("Error: INVALID TRANSACTION!")
		}
	}

	// 获取当前数据库最长区块的高度
	lastHeight := bc.GetBestHeight()

	// 根据前一区块hash和高度构建当前区块, 新的区块的高度比上一区块增加1
	newBlock := NewBlock(transactions, bc.currentHash, lastHeight + 1)

	// 更新当前区块链所在的数据库
	err := bc.db.Update(func(tx *bolt.Tx) error {
		// 打开当前桶
		bucket := tx.Bucket([]byte(blockBucket))

		// 往桶里插入key:当前区块Hash, value: 当前区块的序列化值
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		// 将最新区块的Hash更新到key=l，数据库中l对应了最新的区块
		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return err
		}

		// 更新当前区块链的最近一个区块hash为当前区块hash
		bc.currentHash = newBlock.Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// 循环打印区块链
func (bc *Blockchain) PrintBlockchain() {
	bci := bc.iterator()

	for {
		block := bci.Next()
		block.String()
		fmt.Println()

		// 当前区块的前一区块的Hash值为0, 则表示当前区块为第一个区块
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

// 查询所有交易的未花费输出
func (bc *Blockchain) FindAllUTXO() map[string]transaction.TXOutputs {
	// 存放所有交易的未花费输出 string：交易的ID --> []TXOutput：交易对应的未被花费的输出集合
	UTXO := make(map[string]transaction.TXOutputs)

	// 存放已经花费的输出的交易 string：交易的ID --> []int：已经被花费的输出的序号
	spendTXOs := make(map[string][]int)

	// 通过区块链的迭代器循环区块链
	bci := bc.iterator()
	for {
		block := bci.Next()

		// 循环遍历区块的交易
		for _, tx := range block.Transactions {
			// 将交易的ID转为string
			txID := hex.EncodeToString(tx.ID)

			// 循环遍历交易的输出
			Outputs:
			for outId, out := range tx.Vout {
				// 标识当前交易已花费的输出
				spentOuts :=  spendTXOs[txID]

				// 不等于nil, 表示当前交易一定有已被花费的输出
				if spentOuts != nil {
					// 循环当前交易已花费的输出, 判断当前输出的id是否已存在, 存在则表示当前输出已被花费
					for _, spentOut := range spentOuts {
						if spentOut == outId {  // 存在代表当前输出已经被花费, 退出当前循环
							continue Outputs
						}
					}
				}

				// 在已花费的UTXO中未找到当前输出, 表示当前输出尚未被花费, 存入UTXO
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			// 如果区块不是CoinBase, 则记录交易的输入到已花费UTXO中
			if tx.IsCoinBase() == false {
				// 循环交易的输入
				for _, in := range tx.Vin {
					inTxId := hex.EncodeToString(in.TXid)
					// 记录交易的输入到已花费UTXO中  string：交易的ID --> []int：输入对应的输出
					spendTXOs[inTxId] = append(spendTXOs[txID], in.VoutIndex)
				}
			}
		}

		// 如果上一个区块的Hash为空, 则退出循环
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// 根据公钥获取存在未花费输出的交易
func (bc *Blockchain) FindUnspentTransactions(pubkeyHash []byte) []transaction.Transaction {
	// 存放所有存在未花费输出的交易
	var unspentTXs []transaction.Transaction

	// 存放已经花费的输出的交易 string：交易的ID --> []int：已经被花费的输出的序号
	spendTXOs := make(map[string][]int)

	// 通过区块链的迭代器循环区块链
	bci := bc.iterator()
	for {
		// 获取前一个区块
		block := bci.Next()

		// 循环遍历区块的交易
		for _, tx := range block.Transactions {
			// 将交易的ID转为string
			txID := hex.EncodeToString(tx.ID)

			// 循环遍历交易的输出
			output:
			for outId, out := range tx.Vout {
				// 标识当前交易已花费的输出
				spentOuts :=  spendTXOs[txID]

				// 不等于nil, 表示当前交易一定有已被花费的输出
				if spentOuts != nil {
					// 循环当前交易已花费的输出, 判断当前输出的id是否已存在, 存在则表示当前输出已被花费
					for _, spentOut := range spentOuts {
						if spentOut == outId {  // 存在退出当前循环
							continue output
						}
					}
				}

				// 如果当前输出未被花费, 且当前输出是属于该地址, 加入未花费交易
				if out.CanBeUnlockedWith(pubkeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			// 循环遍历交易的输入(CoinBase 不用引用前一个输出, 不用存储)
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					// 如果输入属于当前地址, 则将输入的输出id记录已花费中
					if in.CanUnlockOutputWith(pubkeyHash) {
						inTxID := hex.EncodeToString(in.TXid)
						spendTXOs[inTxID] = append(spendTXOs[inTxID], in.VoutIndex)
					}
				}
			}
		}

		// 当前区块的前一区块的Hash值为0, 则表示当前区块为第一个区块
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// 根据Pubkey Hash获取未花费的输出
func (bc *Blockchain) FindUTXO(pubkeyHash []byte) []transaction.TXOutput {
	// 存放未花费的输出
	var UTXOs []transaction.TXOutput

	// 根据地址获取存在未花费输出的交易
	UTXs := bc.FindUnspentTransactions(pubkeyHash)

	// 循环遍历交易, 获取属于当前地址的输出
	for _, tx := range UTXs {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubkeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

/*
	summary：根据转账地址和待转账金额获取能够转账的金额和相应的有效的输出
	address：查询地址
	amount: 需要获取的金额
	return: 获取的总金额； 未花费的输出与交易的映射
 */
func (bc *Blockchain) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	// 存放未花费的输出的交易 string：交易的Hash --> []int：未花费的输出的序号
	unspenTXOs := make(map[string][]int)

	// 根据地址获取存在未花费输出的交易
	unspentTXs := bc.FindUnspentTransactions(pubkeyHash)

	// 获取的总金额
	total := 0

	// 循环遍历未花费的交易
	Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		// 循环遍历交易的输出
		for outIdx, out := range tx.Vout {
			// 输出属于当前地址则记录
			if out.CanBeUnlockedWith(pubkeyHash) && total < amount {
				total += out.Value
				unspenTXOs[txID] = append(unspenTXOs[txID], outIdx)

				// 总金额已大于或等于需要获得金额则退出循环
				if total >= amount {
					break Work
				}
			}
		}
	}

	return total, unspenTXOs
}

func (bc *Blockchain) SignTransaction(tx *transaction.Transaction, privateKey ecdsa.PrivateKey) {
	// 存放当前交易的输入所引用的全部交易 key : 交易ID, value : 交易的结构体
	prevTXs := make(map[string]transaction.Transaction)

	// 遍历交易的所有输入, 得到每个输入对应的交易id和交易结构体
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransactionByID(vin.TXid)
		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	// 根据私钥和交易引用的输入所引用的输出的交易，对交易进行数据签名
	tx.Sign(privateKey, prevTXs)
}

// 通过交易ID查找交易
func (bc *Blockchain) FindTransactionByID(ID []byte) (transaction.Transaction, error) {
	bci := bc.iterator()

	// 循环遍历区块链
	for {
		block := bci.Next()

		// 循环遍历区块的交易
		for _, tx := range block.Transactions {
			// 交易ID相同则退出
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		// 到达第一个区块退出
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("未找到相关交易")
}

// 验证交易是否有效
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) bool {
	// 存放当前交易的输入所引用的全部交易 key : 交易ID, value : 交易的结构体
	prevTXs := make(map[string]transaction.Transaction)

	// 循环遍历交易的输入
	for _, vin := range tx.Vin {
		// 通过输入的TXid(即：当前输入引用的前一笔输出所在的交易号), 寻找到前一笔交易
		prevTX, err := bc.FindTransactionByID(vin.TXid)
		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// 创建新的区块链
func NewBlockchain(address string) *Blockchain {
	// 定义当前最近的一个区块的Hash值
	var tip []byte
	// 打开当前数据库文件
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// 构建一个桶
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			fmt.Println("数据库中不存在区块链，创建一个新的区块链")

			// 创建交易
			newTransaction := transaction.NewCoinBaseTx(address, genesisData)

			// 创建一个创世区块
			genesis := NewGensisBlock([]*transaction.Transaction{newTransaction})
			// 创建一个桶
			bucket, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				log.Panic(err)
			}

			// 将创世区块放入当前桶中，key：创世区块的Hash, value：创世区块的序列化
			err = bucket.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			// 将创世区块的Hash更新到key=l，数据库中l对应了最新的区块Hash
			err = bucket.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}

			// 最近区块就是创世区块
			tip = genesis.Hash
		} else {
			// 获取当前数据库中最新的区块hash
			tip = bucket.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	// 将区块链的UTXO写入数据库
	utxoSet := UTXOSet{&bc}
	utxoSet.Reindex()
	return &bc
}

/*
	summary：构建新的交易（转账）
	from: 转出地址
	to: 转入地址
	amount: 转账金额
	bc: 操作所属的区块链
	return: &Transaction 新的交易对象地址
*/
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *transaction.Transaction {
	var inputs []transaction.TXInput
	var outputs []transaction.TXOutput

	// 读取当前钱包数据
	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}

	// 根据钱包读取转出地址对应的公钥
	newWalet := wallets.GetWallet(from)

	// 将钱包公钥进行Hash得到Pubkey Hash
	fromPubkeyHash := wallet.HashPubKey(newWalet.PublicKey)

	// 根据转账地址和待转账金额获取能够转账的金额和相应的有效的输出
	total, validaoutputs := bc.FindSpendableOutputs(fromPubkeyHash, amount)
	if total < amount {
		log.Panic("当前地址的金额小于待转账金额，转账失败！")
	}

	// 循环遍历有效的输出
	for txId, outs := range validaoutputs {
		// 将交易ID由string转为16进制
		txID, err := hex.DecodeString(txId)
		if err != nil {
			log.Panic(err)
		}

		// 循环遍历交易的输出
		for _, out := range outs {
			// 将有效的输出作为转账的输入, 添加到转账的输入集合
			input := transaction.TXInput{txID, out, nil, newWalet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// 将待转入的金额和地址作为交易的输出
	outputs = append(outputs, *transaction.NewTXOutput(amount, to))

	// 如果当前地址的可用金额大于待转账金额（零钱）, 则将多余的金额转回自己的地址, 并记录到当前交易的输出
	if total > amount {
		outputs = append(outputs, *transaction.NewTXOutput(total - amount, from))
	}

	// 构建交易对象
	tx := transaction.Transaction{nil, inputs, outputs}
	// 当前交易的Hash作为交易的ID
	tx.ID = tx.Hash()

	// 根据私钥对交易进行数据签名
	bc.SignTransaction(&tx, newWalet.PrivateKey)
	return &tx
}