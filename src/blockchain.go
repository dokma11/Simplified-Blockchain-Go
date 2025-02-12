package main

import (
	"fmt"
	"github.com/boltdb/bolt"
)

// TODO: mozda se i ovde mogu vratiti kasnije
const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"

type Blockchain struct {
	Tip []byte
	DB  *bolt.DB
}

func NewGenesisBlock() *Block {
	return NewBlock("-", "Genesis Block")
}

func NewBlockchain() *Blockchain {
	var tip []byte
	db, _ := bolt.Open(dbFile, 0600, nil) // vratiti se ovde mozda treba ipak da se handluje error

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()
			genesisHash := []byte(genesis.Hash)

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				fmt.Printf("Error while creating a bucket: %s", err)
				return err
			}

			err = b.Put(genesisHash, genesis.Serialize())
			err = b.Put([]byte("l"), genesisHash)
			tip = genesisHash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	bc := Blockchain{tip, db}

	return &bc
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		fmt.Printf("Error while reading from DB: %s", err)
	}

	newBlock := NewBlock(string(lastHash), data)
	newBlockHash := []byte(newBlock.Hash)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		err := b.Put(newBlockHash, newBlock.Serialize())
		if err != nil {
			fmt.Printf("Error while writing to DB: %s", err)
			return err
		}
		err = b.Put([]byte("l"), newBlockHash)
		bc.Tip = newBlockHash

		return nil
	})
	if err != nil {
		fmt.Printf("Error while writing to DB: %s", err)
	}
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{string(bc.Tip), bc.DB}
}
