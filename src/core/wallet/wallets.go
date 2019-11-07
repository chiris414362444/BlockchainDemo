package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// 存放钱包的文件
const walletFile = "wallet.dat"

// 存储钱包的集合的对象
type Wallets struct {
	WalletStore map[string]*Wallet  // key: 钱包地址  value:钱包
}

// 创建钱包
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.WalletStore[address] = wallet
	return address
}

// 根据地址获取钱包
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.WalletStore[address]
}

// 获取所有钱包地址
func (ws *Wallets) GetAddress() []string {
	var addresses []string
	for address, _ := range ws.WalletStore {
		addresses = append(addresses, address)
	}

	return addresses
}

// 将钱包存放至文件存储
func (ws *Wallets) SaveToFile() {
	// 申明一个缓冲区存放待存储内容
	var content bytes.Buffer

	// 若要序列化接口, 需要先注册接口, 这里需要序列化椭圆曲线的P256接口
	gob.Register(elliptic.P256())

	// 申明序列化对象
	encoder := gob.NewEncoder(&content)
	// 序列化钱包
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	// 将序列化的钱包写入文件 0777 代表文件操作权限
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0777)
	if err != nil {
		log.Panic(err)
	}
}

// 从文件中读取钱包信息到结构体
func (ws *Wallets) LoadFromFile() error {
	// 判断文件是否存在, 不存在则返回
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	// 读取文件
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	if string(fileContent) == "" { // 文件为空直接返回
		return nil
	}

	var wallets Wallets
	// 注册反序列化接口
	gob.Register(elliptic.P256())
	fileReader := bytes.NewReader(fileContent)
	decoder := gob.NewDecoder(fileReader)
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.WalletStore = wallets.WalletStore
	return nil
}

// 构建Wallets对象
func NewWallets() (*Wallets, error) {
	// 构建一个空Wallets对象
	wallets := Wallets{}
	wallets.WalletStore = make(map[string]*Wallet)

	// 读取文件内容到Wallets结构体对象
	err := wallets.LoadFromFile()
	return &wallets, err
}