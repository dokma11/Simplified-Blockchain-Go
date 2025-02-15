package main

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
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

// NewBlockchain creates a new blockchain starting with the genesis block
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic("ERROR: error while opening file: ", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic("ERROR: error while updating blockchain: ", err)
	}

	return &Blockchain{tip, db}
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	db := initializeDB()
	tip := createGenesisBlock(db, address)

	return &Blockchain{tip, db}
}

func initializeDB() *bolt.DB {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	return db
}

func createGenesisBlock(db *bolt.DB, address string) []byte {
	var tip []byte

	err := db.Update(func(tx *bolt.Tx) error {
		genesis := createGenesisTransaction(address)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		storeBlock(b, genesis)
		tip = []byte(genesis.Hash)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return tip
}

func createGenesisTransaction(address string) *Block {
	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	return NewGenesisBlock(cbtx)
}

func storeBlock(bucket *bolt.Bucket, block *Block) {
	if err := bucket.Put([]byte(block.Hash), block.Serialize()); err != nil {
		log.Panic(err)
	}
	if err := bucket.Put([]byte("l"), []byte(block.Hash)); err != nil {
		log.Panic(err)
	}
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
	validateTransactions(transactions, bc)

	lastHash := bc.getLastBlockHash()

	newBlock := NewBlock(transactions, lastHash)

	bc.saveBlock(newBlock)
}

func validateTransactions(transactions []*Transaction, bc *Blockchain) {
	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}
}

func (bc *Blockchain) getLastBlockHash() []byte {
	var lastHash []byte

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastHash
}

func (bc *Blockchain) saveBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if err := b.Put([]byte(block.Hash), block.Serialize()); err != nil {
			log.Panic(err)
		}
		if err := b.Put([]byte("l"), []byte(block.Hash)); err != nil {
			log.Panic(err)
		}

		bc.Tip = []byte(block.Hash)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if isTransactionUnspent(tx, pubKeyHash, spentTXOs) {
				unspentTXs = append(unspentTXs, *tx)
			}

			if !tx.IsCoinbase() {
				markSpentOutputs(tx, pubKeyHash, spentTXOs)
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func isTransactionUnspent(tx *Transaction, pubKeyHash []byte, spentTXOs map[string][]int) bool {
	for outIdx, out := range tx.Vout {
		if isOutputSpent(tx.ID, outIdx, spentTXOs) {
			continue
		}

		if out.IsLockedWithKey(pubKeyHash) {
			return true
		}
	}
	return false
}

func isOutputSpent(txID string, outIdx int, spentTXOs map[string][]int) bool {
	for _, spentOutIdx := range spentTXOs[txID] {
		if spentOutIdx == outIdx {
			return true
		}
	}
	return false
}

func markSpentOutputs(tx *Transaction, pubKeyHash []byte, spentTXOs map[string][]int) {
	for _, in := range tx.Vin {
		if in.UsesKey(pubKeyHash) {
			spentTXOs[in.TxID] = append(spentTXOs[in.TxID], in.Vout)
		}
	}
}

func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

	for _, tx := range unspentTXs {
		if accumulated >= amount {
			break
		}
		accumulated = collectSpendableOutputs(tx, pubKeyHash, amount, accumulated, unspentOutputs)
	}

	return accumulated, unspentOutputs
}

func collectSpendableOutputs(tx Transaction, pubKeyHash []byte, amount, accumulated int, unspentOutputs map[string][]int) int {
	for outIdx, out := range tx.Vout {
		if accumulated >= amount {
			break
		}

		if out.IsLockedWithKey(pubKeyHash) {
			accumulated += out.Value
			unspentOutputs[tx.ID] = append(unspentOutputs[tx.ID], outIdx)
		}
	}
	return accumulated
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare([]byte(tx.ID), ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTXs := bc.getPreviousTransactions(tx)
	tx.Sign(privateKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := bc.getPreviousTransactions(tx)
	return tx.Verify(prevTXs)
}

func (bc *Blockchain) getPreviousTransactions(tx *Transaction) map[string]Transaction {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction([]byte(vin.TxID))
		if err != nil {
			log.Panic("ERROR: Failed to find previous transaction: ", err)
		}
		prevTXs[prevTX.ID] = prevTX
	}

	return prevTXs
}
