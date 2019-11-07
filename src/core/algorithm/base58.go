package algorithm

import (
	"bytes"
	"math/big"
	"utils"
)

// base58编码
var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func Base58Encode(input []byte) []byte {
	// 定义一个字节切片作为返回值
	var result []byte

	// 把字节数组转为大整数big.Int
	nInput := big.NewInt(0).SetBytes(input)

	// 把base58的长度58转为大整数
	nBaseCount := big.NewInt(int64(len(base58Alphabet)))

	// 0的大整数
	nZero:= big.NewInt(0)

	// 大整数的指针
	modePtr := &big.Int{}

	// 循环不停的对nInput取余，直到除数为0停止
	for nInput.Cmp(nZero) != 0 {
		// 对nInput除以58后取余数，并将除数赋给nInput
		nInput.DivMod(nInput, nBaseCount, modePtr)
		result = append(result, base58Alphabet[modePtr.Int64()])
	}

	// 翻转字节数组
	utils.ReverseBytes(result)

	// 如果字节数组的前面为字节0， 会把它替换为1（因为BTC的特殊处理）
	for _, b := range input {
		if b == 0x00 {
			result = append([]byte { base58Alphabet[0] }, result...)
		} else {
			break
		}
	}

	return result
}

func Base58Decode(input []byte) []byte {
	// 解码结果的大整数值
	result := big.NewInt(0)

	// 标识Base58编码后前面有多少个1，相应的解码后表示有多少个0
	zeroBytes := 0

	// 循环遍历前面有多少个1，有多少个1即对应解码后的0
	for _, b := range input {
		if b == '1' {
			zeroBytes++
		} else {
			break
		}
	}

	// 除去Base58编码后前面的1
	realData := input[zeroBytes:]

	// 循环逆推结果
	for _,b := range realData {
		// 通过字符在Base58中的位置，反推出余数
		charIndex := bytes.IndexByte(base58Alphabet, b)

		// 当前值乘以58， eg: 123可分解为 :0*10+1=1, 1*10+2=12, 12*10+3=123
		result.Mul(result, big.NewInt(58))
		// 当前值加上余数
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	// 将解码后的大整数转为byte数组
	decode := result.Bytes()

	// 将解码后的byte数组前面加上0
	decode = append(bytes.Repeat([]byte { 0x00 }, zeroBytes), decode...)
	return decode
}