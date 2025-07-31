package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type Block struct {
	nonce        int
	previousHash [32]byte
	timestamp    int64
	transactions []*Transaction
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash [32]byte       `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: b.previousHash,
		Transactions: b.transactions,
	})
}

type Blockchain struct {
	transactionsPool []*Transaction
	chain            []*Block
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash)
	bc.chain = append(bc.chain, b)
	b.transactions = append(b.transactions, bc.transactionsPool...)
	bc.transactionsPool = []*Transaction{}
	return b
}

func NewBlockchain() *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.CreateBlock(0, b.Hash())
	return bc
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	fmt.Println(string(m))
	return sha256.Sum256(m)
}

func NewBlock(nonce int, previousHash [32]byte) *Block {
	b := new(Block)
	b.nonce = nonce
	b.timestamp = time.Now().UnixNano()
	b.previousHash = previousHash
	return b
}

func init() {
	log.SetPrefix("Blockchain: ")
}

func (b *Block) Print() {
	fmt.Printf("timestamp        %d\n", b.timestamp)
	fmt.Printf("nonce            %d\n", b.nonce)
	fmt.Printf("previousHash     %x\n", b.previousHash)
	for _, transaction := range b.transactions {
		transaction.Print()
	}
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("#", 25), i, strings.Repeat("#", 25))
		block.Print()
	}
	fmt.Printf("%s", strings.Repeat("=", 60))
}

type Transaction struct {
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

func NewTransaction(sender string, receiver string, value float32) *Transaction {
	return &Transaction{
		senderBlockchainAddress:    sender,
		recipientBlockchainAddress: receiver,
		value:                      value,
	}
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" sender_blockchain_address:                    %s\n", t.senderBlockchainAddress)
	fmt.Printf(" recipient_blockchain_address:                 %s\n", t.recipientBlockchainAddress)
	fmt.Printf(" value:                                        %.1f\n", t.value)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender"`
		Recipient string  `json:"recipient"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32) {
	bc.transactionsPool = append(bc.transactionsPool, NewTransaction(sender, recipient, value))
}

func main() {
	bc := NewBlockchain()
	bc.Print()

	bc.AddTransaction("A", "B", 100.0)
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(1, previousHash)
	bc.Print()

	bc.AddTransaction("B", "C", 200.0)
	previousHash = bc.LastBlock().Hash()
	bc.CreateBlock(2, previousHash)
	bc.Print()
}
