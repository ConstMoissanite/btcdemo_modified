package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	version    uint64
	TimeStamp  uint64
	PrevHash   []byte
	Hash       []byte
	MerkleRoot []byte
	bits       uint64
	Nonce      uint64
	Txs        []*Transaction
}

func NewBlock(txs []*Transaction, prevHash []byte) (res *Block) {
	block := Block{
		version:    0, //没有动态版本和动态难度
		PrevHash:   prevHash,
		Hash:       nil,
		MerkleRoot: nil,
		TimeStamp:  uint64(time.Now().Unix()),
		bits:       0x0,
		Nonce:      0,
		Txs:        txs,
	}

	block.HashTransactionMerkleRoot()
	fmt.Printf("merkleRoot:%x\n", block.MerkleRoot)
	pow := NewProofOfWork(&block)
	hash, Nonce := pow.Run()
	block.Hash = hash
	block.Nonce = Nonce
	res = &block
	return
}

// Tx生成merkle根
func (block *Block) HashTransactionMerkleRoot() {
	var info [][]byte
	for _, tx := range block.Txs {
		txHashValue := tx.TXID //[]byte
		info = append(info, txHashValue)
	}

	value := bytes.Join(info, []byte{})
	hash := sha256.Sum256(value)
	block.MerkleRoot = hash[:]
}

// 双向查询
func (b *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(b)
	if err != nil {
		fmt.Printf("Encode err:%s", err)
		return nil
	}
	return buffer.Bytes()
}
func Deserialize(src []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(src))
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Printf("decode err:%s", err)
		return nil
	}
	return &block
}
