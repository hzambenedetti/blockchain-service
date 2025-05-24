package cli

import (
	"blockchain-service/internal/blockchain"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{
}

func (cli *CommandLine) printUsage(){
	fmt.Println("Usage: ")
	fmt.Println(" add -block BLOCK_DATA  -- Add a block to the blockchain")
	fmt.Println(" print  -- Print the blocks in the chain")
}

func (cli *CommandLine) validateArgs(){
	if len(os.Args) < 2{
		cli.printUsage()
		runtime.Goexit()
	}
}

// func (cli *CommandLine) addBlock(data string){
// 	cli.blockchain.AddBlock(data)
// 	fmt.Println("Block Added!")
// }

func (cli *CommandLine) printChain(){
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for{
		block := iter.Next()

		fmt.Printf("Prev Hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		
		pow := blockchain.NewProof(block)
		fmt.Printf("Pow: %s\n\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0{
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address string){
	chain := blockchain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("BlockChain Created")
}

func (cli *CommandLine) getBalance(address string){
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUTXO(address)

	for _, out := range UTXOs{
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int){
	chain := blockchain.ContinueBlockChain(from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Printf("Sent %d tokens from %s to %s", amount, from, to)
}

func (cli *CommandLine) run(){
	cli.validateArgs()
	
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	getBalanceAddr := getBalanceCmd.String("address", "", "")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "")
	sendFrom := sendCmd.String("from", "", "")
	sendTo := sendCmd.String("to", "", "")
	sendAmount := sendCmd.Int("amount", 0, "")


	switch os.Args[1]{
		case "getbalance":
			err := getBalanceCmd.Parse(os.Args[2:])
			if err != nil{
				log.Panic(err)	
			}
		case "send":
			err := sendCmd.Parse(os.Args[2:])
			if err != nil{
				log.Panic(err)	
			}
		case "createblockchain":
			err := createBlockchainCmd.Parse(os.Args[2:])
			if err != nil{
				log.Panic(err)
			}
		case "print":
			err := printChainCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		default:
			cli.printUsage()
			runtime.Goexit()
	}
	
	if getBalanceCmd.Parsed(){
		if *getBalanceAddr == ""{
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddr)
	}
	if sendCmd.Parsed(){
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0{
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
	if createBlockchainCmd.Parsed(){
		if *createBlockchainAddress == ""{
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}
	if printChainCmd.Parsed(){
		cli.printChain()
	}
}
