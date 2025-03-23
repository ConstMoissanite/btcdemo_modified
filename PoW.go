package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	targetStr := "0001000000000000000000000000000000000000000000000000000000000000"/*block.bitstodifficulty(block.bits)*/
	tmpBigInt := new(big.Int)
	//将我们的难度值赋值给bigint
	//我在实现bits转目标字符串的时候使用的是uint256库，原作者用的math/big，但我不想改了
	tmpBigInt.SetString(targetStr, 16)
	pow.target = tmpBigInt
	return &pow
}

func (pow *ProofOfWork) RewardCalculator(block *Block) (rew uint64) {
	return //我发现现在的数据结构里面没有加id这种可以用于直接查询区块位置的，于是id/21w拿去解出奖励的函数暂时搁置一下（悲
}
func (pow *ProofOfWork) Run() ([]byte, uint64) {

	var nonce uint64
	var hash [32]byte
	fmt.Println("Mining...")

	for {
		fmt.Printf("%x\r", hash[:])
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		tmpInt := new(big.Int)
		tmpInt.SetBytes(hash[:])
		if tmpInt.Cmp(pow.target) == -1 {
			fmt.Printf("挖矿成功,hash :%x, nonce :%d\n", hash[:], nonce)
			break
		} else {
			nonce++ //哈希是随机的，只能硬猜了捏
		}
	}
	return hash[:], nonce
}

func (pow *ProofOfWork) PrepareData(nonce uint64) []byte {
	b := pow.block

	tmp := [][]byte{
		uintToByte(b.version),
		b.PrevHash,
		b.MerkleRoot,
		uintToByte(b.TimeStamp),
		uintToByte(b.bits),
		uintToByte(nonce),
	}
	data := bytes.Join(tmp, []byte{})
	return data
}

func (pow *ProofOfWork) IsValid() bool {
	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	tmpInt := new(big.Int)
	tmpInt.SetBytes(hash[:])

	// if tmpInt.Cmp(pow.target) == -1 {
	// 	return true
	// }
	// return false

	//满足条件，返回true
	return tmpInt.Cmp(pow.target) == -1
}
