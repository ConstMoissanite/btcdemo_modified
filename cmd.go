package main

import (
	"fmt"
	"os"
	"encoding/json"
)

func (cli *CLI) addBlock(data string) {
	// fmt.Println("添加区块被调用!")
	// bc, _ := GetBlockChainInstance()
	// err := bc.AddBlock(data)
	// if err != nil {
	// 	fmt.Println("AddBlock failed:", err)
	// 	return
	// }
	// fmt.Println("添加区块成功!")
}

/*
猜猜原作者为什么把这段注释掉？因为输入参数调用的时候给的根本就不是Tx，
给的string和bc.addblock是不匹配的。我在底部尝试实现了json读入，可能可以配合使用
*/
func (cli *CLI) createBlockChain(address string) {
	if !isValidAddress(address) {
		fmt.Println("Invalid Input Address:", address)
		return
	}

	err := CreateBlockChain(address)
	if err != nil {
		fmt.Println("CreateBlockChain failed:", err)
		return
	}
	fmt.Println("执行完毕!")
}

func (cli *CLI) print() {
	bc, err := GetBlockChainInstance()
	if err != nil {
		fmt.Println("print err:", err)
		return
	}

	defer bc.db.Close()

	//调用迭代器，输出blockChain
	it := bc.NewIterator()
	for {
		block := it.Next()

		fmt.Printf("\n++++++++++++++++++++++\n") //硬核手动分割线
		fmt.Printf("Version : %d\n", block.version)
		fmt.Printf("PrevHash : %x\n", block.PrevHash)
		fmt.Printf("MerkleRoot : %x\n", block.MerkleRoot)
		fmt.Printf("TimeStamp : %d\n", block.TimeStamp)
		fmt.Printf("Bits : %d\n", block.bits)
		fmt.Printf("Nonce : %d\n", block.Nonce)
		fmt.Printf("Hash : %x\n", block.Hash)
		fmt.Printf("Data : %s\n", block.Txs[0].TXInputs[0].ScriptSig)
		pow := NewProofOfWork(block)
		fmt.Printf("IsValid: %v\n", pow.IsValid())

		if block.PrevHash == nil {
			fmt.Println("区块链遍历结束!")
			/*遍历这个事说明了区块链就是天生适合用链表来玩，
			只是原作者老师用的是数组模拟，此外用C++ vector之类的STL容器也可以轻松实现*/
			break
		}
	}

}

func (cli *CLI) getBalance(address string) {
	if !isValidAddress(address) {
		fmt.Println("Invalid Input Address:", address)
		return
	}
	bc, err := GetBlockChainInstance()
	if err != nil {
		fmt.Println("getBalance err:", err)
		return
	}

	defer bc.db.Close()
	pubKeyHash := getPubKeyHashFromAddress(address)
	utxoinfos := bc.FindMyUTXO(pubKeyHash)
	total := 0.0

	for _, utxo := range utxoinfos {
		total += utxo.TXOutput.Value
	}

	fmt.Printf("'%s'的金额为:%f\n", address, total)
}

func (cli *CLI) send(from, to string, amount float64, miner, data string) {
	if !isValidAddress(from) {
		fmt.Println("Invalid Source Address:", from)
		return
	}

	if !isValidAddress(to) {
		fmt.Println("Invalid Destination Address:", to)
		return
	}

	if !isValidAddress(miner) {
		fmt.Println("Invalid Verifier Address:", miner)
		return
	}

	bc, err := GetBlockChainInstance()

	if err != nil {
		fmt.Println("send err:", err)
		return
	}

	defer bc.db.Close()
	coinbaseTx := NewCoinbaseTx(miner, data)
	txs := []*Transaction{coinbaseTx}
	tx := NewTransaction(from, to, amount, bc)
	if tx != nil {
		fmt.Println("找到一笔有效的转账交易!")
		txs = append(txs, tx)
	} else {
		fmt.Println("注意，找到一笔无效的转账交易, 不添加到区块!")
	}

	err = bc.AddBlock(txs) //这里调用的还是在之前文件里面定义的bc的方法，顶上那个方法无意义
	if err != nil {
		fmt.Println("添加区块失败，转账失败!")
	}

	fmt.Println("添加区块成功，转账成功!")
}

func (cli *CLI) createWallet() {
	wm := NewWalletManager()
	if wm == nil {
		fmt.Println("createWallet失败!")
		return
	}
	address := wm.createWallet()

	if len(address) == 0 {
		fmt.Println("创建钱包失败！")
		return
	}

	fmt.Println("新钱包地址为:", address)
}

func (cli *CLI) listAddress() {
	wm := NewWalletManager()
	if wm == nil {
		fmt.Println(" NewWalletManager 失败!")
		return
	}

	addresses := wm.listAddresses()
	for _, address := range addresses {
		fmt.Printf("%s\n", address)
	}
}

func (cli *CLI) printTx() {
	bc, err := GetBlockChainInstance()
	if err != nil {
		fmt.Println("getBalance err:", err)
		return
	}

	defer bc.db.Close()

	it := bc.NewIterator()
	for {
		block := it.Next()
		fmt.Println("\n+++++++++++++++++ 区块分割 +++++++++++++++")

		for _, tx := range block.Txs {
			fmt.Println(tx)
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CLI) TxUnmarshal(filename string) (txs []Transaction) {
	TxInBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Invalid Filereader:", err)
		return
	}
	err = json.Unmarshal(TxInBytes, &txs)
	if err != nil {
		fmt.Println("Invalid Unmarshaler:", err)
		return
	}
	return
} //这个方法配合cli.addblock使用就可以从json传入tx档案了
