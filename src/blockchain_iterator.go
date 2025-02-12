package main

import (
	"fmt"
	"github.com/boltdb/bolt"
)

type BlockchainIterator struct {
	CurrentHash string
	DB          *bolt.DB
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get([]byte(i.CurrentHash))
		block = DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		fmt.Printf("Error while increasing the iterator value: %v", err)
	}

	i.CurrentHash = block.PreviousHash

	return block
}
