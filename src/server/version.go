package server

import "fmt"

type Version struct {
	Version int32
	BestHeight int32
	AddrFrom string  // 发送地址
}

func (ver *Version) String() {
	fmt.Printf("当前区块的版本是: %d\n", ver.Version)
	fmt.Printf("当前区块高度是: %d\n", ver.BestHeight)
	fmt.Printf("发送当前信息的地址是: %s\n", ver.AddrFrom)
}