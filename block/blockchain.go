package block

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func (bc *Blockchain) GetDb() *bolt.DB {
	return bc.db
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

func (bc *Blockchain) LastBlock() *Block {
	var lastBlockByte []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		lastBlockByte = b.Get(lastHash)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return DeserializedBlock(lastBlockByte)
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func NewBlockchain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip, db}
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		gensis := NewGenesisBlock(cbtx)
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}
		err = b.Put(gensis.Hash, gensis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), gensis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = gensis.Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return &Blockchain{tip, db}
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializedBlock(encodedBlock)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash
	return block
}

// FindUnspentTransactions returns a list of transactions containing unspent output
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXS []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		b := bci.Next()
		for _, tx := range b.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, vout := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if vout.CanBeUnlockedWith(address) {
					unspentTXS = append(unspentTXS, *tx)
				}
			}

			if !tx.IsCoinBase() {
				for _, vin := range tx.Vin {
					if vin.CanUnlockOutputWith(address) {
						inTxId := hex.EncodeToString(vin.Txid)
						spentTXOs[inTxId] = append(spentTXOs[inTxId], vin.Vout)
					}
				}
			}
		}

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXS
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	accmulated := 0
	unspentOutputs := make(map[string][]int)
	unspentTXS := bc.FindUnspentTransactions(address)

Work:
	for _, tx := range unspentTXS {
		txId := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				unspentOutputs[txId] = append(unspentOutputs[txId], outIdx)
				accmulated += out.Value

				if accmulated >= amount {
					break Work
				}
			}
		}
	}

	return accmulated, unspentOutputs
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var prevHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		prevHash = bucket.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, prevHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		err = bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
}
