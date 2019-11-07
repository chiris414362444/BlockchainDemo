package main

import (
	"bytes"
	"core/blockchain"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"utils"
)

func main() {
	//test.TestPow()
	//test.TestNewSerialize()
	//test.TestNewGensisBlock()
	//test.TestBoltDB()
	//test.TestWallet()

	bc := blockchain.NewBlockchain("1FdsuGae3QNWcJLg2yKNQ1vZkZ5Cdg3KUm")
	cli := CLI{bc}
	cli.Run()
}

func main2() {

	fmt.Println("----------------验证当前区块的Hash值是否正确------------------")
	// 当前区块版本号
	var version int32 = 2
	versionHex := utils.IntToHex(version, true)
	fmt.Printf("versionHex: %x\n", versionHex)

	// 前一个区块的hash
	prevHash, _ := hex.DecodeString("000000000000000016145aa12fa7e81a304c38aec3d7c5208f1d33b587f966a6")
	utils.ReverseBytes(prevHash)
	fmt.Printf("prevHash: %x\n", prevHash)

	// 默克尔根
	merkleRootHash, _ := hex.DecodeString("3a4f410269fcc4c7885770bc8841ce6781f15dd304ae5d2770fc93a21dbd70d7")
	utils.ReverseBytes(merkleRootHash)
	fmt.Printf("merkleRootHash: %x\n", merkleRootHash)

	// 当前时间
	var time int32 = 1418755780
	timeHex := utils.IntToHex(time, true)
	fmt.Printf("time: %x\n", timeHex)

	// 当前区块难度
	var bits int32 = 404454260
	bitsHex := utils.IntToHex(bits, true)
	fmt.Printf("bits: %x\n", bitsHex)

	// 随机数
	var nonce int32 = 1865996595
	nonceHex := utils.IntToHex(nonce, true)
	fmt.Printf("nonce: %x\n", nonceHex)

	// 将当前区块中所有信息拼接
	tempResult := bytes.Join([][]byte{ versionHex, prevHash, merkleRootHash, timeHex, bitsHex, nonceHex }, []byte{})
	fmt.Printf("tempResult: %x\n", tempResult)

	// 对当前区块进行序列化（双重hash）
	firstHash := sha256.Sum256(tempResult)
	secodeHash := sha256.Sum256(firstHash[:])
	currentBlockHash := secodeHash[:]
	utils.ReverseBytes(currentBlockHash)
	fmt.Printf("计算当前区块的Hash值: %x\n", currentBlockHash)

	fmt.Println("----------------计算当前区块的目标值------------------")
	// 计算当前区块的目标值
	targetHash := utils.CalculateTargetFast(bits)
	fmt.Printf("计算当前区块的目标值: %x\n", targetHash)

	fmt.Println("----------------模拟挖矿，计算当前区块的Nonce------------------")
	// 初始化区块对象，nonce=0
	block := blockchain.Block{
		Version:       2,
		PrevBlockHash: prevHash,
		MerkleRoot:    merkleRootHash,
		Hash:          []byte{},
		Time:          time,
		Bits:          bits,
		Nonce:         0,
	}

	var target, tempHash big.Int
	// 将目标hash值转为big.Int
	target.SetBytes(targetHash)

	//core.nonce = 1865996590
	// 循环计算nonce已满足目标hash
	for block.Nonce < math.MaxInt32 {
		// 对当前区块进行序列化
		data := block.Serialize()
		//双重Hash计算当前Nonce下区块的hash值
		firstHash = sha256.Sum256(data)
		secodeHash = sha256.Sum256(firstHash[:])
		utils.ReverseBytes(secodeHash[:])
		tempHash.SetBytes(secodeHash[:])
		fmt.Printf("当前Nonce值是：%d；对应的hash值是：%x\n", block.Nonce, secodeHash)

		// 比较当前区块hash与目标hash，若下雨目标hash则满足退出循环
		if tempHash.Cmp(&target) == -1 {
			break
		}

		block.Nonce++
	}
}