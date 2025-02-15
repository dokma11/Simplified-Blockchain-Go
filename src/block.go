package main

import (
	"bytes"
	"crypto/sha256"
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
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		transactions,
		strconv.FormatInt(time.Now().Unix(), 10),
		string(prevBlockHash),
		"",
		0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = string(hash[:])
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
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

// DeserializeBlock TODO: videti samo da li da bude metoda ili obicna fja
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
	var txHashes []string
	var txHashesByte [][]byte
	var txHash [32]byte

	for _, hash := range txHashes {
		txHashesByte = append(txHashesByte, []byte(hash))
	}

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashesByte, []byte{}))

	return string(txHash[:])
}
