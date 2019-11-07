package server

import (
	"bytes"
	"core/blockchain"
	"fmt"
	"io"
	"log"
	"net"
	"utils"
)

// 发送区块链的版本信息
func sendVersion(address string, bc *blockchain.Blockchain) {
	bestHeight := bc.GetBestHeight()

	// 构建待发送的版本信息结构体
	version := Version{nodeVersion, bestHeight, nodeAddress}

	// 序列化版本结构体
	payload := utils.EncodeData(version)

	// 构建请求（前面的version表示命令 ）
	request := append(commandToBytes("version"), payload...)
	sendData(address, request)
}

// 发送请求获取区块链
func sendGetBlockChain(address string)  {
	// 序列化获取区块链请求的地址
	payload := utils.EncodeData(nodeAddress)

	// 构建请求（前面的getblocks表示命令 ）
	request := append(commandToBytes("getblockchain"), payload...)
	sendData(address, request)
}

// 发送请求获取区块数据
func senGetBlockData(address, kind string, blockHash []byte) {
	payload := utils.EncodeData(GetData {nodeAddress, kind, blockHash})
	request := append(commandToBytes("getblock"), payload...)
	sendData(address, request)
}

// 发送区块链清单
func sendInventory(address string, kind string, allBlocksHash [][]byte) {
	inventory := Inventory {nodeAddress, kind, allBlocksHash}
	payload := utils.EncodeData(inventory)
	request := append(commandToBytes("inventory"), payload...)
	sendData(address, request)
}

// 发送区块信息
func sendBlock(address string, block *blockchain.Block) {
	data := SendBlock{nodeAddress, block.Serialize()}
	payload := utils.EncodeData(data)
	request := append(commandToBytes("sendblock"), payload...)
	sendData(address, request)
}

// 根据地址发送数据
func sendData(address string, data []byte) {
	// 与address建立连接
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("%s 地址不可用！\n", address)

		// 当前地址不可用更新节点地址
		var updateNodeAddress []string
		for _, node := range knownNodes {
			if node != address {
				updateNodeAddress = append(updateNodeAddress, node)
			}
		}

		knownNodes = updateNodeAddress
	}

	defer conn.Close()

	// 往地址连接传递数据, 将资源发送到通道
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}