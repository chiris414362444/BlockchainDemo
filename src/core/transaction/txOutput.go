package transaction

import (
	"bytes"
	"core/algorithm"
)

// 交易输出结构体
type TXOutput struct {
	Value int
	PublicKeyHash []byte  // 公钥Hash
}

// 通过地址得到公钥的Hash
func (out *TXOutput) GetPubkeyHash(address []byte) {
	decodeAddress := algorithm.Base58Decode(address)
	pubkeyHash := decodeAddress[1 : len(decodeAddress) - 4]
	out.PublicKeyHash = pubkeyHash
}

// 判断交易输出是否属于地址
func (out *TXOutput) CanBeUnlockedWith(pubkeyHash []byte) bool {
	return bytes.Compare(out.PublicKeyHash, pubkeyHash)  == 0
}

// 根据金额和地址，构建一个输出
func NewTXOutput(value int, address string) *TXOutput {
	txo := TXOutput{value, nil}
	txo.GetPubkeyHash([]byte(address))
	return &txo
}