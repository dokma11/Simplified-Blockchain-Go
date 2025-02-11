package main

import (
	"crypto/sha256"
	"math/big"
	"time"
)

type Block struct {
	Data         string
	Timestamp    string
	Hash         string
	PreviousHash string
	Nonce        int
}

func NewBlock(previousBlock *Block, data string) *Block {
	block := &Block{
		Timestamp:    time.Now().String(),
		Data:         data,
		PreviousHash: previousBlock.Hash,
		Nonce:        0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = string(hash[:])
	block.Nonce = nonce

	return block
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.Block.Nonce)
	hash := sha256.Sum256([]byte(data))
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.Target) == -1

	return isValid
}
