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

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks" // BC genesis block data

type Blockchain struct {
	Tip []byte
	DB  *bolt.DB
}

// NewBlockchain creates a new blockchain starting with the genesis block
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
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
func CreateBlockchain(address string, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
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

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get([]byte(block.Hash))

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put([]byte(block.Hash), blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), []byte(block.Hash))
			if err != nil {
				log.Panic(err)
			}
			bc.Tip = []byte(block.Hash)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	validateTransactions(transactions, bc)

	lastHash, lastHeight := bc.getLastBlockHash()

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

	bc.saveBlock(newBlock)

	return newBlock
}

func validateTransactions(transactions []*Transaction, bc *Blockchain) {
	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}
}

func (bc *Blockchain) getLastBlockHash() ([]byte, int) {
	var lastHash []byte
	var lastHeight int

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})
	if err != nil {
		log.Panic("ERROR: error while getting last block hash: ", err)
	}

	return lastHash, lastHeight
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

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			checkOutputs(tx, UTXO, spentTXOs)

			if !tx.IsCoinbase() {
				markSpentOutputs(tx, spentTXOs)
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return UTXO
}

func checkOutputs(tx *Transaction, UTXO map[string]TXOutputs, spentTXOs map[string][]int) {
	for outIdx, out := range tx.Vout {
		if isOutputSpent(tx.ID, outIdx, spentTXOs) {
			continue
		}

		outs := UTXO[tx.ID]
		outs.Outputs = append(outs.Outputs, out)
		UTXO[tx.ID] = outs
	}
}

func isOutputSpent(txID string, outIdx int, spentTXOs map[string][]int) bool {
	for _, spentOutIdx := range spentTXOs[txID] {
		if spentOutIdx == outIdx {
			return true
		}
	}
	return false
}

func markSpentOutputs(tx *Transaction, spentTXOs map[string][]int) {
	for _, in := range tx.Vin {
		spentTXOs[in.TxID] = append(spentTXOs[in.TxID], in.Vout)
	}
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
	if tx.IsCoinbase() {
		return true
	}
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

func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("block is not found")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, []byte(block.Hash))

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return blocks
}
