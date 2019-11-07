package server

import (
	"core/blockchain"
	"fmt"
	"log"
	"net"
)

// 请求命令长度
const commandLength = 20

// 节点版本
const nodeVersion = 0x00

// 本地节点地址
var nodeAddress string

// 已知节点
var knownNodes = []string{"localhost:3000"}

// 开启服务器
func StrartServer(nodeId, minerAddress string, bc *blockchain.Blockchain) {
	// 当前节点地址
	nodeAddress = fmt.Sprintf("localhost:%s", nodeId)

	// 监听节点地址, 参数1: 网络协议, 参数2: 监听地址
	listener, err := net.Listen("tcp", nodeAddress)
	if err != nil {
		log.Panic(err)
	}

	defer listener.Close()

	// 如果区块链不存在构建区块链对象
	if bc == nil {
		bc = blockchain.NewBlockchain("1FdsuGae3QNWcJLg2yKNQ1vZkZ5Cdg3KUm")
	}

	if nodeAddress != knownNodes[0] {
		// 向外部节点发送当前节点的区块链版本信息
		sendVersion(knownNodes[0], bc)
	}

	for {
		// 接收所监听地址返回的连接
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}

		defer conn.Close()

		// 处理连接
		go handleConnection(conn, bc)
	}
}