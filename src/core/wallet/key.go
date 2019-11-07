package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

// 根据椭圆曲线生成私钥和公钥
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	// 生成椭圆曲线, secp256r1: go语言内置曲线; secp256k1: 比特币中的曲线
	curve := elliptic.P256()

	// 生成私钥结构体
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	// 生成公钥，公钥是曲线上的x点和y点拼接在一起
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey
}

// 对公钥进行ripemd160, 得到Pubkey Hash
func HashPubKey(pubkey []byte) []byte {
	// 对公钥进行SHA256(PubKey)
	pubkeyHash256 := sha256.Sum256(pubkey)
	PIPEMD160Hash := ripemd160.New()

	_, err := PIPEMD160Hash.Write(pubkeyHash256[:])
	if err != nil {
		log.Panic(err)
	}

	// 对公钥进行RIPEMD160(SHA256(PubKey)) 计算public key hash
	pubkeyRIPEMD160 := PIPEMD160Hash.Sum(nil)
	return pubkeyRIPEMD160
}

// 计算检查值
func checkSum(payload []byte) []byte {
	// 进行双重hash计算检查值
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	// 检查值，比特币中的检查值是前4个字节
	checkSum := secondSHA[:4]
	return checkSum
}