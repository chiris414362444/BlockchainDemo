/*
  测试文件，存放测试方法
*/
package test

import (
	"core/blockchain"
	"core/transaction"
	"core/wallet"
	"fmt"
)

// 测试创建区块的默克尔根
func TestCreateMerkleTreeRoot() {
	// 初始化区块对象
	block := blockchain.Block{
		Version:       2,
		PrevBlockHash: []byte{},
		MerkleRoot:    []byte{},
		Hash:          []byte{},
		Time:          1418755780,
		Bits:          404454260,
		Nonce:         0,
		Height:         0,
	}

	txIn1 := transaction.TXInput{[]byte{}, -1, nil, nil}
	txOut1 := transaction.NewTXOutput(transaction.Subsidy, "first")
	tx1 := transaction.Transaction{nil, []transaction.TXInput{txIn1 }, []transaction.TXOutput{*txOut1 }}

	txIn2 := transaction.TXInput{[]byte{}, -1, nil,nil}
	txOut2 := transaction.NewTXOutput(100, "second")
	tx2 := transaction.Transaction{nil, []transaction.TXInput{txIn2 }, []transaction.TXOutput{*txOut2 }}

	var transactions []*transaction.Transaction
	transactions = append(transactions, &tx1, &tx2)
	block.CreateMerkleTreeRoot(transactions)
	fmt.Printf("当前测试区块的默克尔根：%x", block.MerkleRoot)
}

// 测试挖矿
func TestPow() {
	// 初始化区块
	block := &blockchain.Block{
		Version:       2,
		PrevBlockHash: []byte{},
		MerkleRoot:    []byte{},
		Hash:          []byte{},
		Time:          1418755780,
		Bits:          404454260,
		Nonce:         0,
		Height:         0,
		Transactions:  []*transaction.Transaction{},
	}

	// 初始化POW
	pow := blockchain.NewProofOfWork(block)
	// 开始挖矿, 并计算nonce
	nonce, _ := pow.Mine()
	block.Nonce = nonce
	fmt.Println("当前Nonce：", nonce)
	fmt.Println("POW验证是否成功：", pow.Validate())
}

// 区块序列化测试
func TestNewSerialize() {
	// 初始化区块
	block := &blockchain.Block{
		Version:       2,
		PrevBlockHash: []byte{},
		MerkleRoot:    []byte{},
		Hash:          []byte{},
		Time:          1418755780,
		Bits:          404454260,
		Nonce:         0,
		Height:         0,
		Transactions:  []*transaction.Transaction{},
	}

	// 序列化区块
	result := block.Serialize()
	fmt.Printf("当前区块的序列化值：%x\n", result)

	// 反序列化
	blockData := blockchain.DeserializeBlock(result)
	blockData.String()
}

func TestNewGensisBlock() {
	blockchain.NewGensisBlock([]*transaction.Transaction{})
}

// 测试区块链存储数据库
func TestBoltDB() {
	blockchain := blockchain.NewBlockchain("1FdsuGae3QNWcJLg2yKNQ1vZkZ5Cdg3KUm")
	blockchain.MineBlock([]*transaction.Transaction{})
	blockchain.MineBlock([]*transaction.Transaction{})
	blockchain.PrintBlockchain()
}

// 验证钱包功能
func TestWallet() {
	newWallet := wallet.NewWallet()
	address := newWallet.GetAddress()

	fmt.Printf("钱包私钥: %x\n", newWallet.PrivateKey.D.Bytes())
	fmt.Printf("钱包公钥: %x\n", newWallet.PublicKey)
	fmt.Printf("钱包地址: %x\n", address)
	fmt.Printf("地址是否有效: %d\n", wallet.ValidateAddress(address))
}