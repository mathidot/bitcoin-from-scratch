package cli

import (
	"blockchain/block"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) printChain() {
	bc := block.NewBlockchain()
	bci := bc.Iterator()
	for {
		b := bci.Next()
		b.Print()
		pow := block.NewProofOfWork(b)
		fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := block.CreateBlockchain(address)
	bc.GetDb().Close()
	fmt.Println("Done!")
}

func (cli *CLI) send(from, to string, amount int) {
	bc := block.NewBlockchain()
	defer bc.GetDb().Close()

	tx := block.NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*block.Transaction{tx})
	fmt.Println("Success")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) getbalance(address string) {
	bc := block.NewBlockchain()
	defer bc.GetDb().Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) Run() {
	cli.validateArgs()
	// getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	// sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendData := sendCmd.String("from", "cuidongliang", "from user")
	toData := sendCmd.String("to", "cuidongliang", "to user")
	amountData := sendCmd.Int("amount", 10, "amount")
	getBalanceData := getBalanceCmd.String("address", "cuidongliang", "user's balance")
	createBlockchainData := createBlockchainCmd.String("address", "cuidongliang", "Blockchain address")

	switch os.Args[1] {
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		cli.createBlockchain(*createBlockchainData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if getBalanceCmd.Parsed() {
		cli.getbalance(*getBalanceData)
	}

	if sendCmd.Parsed() {
		cli.send(*sendData, *toData, *amountData)
	}

}
