package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

// TODO: mozda se i ovde mogu vratiti kasnije
const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks" // BC genesis block data

type Blockchain struct {
	Tip []byte
	DB  *bolt.DB
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, "")
}

// NewBlockchain creates a new blockchain starting with the genesis block
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip, db}
}

// CreateBlockchain creates a new blockchain DB
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
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte(genesis.Hash), genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), []byte(genesis.Hash))
		if err != nil {
			log.Panic(err)
		}
		tip = []byte(genesis.Hash)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := Blockchain{tip, db}
	return &bc
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{string(bc.Tip), bc.DB}
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, string(lastHash))

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put([]byte(newBlock.Hash), newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), []byte(newBlock.Hash))
		if err != nil {
			log.Panic(err)
		}
		bc.Tip = []byte(newBlock.Hash)
		return nil
	})
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {

		Outputs:
			for outIdx, out := range tx.Vout {
				// Check if the output was spent
				if spentTXOs[tx.ID] != nil {
					for _, spentOut := range spentTXOs[tx.ID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						spentTXOs[in.TxID] = append(spentTXOs[in.TxID], in.Vout)
					}
				}
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return unspentTXs
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
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

	for _, tx := range unspentTXs {
		if accumulated >= amount {
			break
		}
		accumulated = collectSpendableOutputs(tx, address, amount, accumulated, unspentOutputs)
	}

	return accumulated, unspentOutputs
}

func collectSpendableOutputs(tx Transaction, address string, amount, accumulated int, unspentOutputs map[string][]int) int {
	for outIdx, out := range tx.Vout {
		if accumulated >= amount {
			break
		}

		if out.CanBeUnlockedWith(address) {
			accumulated += out.Value
			unspentOutputs[tx.ID] = append(unspentOutputs[tx.ID], outIdx)
		}
	}
	return accumulated
}
