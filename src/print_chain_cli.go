package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"strconv"
)

func (cli *CLI) printChain(nodeID string) {
	bc := NewBlockchain(nodeID)
	defer func(DB *bolt.DB) {
		err := DB.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.DB)

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PreviousHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PreviousHash) == 0 {
			break
		}
	}
}
