/*
  交易包，存放区块链交易类及其方法函数
*/
package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const Subsidy = 100

// 交易结构体
type Transaction struct {
	ID []byte   // 交易的Hash
	Vin []TXInput
	Vout []TXOutput
}

// 计算交易的hash值，即计算交易的ID
func (tx *Transaction) Hash() []byte {
	txcopy := *tx
	txcopy.ID = []byte{}
	hash := sha256.Sum256(txcopy.Seialize())
	return hash[:]
}

// 根据私钥对交易进行数据签名
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	// CoinBase交易不用处理签名
	if tx.IsCoinBase() {
		return
	}

	// 构建交易副本
	txCopy := tx.CopyTransaction()

	// 循环遍历交易的输入
	for inID, vin := range txCopy.Vin {
		// 将这笔输入的ID转为string
		vinId := hex.EncodeToString(vin.TXid)

		// 获取前一比交易的结构体
		prevTx := prevTXs[vinId]
		if prevTx.ID == nil {
			log.Panic(fmt.Sprintf("未找到输入ID: %s, 所在的交易！", vinId))
		}

		// 将这笔交易的这笔输入的签名置为nil
		txCopy.Vin[inID].Signature = nil

		// 这笔交易的这笔输入引用的前一笔交易的输出的公钥哈希
		txCopy.Vin[inID].Pubkey = prevTx.Vout[vin.VoutIndex].PublicKeyHash

		// 设置交易副本的ID为交易的Hash
		txCopy.ID = txCopy.Hash()

		// 数据签名得到椭圆曲线的r和s
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}

		// 交易的数据签名是由 r + s拼接而成
		signature := append(r.Bytes(), s.Bytes()...)

		// 将数据签名赋给真实的交易的输入
		tx.Vin[inID].Signature = signature
	}
}

// 验证交易是否有效
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	// CoinBase 交易不需要验证
	if tx.IsCoinBase() {
		return true
	}

	// 构建交易副本
	txCopy := tx.CopyTransaction()

	// 构建一个椭圆曲线
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		// 将这笔输入的ID转为string
		vinId := hex.EncodeToString(vin.TXid)

		// 获取前一比交易的结构体
		prevTx := prevTXs[vinId]
		if prevTx.ID == nil {
			log.Panic(fmt.Sprintf("未找到输入ID: %s, 所在的交易！", vinId))
		}

		// 将这笔交易的这笔输入的签名置为nil
		txCopy.Vin[inID].Signature = nil

		// 这笔交易的这笔输入引用的前一笔交易的输出的公钥哈希
		txCopy.Vin[inID].Pubkey = prevTx.Vout[vin.VoutIndex].PublicKeyHash

		// 设置交易副本的ID为交易的Hash
		txCopy.ID = txCopy.Hash()

		// 用大整数表示椭圆曲线的r和s
		r := big.Int{}
		s := big.Int{}

		// 将输入的数据签名一分为二, 分别表示r和s点, 并将[]byte转为大整数
		signHalfLen := len(vin.Signature) / 2
		r.SetBytes(vin.Signature[: signHalfLen])
		s.SetBytes(vin.Signature[signHalfLen : ])

		// 计算公钥的大整数
		x := big.Int{}
		y := big.Int{}

		// 将输入的公钥一分为二, 分别表示x和y点, 并将[]byte转为大整数
		pubkeyHalfLen := len(vin.Pubkey) / 2
		x.SetBytes(vin.Pubkey[: pubkeyHalfLen])
		y.SetBytes(vin.Pubkey[pubkeyHalfLen : ])

		// 根据椭圆曲线和x、y轴所在的点构建公钥
		rawPubkey := ecdsa.PublicKey{curve, &x, &y }

		// 根据公钥、交易hash, 通过椭圆曲线验证数据签名的r和s点是否有效
		if ecdsa.Verify(&rawPubkey, txCopy.ID, &r, &s) == false {
			return false
		}

		// 将交易副本的ID置为空, 以免影响下次循环对交易Hash的计算
		txCopy.Vin[inID].Pubkey = nil
	}

	return true
}

// 构建交易副本
func (tx *Transaction) CopyTransaction() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.TXid, vin.VoutIndex, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PublicKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

// 交易序列化
func (tx Transaction) Seialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// 标准化打印
func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--- Transaction %x", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("    Input        %d", i))
		lines = append(lines, fmt.Sprintf("    TXID:        %d:", input.TXid))
		lines = append(lines, fmt.Sprintf("    Out:         %d:", input.VoutIndex))
		lines = append(lines, fmt.Sprintf("    Signature:   %x:", input.Signature))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("    Output       %d", i))
		lines = append(lines, fmt.Sprintf("    Value:       %d:", output.Value))
		lines = append(lines, fmt.Sprintf("    Script:      %x:", output.PublicKeyHash))
	}

	return strings.Join(lines, "\n")
}

// 判断是否是区块的第一笔交易
func (tx Transaction) IsCoinBase() bool {
	// 区块的第一笔交易只有一个输入, 且第一笔输入的id为空, 且引用的输出为-1
	return len(tx.Vin) == 1 && len(tx.Vin[0].TXid) == 0 && tx.Vin[0].VoutIndex == -1
}

// 构建第一笔coinbase交易
func NewCoinBaseTx(to, data string) *Transaction {
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(Subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}