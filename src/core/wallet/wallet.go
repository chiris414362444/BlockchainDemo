package wallet

import (
	"crypto/ecdsa"
)

// 版本（比特币主网的版本号为0，占1个字节）
const version = byte(0x00)

// 钱包对象
type Wallet struct {
	PrivateKey ecdsa.PrivateKey  // 私钥
	PublicKey []byte  // 公钥
}

// 新构建钱包
func NewWallet() *Wallet {
	privateKey, publicKey := NewKeyPair()
	walllet := Wallet{privateKey, publicKey}
	return &walllet
}