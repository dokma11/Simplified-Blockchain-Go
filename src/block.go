package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	Data         string
	Timestamp    string
	Hash         string
	PreviousHash string
	Nonce        int
}

func NewBlock(previousBlockHash string, data string) *Block {
	block := &Block{
		Timestamp:    time.Now().String(),
		Data:         data,
		PreviousHash: previousBlockHash,
		Nonce:        0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = string(hash[:])
	block.Nonce = nonce

	return block
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
