package main

/*
这里是我觉得好一些的原始定义
type Block_Content struct {
	Version   uint64
	TimeStamp uint64
	Bits      uint64 //这里其实32位就足够了
	Nonce     uint64

	PrevHash   []byte
	MerkleRoot []byte

	Txs []*Transaction
}

type Block struct {
	Content  Block_Content
	PresHash []byte
}

type BlockChain_Node struct {
	Block
	PrevPtr *Block
} //链表式实现，适用于物理存储

type BlockChain struct {
	length uint64
	tail   *BlockChain_Node
}



func rewardcalculator(bc *BlockChain)(res float64)
{
 const term uint64=210000
 res=math.Pow(2,bc.length/term)*50
 return
}

以下需要uint256
*/
