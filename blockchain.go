package main

import (
	"fmt"
	"log"
	"strings"
)

const (
	MINING_DIFFICULTY int     = 3
	MINING_REWARD     float32 = 1.0
	CHAIN_ADDRESS     string  = "COIN_BASE"
)

type Blockchain struct {
	transactionsPool  []*Transaction
	chain             []*Block
	blockchainAddress string
}

func (bc *Blockchain) CopyTransactionsPool() []*Transaction {
	transactions := make([]*Transaction, 0, len(bc.transactionsPool))
	for _, t := range bc.transactionsPool {
		transactions = append(transactions, t.Clone())
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{
		nonce:        nonce,
		previousHash: previousHash,
		timestamp:    0,
		transactions: transactions,
	}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	fmt.Println(guessHashStr)
	return guessHashStr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	previousHash := bc.LastBlock().previousHash
	transactions := bc.CopyTransactionsPool()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash)
	bc.chain = append(bc.chain, b)
	b.transactions = append(b.transactions, bc.transactionsPool...)
	bc.transactionsPool = []*Transaction{}
	return b
}

func NewBlockchain(blockchainAddress string) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	bc.CreateBlock(0, b.Hash())
	return bc
}

func init() {
	log.SetPrefix("Blockchain: ")
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("#", 25), i, strings.Repeat("#", 25))
		block.Print()
	}
	fmt.Printf("%s", strings.Repeat("=", 60))
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32) {
	bc.transactionsPool = append(bc.transactionsPool, NewTransaction(sender, recipient, value))
}

func (bc *Blockchain) Mining() bool {
	bc.AddTransaction(CHAIN_ADDRESS, bc.blockchainAddress, MINING_REWARD)
	none := bc.ProofOfWork()
	previousHash := bc.LastBlock().previousHash
	bc.CreateBlock(none, previousHash)
	return true
}

func main() {
	bc := NewBlockchain("cuidongliang")
	bc.Print()

	bc.AddTransaction("A", "B", 100.0)
	bc.Mining()
	bc.Print()

	bc.AddTransaction("B", "C", 200.0)
	bc.AddTransaction("X", "Y", 300.0)
	bc.Mining()
	bc.Print()
}
