package server

import "fmt"

// 判断当前外部节点是否已在已知节点集合里
func nodeIsKnow(address string) bool {
	for _, node := range knownNodes {
		if node == address {
			return true
		}
	}

	return false
}

// string命令转为字节数组
func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

// 字节数组转为string命令
func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x00 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}
