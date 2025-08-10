package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
)

const subsidy int = 10

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (t *Transaction) Print() {
	fmt.Printf("id: %x\n", t.ID)
	for _, ti := range t.Vin {
		ti.Print()
	}
	for _, to := range t.Vout {
		to.Print()
	}
}

func (ti *TXInput) Print() {
	fmt.Println(strings.Repeat("=", 25))
	fmt.Printf("Txid: %x\n", ti.Txid)
	fmt.Printf("Vout: %d\n", ti.Vout)
	fmt.Printf("ScriptSig: %s\n", ti.ScriptSig)
}

func (to *TXOutput) Print() {
	fmt.Println(strings.Repeat("=", 25))
	fmt.Printf("Value: %d\n", to.Value)
	fmt.Printf("ScriptPubKey: %s\n", to.ScriptPubKey)
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 &&
		len(tx.Vin[0].Txid) == 0 &&
		tx.Vin[0].Vout == -1
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}

		// Build a list of outputs
		outputs = append(outputs, TXOutput{amount, to})
		if acc > amount {
			outputs = append(outputs, TXOutput{acc - amount, from})
		}
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
