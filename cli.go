package main

import (
	"fmt"
	"os"
	"strconv"
)

type CLI struct 
{
}

const Usage = 
`
正确使用方法：
	./blockchain create <地址> "创建区块链"
	./blockchain addBlock <需要写入的的数据> "添加区块"
	./blockchain print "打印区块链"
	./blockchain getBalance <地址> "获取余额"
	./blockchain send <FROM> <TO> <AMOUNT> <MINER> <DATA>
	./blockchain createWallet "创建钱包"
	./blockchain listAddress "列举所有的钱包地址"
	./blockchain printTx "打印区块的所有交易"
`


func (cli *CLI) Run() 
{
	cmds := os.Args
	if len(cmds) < 2 
	{
		fmt.Println("输入参数无效，请检查!")
		fmt.Println(Usage)
		return
	}

	switch cmds[1] 
	{
	case "create":
		fmt.Println("创建区块被调用!")
		if len(cmds) != 3 
		{
			fmt.Println("输入参数无效，请检查!")
			return
		}
		address := cmds[2]
		cli.createBlockChain(address)

	case "addBlock":
		if len(cmds) != 3 
		{
			fmt.Println("输入参数无效，请检查!")
			return
		}
		data := cmds[2] 
		cli.addBlock(data)
		/*这玩意不可用，除非读取文件输入Txs,可以考虑使用json/csv作为临时数据介质，
		然后用json.Unmarshall这种方法读取到[]Transaction后进行*/
	case "print":
		fmt.Println("打印区块被调用!")
		cli.print()
	case "getBalance":
		fmt.Println("获取余额命令被调用!")
		if len(cmds) != 3 
		{
			fmt.Println("输入参数无效，请检查!")
			return
		}
		address := cmds[2] 
		cli.getBalance(address)
	case "send":
		fmt.Println("send命令被调用")
		if len(cmds) != 7 
		{
			fmt.Println("输入参数无效，请检查!")
			return
		}

		from := cmds[2]
		to := cmds[3]
		amount, _ := strconv.ParseFloat(cmds[4], 64)
		miner := cmds[5]
		data := cmds[6]
		cli.send(from, to, amount, miner, data)
	case "createWallet":
		fmt.Println("创建钱包命令被调用!")
		cli.createWallet()
	case "listAddress":
		fmt.Println("listAddress 被调用")
		cli.listAddress()
	case "printTx":
		cli.printTx()
	default:
		fmt.Println("输入参数无效，请检查!")
		fmt.Println(Usage)
	}
}
//这个没什么好改的，除非做成交互式的分布解析