package server

import (
	"bytes"
	"core/blockchain"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

// 存放当前节点已有的区块的Hash
var blockInTransit [][]byte

// 处理请求连接var
func handleConnection(conn net.Conn, bc *blockchain.Blockchain) {
	// 读取请求发送的数据
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	// 获取指令
	command := bytesToCommand(request[:commandLength])

	// 根据命令处理不同的逻辑
	switch command {
	case "version":
		handleVersion(request, bc)
	case "inventory":
		handleInventory(request, bc)
	case "getblockchain":
		handleGetBlockchain(request, bc)
	case "getblock":
		handleGetBlock(request, bc)
	case "sendblock":
		handleSendBlock(request, bc)
	}
}

// 处理接收到的区块版本信息
func handleVersion(request []byte, bc *blockchain.Blockchain) {
	var payload Version
	var buff bytes.Buffer
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// 打印接受到的区块链版本信息
	payload.String()

	// 当前节点的区块链高度
	myBestHeight := bc.GetBestHeight()

	// 外部节点的区块链高度
	foreignerBestHeight := payload.BestHeight

	// 当前节点小于外部节点, 则发送获取区块的请求
	if myBestHeight < foreignerBestHeight {
		sendGetBlockChain(payload.AddrFrom)
	} else { // 当前节点大于或等于外部节点的区块链高度, 则只需要向外部节点发送当前节点的区块链版本信息
		sendVersion(payload.AddrFrom, bc)
	}

	// 判断当前外部节点是否已在已知节点集合里, 不在则加入已知节点
	if !nodeIsKnow(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

// 处理接收到的区块清单
func handleInventory(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload Inventory
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("接收到区块清单, 版本: %s, 区块个数: %d\n", payload.Type, len(payload.AllBlocksHash))

	if payload.Type == "block" {
		blockInTransit = payload.AllBlocksHash

		// 0号Hash表示外部节点的最后一个区块的Hash, 即最新的区块（区块的遍历是从后往前遍历）
		blockHash := payload.AllBlocksHash[0]

		// 发送数据获取最新的区块
		senGetBlockData(payload.AddrFrom, "block", blockHash)

		// 循环遍历当前节点已有的区块
		newInTransit := [][]byte{}
		for _, b := range blockInTransit {
			// 从当前已有的区块的Hash集合中，剔除已获取的最新区块
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}

		// 更新当前节点已有区块Hash集合
		blockInTransit = newInTransit
	}
}

// 处理发送区块链信息的方法
func handleGetBlockchain(request []byte, bc *blockchain.Blockchain) {
	var payload string
	var buff bytes.Buffer
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// 获取当前区块链所有区块的Hash
	allBlocksHash := bc.GetAllBlockHash()
	sendInventory(payload, "block", allBlocksHash)
}

// 处理发送区块信息的方法
func handleGetBlock(request []byte, bc *blockchain.Blockchain) {
	var payload GetData
	var buff bytes.Buffer
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// 根据区块的Hash获取区块信息
	if payload.Type == "block" {
		block, err := bc.GetBlockById([]byte(payload.BlockHash))
		if err != nil {
			log.Panic(err)
		}

		// 向请求节点发送区块信息
		sendBlock(payload.AddrFrom, &block)
	}
}

// 处理外部节点发送的区块
func handleSendBlock(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload SendBlock
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// 外部节点发送的区块的序列号化
	blockData := payload.Block

	// 反序列化得到区块
	block := blockchain.DeserializeBlock(blockData)

	// 将区块加入当前区块链
	bc.AddBlock(block)
	fmt.Printf("已接收到区块: %x\n", block.Hash)

	// 判断当前存储的已有的区块是否>0
	if len(blockInTransit) > 0 {
		// 0号Hash表示节点的最后一个区块的Hash, 即最新的区块（区块的遍历是从后往前遍历）
		blockHash := blockInTransit[0]

		// 继续发送数据获取最新的区块
		senGetBlockData(payload.AddrFrom, "block", blockHash)

		// 更新当前节点已有区块Hash集合, 剔除已获取的区块
		blockInTransit = blockInTransit[1:]
	} else {
		// 当前节点已获取全部区块, 则更新UTXO
		set := blockchain.NewUTXOSet(bc)
		set.Reindex()
	}
}