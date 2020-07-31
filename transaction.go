package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
)

const REWARD = 12.5

type Transaction struct {
	// 交易id
	TXID []byte
	// 输入
	TXInputs []TXInput
	// 输出
	TXOutputs []TXOutput
}

type TXInput struct {
	// 所引用输出的交易id
	TXID []byte
	// 所引用output的索引值
	Vout int64
	// 解锁脚本，指名可以使用某个output操作
	ScripSig string
}

// 检查当前用户能否解开引用的utxo
func (input * TXInput)CanUnlockUTXOWith(unlockData string) bool {
	return input.ScripSig == unlockData
}

type TXOutput struct {
	// 支付给对方的金额
	Value float64
	// 锁定脚本，指定收款方的地址
	ScriptPubKey string
}
// 检查当前用户是否是这个utxo的所有者
func (output * TXOutput)CanBeUnlockedWith(unlockData string) bool{
	return output.ScriptPubKey == unlockData
}

// 设置交易id，是一个hash值
func (tx *Transaction)SetTXID()  {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)

	CheckErr("SetTXID", err)
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXID = hash[:]

}

func NewCoinbaseTx(address string, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s %d btc", address, REWARD)
	}
	input := TXInput{[]byte{},-1,data}
	output := TXOutput{REWARD, address}
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{output}}
	tx.SetTXID()
	return &tx
}

func (tx *Transaction)IsCoinbase() bool {
	if len(tx.TXInputs) == 1 {
		if len(tx.TXInputs[0].TXID) == 0 && tx.TXInputs[0].Vout == -1 {
			return true
		}
	}
	return false
}

func NewTransaction(from string, to string, amount float64, bc *BlockChain) *Transaction  {

	validUTXOs := make(map[string][]int64)
	var total float64

	validUTXOs,  total = bc.FindSuitableUTXOs(from, amount)

	if total < amount {
		fmt.Println("Not enough money")
		os.Exit(1)
	}

	var inputs []TXInput
	var outputs []TXOutput
	for txId, outputIndexes := range validUTXOs{
		for _, index := range outputIndexes{
			input := TXInput{TXID: []byte(txId), Vout: int64(index), ScripSig: from}
			inputs = append(inputs, input)
		}
	}

	// 2.创建outputs
	output := TXOutput{amount, to}
	outputs = append(outputs, output)

	if total > amount {
		output := TXOutput{total -amount, from}
		outputs = append(outputs, output)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetTXID()
	return &tx
}