package cli

import (
	"blockchain-service/internal/blockchain"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{
	blockchain *blockchain.BlockChain
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

func (cli *CommandLine) addBlock(data string){
	cli.blockchain.AddBlock(data)
	fmt.Println("Block Added!")
}

func (cli *CommandLine) printChain(){
	iter := cli.blockchain.Iterator()

	for{
		block := iter.Next()

		fmt.Printf("Prev Hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		
		pow := blockchain.NewProof(block)
		fmt.Printf("Pow: %s\n\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0{
			break
		}
	}
}

func (cli *CommandLine) run(){
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "BlockData")

	switch os.Args[1]{
		case "add":
			err := addBlockCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		case "print":
			err := printChainCmd.Parse(os.Args[2:])
			blockchain.Handle(err)
		default:
			cli.printUsage()
			runtime.Goexit()
	}

	if addBlockCmd.Parsed(){
		if *addBlockData == ""{
			cli.printUsage()
			runtime.Goexit()
		}

		cli.blockchain.AddBlock(*addBlockData)
	}

	if printChainCmd.Parsed(){
		cli.printChain()
	}
}
