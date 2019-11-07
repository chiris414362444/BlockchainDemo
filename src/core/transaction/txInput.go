package transaction

import (
	"bytes"
	"core/wallet"
)

// 交易输入结构体
type TXInput struct {
	TXid []byte
	VoutIndex int
	Signature []byte  // 数据签名
	Pubkey []byte  // 公钥
}

// 判读输入是否属于地址
func (in *TXInput) CanUnlockOutputWith(unlockData []byte) bool {
	lockingHash := wallet.HashPubKey(in.Pubkey)
	return bytes.Compare(lockingHash, unlockData)  == 0
}