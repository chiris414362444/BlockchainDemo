/*
  区块链工作量证明算法包
*/
package blockchain

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"utils"
)

// 挖矿难度
const targetBits = 16

type ProofOfWork struct {
	block *Block
	target *big.Int  // 目标值
}

func NewProofOfWork(block *Block) *ProofOfWork {
	// 用大整数标识目标值
	target := big.NewInt(1)
	// 将目标值左移相应位数，标识挖矿难度
	target.Lsh(target, uint(256 - targetBits))
	//fmt.Printf("当前目标值：%x\n", target.Bytes())

	// 根据区块和当前区块挖矿难度初始化POW
	pow := &ProofOfWork{block, target}
	return pow
}

// 序列化当前区块
func (pow *ProofOfWork) Serialize(nonce int32) []byte {
	data := bytes.Join([][]byte{
		utils.IntToHex(pow.block.Version, true),
		pow.block.PrevBlockHash,
		pow.block.MerkleRoot,
		utils.IntToHex(pow.block.Time, true),
		utils.IntToHex(pow.block.Bits, true),
		utils.IntToHex(nonce, true),
	}, []byte{})

	return data
}

// 挖矿函数
func (pow *ProofOfWork) Mine() (int32, []byte) {
	var nonce int32 = 0
	var currentHash [32]byte
	var bIntCurrent big.Int

	for {
		// 序列化当前区块
		data := pow.Serialize(nonce)
		// double hash
		firstHash := sha256.Sum256(data)
		currentHash = sha256.Sum256(firstHash[:])

		// 打印当前挖矿过程
		//fmt.Printf("%x\n", currentHash)

		// 将当前hash转为大整型
		bIntCurrent.SetBytes(currentHash[:])

		// 和目标值比较, 小于当前POW的目标值则挖矿成功
		if bIntCurrent.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return nonce, currentHash[:]
}

// 验证是否小于当前目标值
func (pow *ProofOfWork) Validate() bool {
	var hasInt big.Int
	data := pow.Serialize(pow.block.Nonce)

	// double hash
	firstHash := sha256.Sum256(data)
	secondHash := sha256.Sum256(firstHash[:])
	hasInt.SetBytes(secondHash[:])

	// 验证当前hash值是否小于目标值
	isValidate := hasInt.Cmp(pow.target) == -1
	return isValidate
}

