package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

type ProofOfWork struct {
	block *Block
	//目标值
	target *big.Int
}

const targetBits = 10

func NewProofOfWork(block *Block) *ProofOfWork {

	target := big.NewInt(1)
	target.Lsh(target,uint(256 - targetBits))
	pow := ProofOfWork{block: block, target: target}
	return &pow
}

func (pow *ProofOfWork)PrepareData(nonce int64)[]byte  {
	block := pow.block
	tmp := [][]byte{
		IntToByte(block.Version),
		block.PrevBlockHash,
		block.TransactionHash(),
		IntToByte(block.TimeStamp),
		IntToByte(targetBits),
		IntToByte(nonce),
		//block.Data
	}
	data := bytes.Join(tmp, []byte{})
	return data
}

func (pow *ProofOfWork)Run() (int64, []byte) {
	// 1. 拼装数据
	// 2. Hash值转big.Int类型

	var hash [32]byte
	var nonce int64 = 0
	var hashInt big.Int
	fmt.Println("Begin Mining...")
	fmt.Printf("target hash : %x\n", pow.target.Bytes())
	for nonce < math.MaxInt64 {
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("found hash: %x\n", hash)
			fmt.Printf("found nonce: %d\n", nonce)
			break
		} else {
			//fmt.Printf("not found nonce, nonce: %d, hash: %x\n", nonce, hash)
			nonce++
		}
	}
	return nonce, hash[:]
}

func (pow *ProofOfWork)IsValid() bool  {
	var hashInt big.Int

	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}