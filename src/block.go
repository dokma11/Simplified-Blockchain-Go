package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"
)

type Block struct {
	Transactions []*Transaction
	Timestamp    string
	Hash         string
	PreviousHash string
	Nonce        int
	Height       int
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{
		transactions,
		strconv.FormatInt(time.Now().Unix(), 10),
		string(prevBlockHash),
		"",
		0,
		height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = string(hash[:])
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		fmt.Printf("Serialization error: %s", err)
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Printf("Deserialization error: %s", err)
	}

	return &block
}

func (b *Block) HashTransactions() string {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := NewMerkleTree(transactions)

	return string(mTree.RootNode.Data)
}
