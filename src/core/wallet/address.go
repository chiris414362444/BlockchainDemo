package wallet

import (
	"bytes"
	"core/algorithm"
)

// 计算钱包地址
func (w *Wallet) GetAddress() []byte {
	// 对公钥取Hash
	pubkeyHash := HashPubKey(w.PublicKey)

	// 拼接版本号
	versionPayload := append([]byte{version}, pubkeyHash...)

	// 计算检查值
	checksum := checkSum(versionPayload)

	// 拼接检查值
	fullPayload := append(versionPayload, checksum...)

	// 对完整的地址“Version + Public Key Hash + CheckSum”进行Base58编码得到比特币地址
	address := algorithm.Base58Encode(fullPayload)
	return address
}

// 验证地址是否有效
func ValidateAddress(address []byte) bool {
	// Base58解码地址得到public hash
	pubkeyHash := algorithm.Base58Decode(address)

	// 获取真实检查值（public hash的后4个字节)）
	actualChecksum := pubkeyHash[len(pubkeyHash) - 4 : ]

	// 获取除去版本1个字节和检查值后4个字节的中间部分的Hash
	centerPubkeyHash := pubkeyHash[1 : len(pubkeyHash) - 4]

	// 将版本号+中间部分，计算检查值
	targetChecksum := checkSum(append([]byte{version}, centerPubkeyHash...))

	// 比较真实检查值和计算得到的检查值是否相等
	return bytes.Compare(actualChecksum, targetChecksum)  == 0
}