package server

// 发送区块信息的清单结构体
type Inventory struct {
	AddrFrom string  // 请求的地址
	Type string  // 类型
	AllBlocksHash [][]byte  // 区块链所有区块的Hash
}

// 请求区块的结构体
type GetData struct {
	AddrFrom string  // 请求的地址
	Type string  // 类型
	BlockHash []byte  // 区块的Hash
}

// 发送区块信息的结构体
type SendBlock struct {
	AddrFrom string  // 发往的地址
	Block []byte  // 区块的序列化
}