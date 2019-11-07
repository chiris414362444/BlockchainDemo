/*
  工具包，存放常用的工具方法
*/
package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
)

// 计算两个数的最小值M
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// 将整数转为16进制字节数组
func IntToHex(num int32, isLittle bool) []byte {
	var err error
	buff := new(bytes.Buffer)
	// littkeEndian小端模式排序
	if isLittle {
		err = binary.Write(buff, binary.LittleEndian, num)
	} else { // BigEndian大端模式排序
		err = binary.Write(buff, binary.BigEndian, num)
	}
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// 字节翻转
func ReverseBytes(data []byte) {
	for i, j := 0, len(data) - 1; i < j; i, j = i + 1, j - 1 {
		data[i], data[j] = data[j], data[i]
	}
}

// 计算目标值，当前区块的hash值必须小于目标值
func CalculateTargetFast(nBits int32) []byte {
	// 大端
	bits := IntToHex(nBits, false)

	// Bits难度，第一个字节表示指数
	exponent := bits[:1]
	fmt.Printf("%x\n", exponent)

	// Bits后面3个字节表示系数
	coeffient := bits[1:]
	fmt.Printf("%x\n", coeffient)

	// 将字节数组转为int，可先将字节数组转为string，再将string转为int
	// 将字节的16进制为0x18，转为了string"18"
	strExponent := hex.EncodeToString(exponent)
	// 将字符串"18"转为了10进制int64 24
	nExp, _ := strconv.ParseInt(strExponent, 16, 8)

	// 在系数前面加上32-nExp个0
	result := append(bytes.Repeat([]byte{0x00}, 32-int(nExp)), coeffient...)

	// 为了让result保持32位，若不足32位后面补0
	result = append(result, bytes.Repeat([]byte{0x00}, 32-len(result))...)
	return result
}

// 序列化数据
func EncodeData(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}