package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `
	createChain --address ADDRESS "create a blockchain"
	send --from FROM --to TO --amount AMOUNT "send coin from FROM to TO"
	getBalance --address "get balance of the address"
	printChain				"print all blocks"
`

const PrintChainCmdString = "printChain"
const CreateChainCmdString = "createChain"
const getBalanceCmdString = "getBalance"
const sendCmdString = "send"

type CLI struct {
	//bc *BlockChain
}

func (cli *CLI) PrintUsage() {
	fmt.Println(usage)
	os.Exit(1)
}

//func (cli *CLI) ParamCheck() {
//	fmt.Println(usage)
//	os.Exit(1)
//}

func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		fmt.Println("invalid input!")
		cli.PrintUsage()
	}
	printChainCmd := flag.NewFlagSet(PrintChainCmdString, flag.ExitOnError)
	createChainCmd := flag.NewFlagSet(CreateChainCmdString, flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet(getBalanceCmdString, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(sendCmdString, flag.ExitOnError)

	createChainCmdPara := createChainCmd.String("address", "", "address info!")
	getBalanceCmdPara := getBalanceCmd.String("address", "", "address info!")

	// send相关参数
	fromPara := sendCmd.String("from", "", "sender address info!")
	toPara := sendCmd.String("to", "", "receiver address info!")
	amountPara := sendCmd.Float64("amount", 0, "amount info!")

	switch os.Args[1] {
	case sendCmdString:
		// 发送交易
		err := sendCmd.Parse(os.Args[2:])
		CheckErr("Run()", err)
		if sendCmd.Parsed() {
			if *fromPara == "" || *toPara == "" || *amountPara == 0 {
				fmt.Println("send cmd parameters invalid!")
				cli.PrintUsage()
			}
			cli.Send(*fromPara,*toPara,*amountPara)
		}
	case CreateChainCmdString:
		// 创建区块链
		err := createChainCmd.Parse(os.Args[2:])
		CheckErr("Run()", err)
		if createChainCmd.Parsed() {
			if *createChainCmdPara == "" {
				fmt.Println("address should not be empty!")
				cli.PrintUsage()
			}
			cli.CreateChain(*createChainCmdPara)
		}
	case getBalanceCmdString:
		// 创建区块链
		err := getBalanceCmd.Parse(os.Args[2:])
		CheckErr("Run()", err)
		if getBalanceCmd.Parsed() {
			if *getBalanceCmdPara == "" {
				fmt.Println("address1 should not be empty!")
				cli.PrintUsage()
			}
			cli.GetBalance(*getBalanceCmdPara)
		}
	case PrintChainCmdString:
		// 打印输出
		err := printChainCmd.Parse(os.Args[2:])
		CheckErr("Run2()", err)
		if printChainCmd.Parsed() {
			cli.PrintChain()
		}
	default:
		cli.PrintUsage()

	}
}
