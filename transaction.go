package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

func (t *Transaction) Clone() *Transaction {
	return &Transaction{
		senderBlockchainAddress:    t.senderBlockchainAddress,
		recipientBlockchainAddress: t.recipientBlockchainAddress,
		value:                      t.value,
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
