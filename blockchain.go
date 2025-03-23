package main

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"strconv"
	"strings"
)

//	type BlockNode struct{
//		block Block
//	}使用deserializer可以直接实现为类链表,此处依照原教程使用bolt.db用数组模拟
type BlockChain struct {
	db   *bolt.DB
	tail []byte
}

const genesisInfo = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
const blockchainDBFile = "blockchain.db"
const bucketBlock = "bucketBlock"
const lastBlockHashKey = "lastBlockHashKey"

//教程中这一段集成了部分常用文本数据如创世语，桶和访问字段

func CreateBlockChain(address string) error {
	//鲁棒
	if isFileExist(blockchainDBFile) {
		fmt.Println("区块链文件已经开始存在！")
		return nil
	}

	db, err := bolt.Open(blockchainDBFile, 0600, nil)
	if err != nil {
		return err
	}

	//延迟
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))
		if bucket == nil {
			bucket, err := tx.CreateBucket([]byte(bucketBlock))
			if err != nil {
				return err
			}

			coinbase := NewCoinbaseTx(address, genesisInfo)
			txs := []*Transaction{coinbase}
			genesisBlock := NewBlock(txs, nil)
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte(lastBlockHashKey), genesisBlock.Hash)
		}
		return nil
	})
	return err
}

func GetBlockChainInstance() (*BlockChain, error) {
	//通用的鲁棒控制，函数必吃榜
	if !isFileExist(blockchainDBFile) {
		return nil, errors.New("区块链文件不存在，请先创建")
	}

	var lastHash []byte

	db, err := bolt.Open(blockchainDBFile, 0600, nil) //rwx  0110 => 6
	if err != nil {
		return nil, err
	}

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))

		if bucket == nil {
			return errors.New("bucket不应为nil") //鲁棒+1
		} else {
			lastHash = bucket.Get([]byte(lastBlockHashKey))
		}

		return nil
	})
	bc := BlockChain{db, lastHash}
	return &bc, nil
}

func (bc *BlockChain) AddBlock(txs1 []*Transaction) error {
	txs := []*Transaction{}
	fmt.Println("校验中") //你知道我要说什么
	for _, tx := range txs1 {
		if bc.verifyTransaction(tx) {
			fmt.Printf("当前交易校验成功:%x\n", tx.TXID)
			txs = append(txs, tx)
		} else {
			fmt.Printf("当前交易校验失败:%x\n", tx.TXID)
		}
	}

	lastBlockHash := bc.tail

	newBlock := NewBlock(txs, lastBlockHash)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))
		if bucket == nil {
			return errors.New("error:nil bucket in block adding")
		}

		bucket.Put(newBlock.Hash, newBlock.Serialize())
		bucket.Put([]byte(lastBlockHashKey), newBlock.Hash)

		//其实这里本质上在链表式区块链中就是更新头插点
		bc.tail = newBlock.Hash
		return nil
	})

	return err
}

type Iterator struct {
	db          *bolt.DB
	currentHash []byte //指针平替(bushi)
}

// bind
func (bc *BlockChain) NewIterator() *Iterator {
	it := Iterator{
		db:          bc.db,
		currentHash: bc.tail,
	}

	return &it
}

func (it *Iterator) Next() (block *Block) {

	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))
		if bucket == nil {
			return errors.New("error: nil bucket in iteration process")
		}

		blockTmpInfo := bucket.Get(it.currentHash)
		block = Deserialize(blockTmpInfo)
		it.currentHash = block.PrevHash

		return nil
	})
	if err != nil {
		fmt.Println("iterator next err:", err)
		return nil
	}
	return
}

type UTXOInfo struct {
	Txid     []byte
	Index    int64
	TXOutput //Annoymous Announcement
}

func (bc *BlockChain) FindMyUTXO(pubKeyHash []byte) (utxoInfos []UTXOInfo) {
	spentUtxos := make(map[string][]int)
	it := bc.NewIterator()
	for {
		block := it.Next()
		for _, tx := range block.Txs {
		LABEL:
			for outputIndex, output := range tx.TXOutputs {
				fmt.Println("outputIndex:", outputIndex)
				if bytes.Equal(output.ScriptPubKeyHash, pubKeyHash) {
					indexArray := spentUtxos[string(tx.TXID)]
					if len(indexArray) != 0 {
						for _, spendIndex := range indexArray {
							if outputIndex == spendIndex {
								continue LABEL
							}
						}
					}
					utxoinfo := UTXOInfo{tx.TXID, int64(outputIndex), output}
					utxoInfos = append(utxoInfos, utxoinfo) //加入主表
				}

			}
			if tx.isCoinbaseTx() /*我超，矿*/ {
				fmt.Println("Coinbase")
				continue
			}

			for _, input := range tx.TXInputs {
				//pub签名导致的
				if bytes.Equal(getPubKeyHashFromPubKey(input.PubKey), pubKeyHash) {
					spentKey := string(input.Txid)
					spentUtxos[spentKey] = append(spentUtxos[spentKey], int(input.Index))
				}
			}

		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return
}

func (bc *BlockChain) findNeedUTXO(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {
	var retMap = make(map[string][]int64)
	var retValue float64
	utxoInfos := bc.FindMyUTXO(pubKeyHash)
	for _, utxoinfo := range utxoInfos {
		retValue += utxoinfo.Value
		key := string(utxoinfo.Txid)
		retMap[key] = append(retMap[key], utxoinfo.Index)
		if retValue >= amount {
			break
		}
	}
	/*标准做法，类似于从一堆裤袋里逐个袋子掏钱包，每个钱包一张一张数，到了足够的金额直接付款，
	但是下一次付款时仍然需要traverse过去的UTXO。我的理解中可以在本地存储自己的UTXO池，
	每一次交易后更新，这样就可以通过参数对交易传入所需的UTXO*/
	return retMap, retValue
}

func (bc *BlockChain) signTransaction(tx *Transaction, priKey *ecdsa.PrivateKey) bool {
	fmt.Println("Transaction-signing initializing") //其实很多fmt输出只是为了cli的交互性(((
	prevTxs := make(map[string]*Transaction)
	for _, input := range tx.TXInputs {
		prevTx := bc.findTransaction(input.Txid)
		if prevTx == nil {
			fmt.Println("没有找到有效引用的交易")
			return false
		}

		fmt.Println("找到了引用的交易")
		prevTxs[string(input.Txid)] = prevTx
	}
	return tx.sign(priKey, prevTxs)
}
func (bc *BlockChain) verifyTransaction(tx *Transaction) bool {
	fmt.Println("verifying Transaction")

	if tx.isCoinbaseTx() {
		fmt.Println("Skip:Coinbase")
		return true
	}
	prevTxs := make(map[string]*Transaction)
	for _, input := range tx.TXInputs {
		prevTx := bc.findTransaction(input.Txid)
		if prevTx == nil {
			fmt.Println("没有找到有效引用的交易")
			return false
		}

		fmt.Println("找到了引用的交易")
		prevTxs[string(input.Txid)] = prevTx
	}

	return tx.verify(prevTxs)
}

// 这一系列的方法其思路没多大差别
func (bc *BlockChain) findTransaction(txid []byte) *Transaction {
	it := bc.NewIterator()

	for {
		block := it.Next()

		for _, tx := range block.Txs {
			if bytes.Equal(tx.TXID, txid) {
				return tx
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return nil
}

func (bc *BlockChain) BitstoDifficulty(bits uint64) (targetStr string) {
	const lent int = 64

	var mid uint64 = bits & 0xFFFFFFFF
	var potential uint64 = (mid >> 24) & 0xFF
	var base1 uint64 = mid & 0x00FFFFFF
	potential = 8 * (potential - 3)
	var oristr = strconv.FormatUint(base1, 16) + strings.Repeat("0", int(potential))
	str_noprefix := strings.TrimPrefix(oristr, "0x")
	targetStr = strings.Repeat("0", lent-len(str_noprefix))+str_noprefix
	return
}

func (bc *BlockChain) DifficultytoBits(diff string)(Bits uint64){
	var count=0
	var tgt=diff
	for {
        if strings.HasSuffix(tgt, "00") {
            tgt = tgt[:len(tgt)-2]
            count++
        } else {
            break
        }
    }
	for {
        if strings.HasPrefix(tgt, "00") {
            tgt = tgt[2:]
        } else {
            break
        }
    }
	var potential uint64=uint64(count+3)<<24
	var base,err =strconv.ParseUint(tgt,16,64)
	if err!=nil{
		return 0
	}
	Bits=(potential+base)&0xFFFFFFFF
	return
}
