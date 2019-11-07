package main

import (
	"core/algorithm"
	"core/blockchain"
	"core/transaction"
	"core/wallet"
	"flag"
	"fmt"
	"log"
	"os"
	"server"
)

type CLI struct {
	bc *blockchain.Blockchain
}

// 验证参数
func (cli *CLI) validateArgs() {
	// 参数小于1, 程序退出
	if len(os.Args) < 1 {
		fmt.Printf("参数不合法")
		os.Exit(1)
	}

	fmt.Println(os.Args)
}

// 命令使用说明
func (cli *CLI) printUsage() {
	fmt.Println("使用说明")
	fmt.Println("输入addblock, 增加区块")
	fmt.Println("输入printChain, 打印区块链")
	fmt.Println("输入getbalance, 查询地址的金额")

}

// 获取地址的金额
func (cli *CLI) getBalance(address string) int {
	balance := 0

	// 根据地址转为Pubkey Hash
	decodeHash := algorithm.Base58Decode([]byte(address))
	pubkeyHash := decodeHash[1 : len(decodeHash) - 4]

	set := blockchain.NewUTXOSet(cli.bc)
	UTXOs := set.FindUTXOByPubkeyHash(pubkeyHash)
	//UTXOs := cli.bc.FindUTXO(pubkeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("地址：%s， 拥有金额：%d\n", address, balance)
	return balance
}

// 转账
func (cli *CLI) send (from, to string, amount int) {
	// 构建交易
	tx := blockchain.NewUTXOTransaction(from, to, amount, cli.bc)
	// 将当前交易记录区块链
	newbBlock := cli.bc.MineBlock([]*transaction.Transaction{tx})

	// 根据区块更新UTXO
	set := blockchain.NewUTXOSet(cli.bc)
	set.UpdateUTXOByBlock(newbBlock)
	fmt.Println("转账成功！")
}

// 创建钱包, 并存储到文件中
func (cli *CLI) createWallet() {
	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}

	address := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("你的钱包地址是：%s\n", address)
}

func (cli *CLI) listAddress() {
	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}

	addresses := wallets.GetAddress()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) startNode(nodeId, minnerAddress string) {
	fmt.Printf("开始运行节点：%s\n", nodeId)
	if len(minnerAddress) >0 {
		if wallet.ValidateAddress([]byte(minnerAddress)) {
			fmt.Printf("%s 矿工正在运行\n", minnerAddress)
		} else {
			log.Panic("矿工地址不合法")
		}
	}

	// 运行服务
	server.StrartServer(nodeId, minnerAddress, cli.bc)
}

func (cli *CLI) Run() {
	// 验证参数
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalace", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressCmd := flag.NewFlagSet("listaddress", flag.ExitOnError)
	getBestHeightCmd := flag.NewFlagSet("getbestheight", flag.ExitOnError)

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "请输入转账的转出地址")
	sendTo := sendCmd.String("to", "", "请输入转账的转入地址")
	sendAmount := sendCmd.Int("amount", 0, "请输入转账的金额")

	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	startNodeMinner := startNodeCmd.String("minner", "", "请输入矿工的地址")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbestheight":
		err := getBestHeightCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddress":
		err := listAddressCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		cli.bc.MineBlock([]*transaction.Transaction{})
	}

	if printChainCmd.Parsed() {
		cli.bc.PrintBlockchain()
	}

	if getBalanceCmd.Parsed() {
		getBalanceAddress := getBalanceCmd.String("address", "", "请输入查询金额的地址")
		if *getBalanceAddress == "" {
			fmt.Println("请输入查询金额的地址")
			os.Exit(1)
		}

		cli.getBalance(*getBalanceAddress)
	}

	if getBestHeightCmd.Parsed() {
		fmt.Printf("当前区块链最大高度：%d\n", cli.bc.GetBestHeight())
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressCmd.Parsed() {
		cli.listAddress()
	}

	if startNodeCmd.Parsed() {
		//  通过系统环境变量获取节点ID
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			fmt.Println("当前节点ID未设置")
			os.Exit(1)
		}

		cli.startNode(nodeID, *startNodeMinner)
	}
}