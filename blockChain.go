package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"os"
)

const dbFile = "blockChain.db"
const blockBucket = "bucket"
const lastHashKey = "key"
const genesisInfo = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type BlockChain struct {
	//blocks []*Block
	// 数据库操作句柄
	db *bolt.DB
	// 尾巴, 表示最后一个区块的哈希值
	tail []byte
}

func InitBlockChain(address string) *BlockChain  {
	if isDBExist() {
		fmt.Println("blockchain exist!")
		os.Exit(1)
	}
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr("InitBlockChain", err)
	var lastHash []byte
	// db.View
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))

		// 创建创世块
		coinbase := NewCoinbaseTx(address, genesisInfo)
		genesis := NewGenesisBlock(coinbase)
		bucket, err = tx.CreateBucket([]byte(blockBucket))
		CheckErr("InitBlockChain2", err)
		bucket.Put(genesis.Hash, genesis.Serialize())
		CheckErr("InitBlockChain3", err)
		bucket.Put([]byte(lastHashKey), genesis.Hash)
		CheckErr("InitBlockChain4", err)
		lastHash = genesis.Hash

		return nil
	})

	return &BlockChain{db, lastHash}
}

func isDBExist() bool {
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func GetBlockChainHandler() *BlockChain {
	if !isDBExist() {
		fmt.Println("Pls create blockchain first!")
		os.Exit(1)
	}
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr("GetBlockChainHandler", err)
	var lastHash []byte
	// db.View
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket != nil {
			// 取出最后区块的Hash值
			lastHash = bucket.Get([]byte(lastHashKey))
		}
		return nil
	})

	return &BlockChain{db, lastHash}
}

func (bc *BlockChain)AddBlock(txs []*Transaction)  {
	var prevBlockHash []byte

	bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			os.Exit(1)
		}

		prevBlockHash =  bucket.Get([]byte(lastHashKey))
		return nil
	})
	block := NewBlock(txs, prevBlockHash)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			os.Exit(1)
		}

		err := bucket.Put(block.Hash, block.Serialize())
		CheckErr("AddBlock1", err)
		err = bucket.Put([]byte(lastHashKey), block.Hash)
		CheckErr("AddBlock2", err)
		bc.tail = block.Hash
		return nil
	})

	CheckErr("AddBlock2", err)
}

// 迭代器，就是一个对象， 它里面包含了一个游标，一直向前（向后）移动

type BlockChainIterator struct {
	currHash []byte
	db *bolt.DB
}

func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{currHash: bc.tail, db: bc.db}
}

func (it *BlockChainIterator)Next()  (block *Block) {
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			os.Exit(1)
		}

		data := bucket.Get(it.currHash)
		block = Deserialize(data)
		it.currHash = block.PrevBlockHash
		return nil
	})
	CheckErr("Next()", err)
	return
}

// 返回指定地址能够支配的utxo的集合
func (bc *BlockChain)FindUTXOTransactions(address string) []Transaction{
	var UTXOTransactions []Transaction
	// 存储使用过的utxo的集合, map[交易id][]int64
	// 0x1111111: 0,1 都是给Alice转账
	spentUTXO := make(map[string][]int64)

	it := bc.NewIterator()
	// 遍历block
	for true {
		block := it.Next()

		// 遍历交易
		for _, tx := range block.Transactions{
			// 遍历input
			// 找到已经消耗的utxo
			if !tx.IsCoinbase() {
				for _,input := range tx.TXInputs{
					if input.CanUnlockUTXOWith(address) {
						spentUTXO[string(input.TXID)] = append(spentUTXO[string(input.TXID)], input.Vout)
					}
				}
			}

		OUTPUTS:
			// 遍历outputs
			for currIndex,output := range tx.TXOutputs{
				// 检查当前的output是否已经被消耗，如果消耗过，那么进行下一个output检验
				if spentUTXO[string(tx.TXID)] != nil {
					indexes := spentUTXO[string(tx.TXID)]
					for _, index := range indexes {
						// 当前的索引和消耗的索引比较，若相同，表明这个欧太普通被消耗了,
						if int64(currIndex) == index {
							continue OUTPUTS
						}
					}
				}
				// 如果当前地址是这个utxo的所有者
				if output.CanBeUnlockedWith(address) {
					UTXOTransactions = append(UTXOTransactions, *tx)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXOTransactions
}

func (bc *BlockChain) FindUTXO(address string) []TXOutput  {
	var UTXOs []TXOutput
	txs := bc.FindUTXOTransactions(address)

	for _, tx := range txs{
		for _, utxo := range tx.TXOutputs{
			if utxo.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, utxo)
			}
		}
	}
	return UTXOs
}

func (bc *BlockChain) FindSuitableUTXOs(adderss string, amount float64) (map[string][]int64, float64) {
	txs := bc.FindUTXOTransactions(adderss)
	validUTXOs := make(map[string][]int64)
	var total float64

	// 遍历交易
	for _, tx := range txs {
		outputs := tx.TXOutputs
		// 遍历Outputs
		for index, output := range outputs {
			if output.CanBeUnlockedWith(adderss) {
				// 判断当前收集的utxo的总金额是否大于所需要花费的金额
				if total < amount {
					total += output.Value
					validUTXOs[string(tx.TXID)] = append(validUTXOs[string(tx.TXID)], int64(index))
				}
			}

		}
	}
	return validUTXOs, total
}

